package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type indexedTraceWant struct {
	Source                           string
	Case, Use, Low, High             int
	Roots, Links, Reachable, Parents int
	Unresolved                       int
	Status                           string
	Complete                         bool
}

func (w indexedTraceWant) value(t *testing.T) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"summary": map[string]any{"source_sha256": w.Source, "case_index": w.Case, "selected_use": w.Use,
			"status": w.Status, "roots": "", "root_count": w.Roots, "reachable_writes": w.Reachable,
			"parent_links": w.Parents, "field_read_edges": 0, "unresolved_links": w.Unresolved,
			"shared_rhs_writes": 0, "scalar_rhs_writes": 0, "recording_complete": w.Complete,
			"local_read_edges": 0, "carrier_writes": 0, "literal_writes": 0, "read_edges": 0,
			"collection_rhs_writes": 0},
		"indexed_low": w.Low, "indexed_high": w.High, "indexed_roots": w.Roots, "element_links": w.Links,
	})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestRecordIterationIndexedGoooLineage(t *testing.T) {
	source, cases := recordIterationExample(t)
	r := recordIterationRuntime(t, source, cases)
	columns, report, err := SelectRecordedIndexedBridgeLineagePage(r, NativeOriginQuery{Kind: "field-uses", Limit: 128})
	if err != nil || len(columns) != 8 || report.InputShape != "gooo/field-local-lineage-columns/v2" ||
		report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 {
		t.Fatal("indexed transport lost actual occurrences or executed producer", err, report)
	}
	wants := []indexedTraceWant{}
	caseIDs := []int{1, 1, 2, 2, 2, 2, 3, 3}
	useIDs := []int{0, 1, 0, 1, 2, 3, 0, 1}
	low := []int{1, 2, 1, 2, 4, 8, 1, 2}
	indices := []int{0, 0, 0, 0, 1, 1, 0, 0}
	for i, column := range columns {
		ref := column.Elements.Refs[column.History.RootAssignment]
		if column.History.CaseIndex != caseIDs[i] || column.History.SelectedUse != useIDs[i] {
			t.Fatal("selected use identity changed", i, column.History)
		}
		found := false
		for j, read := range column.Elements.ReadIDs {
			if ref == read {
				found = column.Elements.Indices[j] == indices[i] && column.Elements.FieldIDs[j] == "urn:record-items:items" &&
					column.Elements.RecordIDs[j] == "urn:record-items:line" && column.Elements.OwnerIDs[j] == "urn:record-items:basket"
			}
		}
		if !found {
			t.Fatal("same-valued inputs lost original field, nominal type or index", i, column.Elements)
		}
		wants = append(wants, indexedTraceWant{Source: r.SelectedSourceSHA256, Case: caseIDs[i], Use: useIDs[i], Low: low[i],
			Roots: 1, Links: 1, Reachable: 1, Status: "SCOPED_INDEXED_FIELD_LOCAL_ORIGINS", Complete: true})
	}
	checkIndexedGoooPrograms(t, columns, wants)
	t.Log("8 actual field uses retain independent collection indices0/1 and nominal IDs;Gooo chooses original write masks;interpreter/native agree;producer4 finite cases remain separate;query executions0/models0")
}

func TestRecordIterationNestedIndexedGoooLineage(t *testing.T) {
	r := recordIterationNestedRuntime(t)
	columns, report, err := SelectRecordedIndexedBridgeLineagePage(r, NativeOriginQuery{Kind: "field-uses", Limit: 128})
	if err != nil || len(columns) != 12 || report.ProducerExecutions != 0 {
		t.Fatal("nested actual reads", err, report)
	}
	wants := []indexedTraceWant{}
	roots := []int{0, 1, 0, 5, 0, 7, 8, 9, 8, 13, 8, 15}
	positions := []int{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 1, 1}
	fields := []string{"items", "items", "items", "others", "items", "others", "items", "items", "items", "others", "items", "others"}
	for i, column := range columns {
		ref := column.Elements.Refs[roots[i]]
		found := false
		for j, read := range column.Elements.ReadIDs {
			if read == ref {
				found = column.Elements.Indices[j] == positions[i] && column.Elements.FieldIDs[j] == "urn:record-items:"+fields[i]
			}
		}
		if !found {
			t.Fatal("nested input collection identity or original index changed", i, column.Elements)
		}
		parents := 0
		if i == 2 || i == 4 || i == 8 || i == 10 {
			parents = 1
		}
		wants = append(wants, indexedTraceWant{Source: r.SelectedSourceSHA256, Use: i, Low: 1 << roots[i],
			Roots: 1, Links: 1, Reachable: 1 + parents, Parents: parents,
			Status: "SCOPED_INDEXED_FIELD_LOCAL_ORIGINS", Complete: true})
	}
	checkIndexedGoooPrograms(t, columns, wants)
	t.Log("12 actual nested field reads keep two same-valued collection identities;four copied reads follow original parent writes;Gooo/native agree with independently named roots and indices")
}

func checkIndexedGoooPrograms(t *testing.T, columns []IndexedBridgeLineageColumns, wants []indexedTraceWant) {
	t.Helper()
	source, err := os.ReadFile("../../examples/local-field-bridge/trace.gooo")
	if err != nil {
		t.Fatal(err)
	}
	program, err := bodycodegen.NewFixedProgram(context.Background(), "trace.gooo", source, "TraceIndexed")
	if err != nil {
		t.Fatal(err)
	}
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for i, column := range columns {
		raw, err := json.Marshal(column)
		if err != nil {
			t.Fatal(err)
		}
		expected := wants[i].value(t)
		tuple, _ := json.Marshal([]json.RawMessage{raw})
		actual, err := program.Evaluate(context.Background(), tuple)
		if err != nil {
			t.Fatal("indexed interpreter", i, err)
		}
		decode := func(raw []byte) any {
			var value any
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if err := decoder.Decode(&value); err != nil {
				t.Fatal(err)
			}
			return value
		}
		if !reflect.DeepEqual(decode(actual.Value), decode(expected)) {
			t.Fatalf("indexed interpreter case%d: got%s want%s", i, actual.Value, expected)
		}
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"TraceIndexed": raw},
			Expected: map[string]json.RawMessage{"TraceIndexed": expected}})
	}
	plan, err := GenerateCompositionWithOptions(context.Background(), "trace.gooo", source, suite, CompositionOptions{EntryActivity: "TraceIndexed"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), "trace.gooo", source, plan, suite, "")
	if err != nil || r.FinitePassed != len(columns) || !r.RuntimeReplayed || r.ModelCalls != 0 {
		t.Fatal("indexed native consumer", err, r.FinitePassed, r.FiniteTotal)
	}
}
