package bodyexecution

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func subjectRuntime23963(t *testing.T, name string) CompositionRuntime {
	t.Helper()
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	f, err := os.Open(filepath.Join(root, "internal/bodyexecution/testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var r CompositionRuntime
	if err := json.NewDecoder(z).Decode(&r); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestOriginSubjectSelection23963(t *testing.T) {
	r := subjectRuntime23963(t, "choice22133-runtime.json.gz")
	q := NativeOriginQuery{Kind: "field-input-locations", FieldID: "urn:assembledlength22133:total", Offset: 2, Limit: 5}
	rows, report, err := SelectRecordedNativeOrigins(r, q)
	if err != nil || len(rows) != 5 || report.SourceRows == nil || *report.SourceRows != 27 || report.TotalRows != 9 || !report.Truncated || report.NextOffset == nil || *report.NextOffset != 7 || report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 {
		t.Fatal("filtered offsets,source population and producer scope", err, report)
	}
	for i, row := range rows {
		var f CompositionFieldFileFact
		if err := json.Unmarshal(row, &f); err != nil {
			t.Fatal(err)
		}
		if f.FieldID != q.FieldID || f.CaseIndex != (i+2)/3 || f.AssignmentIndex != 0 || f.EventIndex != 0 {
			t.Fatal("original event identity", i, f)
		}
	}
	q.RootActivityID = "urn:assembledlength22133:summarize"
	q.ActivityID = q.RootActivityID
	if _, p, err := SelectRecordedNativeOrigins(r, q); err != nil || p.TotalRows != 9 {
		t.Fatal("conjoined identities", err, p)
	}
	q.RootActivityID = "urn:absent:root"
	q.Offset = 0
	rows, report, err = SelectRecordedNativeOriginPage(r, q)
	if err != nil || len(rows) != 0 || report.TotalRows != 0 || report.SourceRows == nil || *report.SourceRows != 27 || report.Truncated || report.NextOffset != nil {
		t.Fatal("empty selected subject", err, report)
	}
	if _, _, err := SelectRecordedNativeOrigins(r, q); err == nil {
		t.Fatal("nonempty row contract changed")
	}
	for _, kind := range []string{"reads", "elements", "local-writes", "local-fill-fragments"} {
		q.Kind = kind
		if err := ValidateNativeOriginQuery(q); err == nil {
			t.Fatal("subject filter silently ignored", kind)
		}
	}
	plain := NativeOriginQuery{Kind: "field-writes", Limit: 1}
	raw, err := json.Marshal(plain)
	if err != nil || string(raw) != `{"kind":"field-writes","offset":0,"limit":1}` {
		t.Fatal("legacy query encoding", err, string(raw))
	}
	_, report, err = SelectRecordedNativeOrigins(r, plain)
	if err != nil || report.SourceRows != nil || report.TotalRows != 9 {
		t.Fatal("legacy unfiltered scope", err, report)
	}
	// Display labels can change while the recorded semantic IDs remain fixed.
	for i := range r.FieldBindings {
		r.FieldBindings[i].Record.Name = "A different display label23963"
		for j := range r.FieldBindings[i].Record.Fields {
			r.FieldBindings[i].Record.Fields[j].Name = "renamed label23963"
		}
	}
	r.FieldObservation.MetadataSHA256 = compositionDigest([]any{r.FieldProjectionSHA256, r.FieldBindings, r.FieldSites, r.FieldUseSites})
	r.BridgeObservation.MetadataSHA256 = compositionDigest([]any{r.FieldProjectionSHA256, r.FieldBindings, r.FieldSites, r.FieldUseSites, r.LocalSites, r.LocalUseSites})
	q = NativeOriginQuery{Kind: "field-writes", FieldID: "urn:assembledlength22133:total", Limit: 128}
	rows, report, err = SelectRecordedNativeOrigins(r, q)
	if err != nil || len(rows) != 3 || report.SourceRows == nil || *report.SourceRows != 9 {
		t.Fatal("stable ID versus display labels", err, report)
	}
	t.Log("locations27 retained/9 selected/5 paged;original case and event numbering;three exact ID predicates;empty selected table;four incompatible kinds rejected;legacy query bytes retained;display labels independent;producer/model0")
}

func TestOriginSubjectLineage23963(t *testing.T) {
	r := subjectRuntime23963(t, "return21917-runtime.json.gz")
	q := NativeOriginQuery{Kind: "field-input-locations", FieldID: "urn:directreturn21917:tag", RootActivityID: "urn:directreturn21917:pick", ActivityID: "urn:directreturn21917:make", Offset: 1, Limit: 1}
	rows, p, err := SelectRecordedNativeOrigins(r, q)
	if err != nil || len(rows) != 1 || p.SourceRows == nil || *p.SourceRows != 20 || p.TotalRows != 2 || p.Truncated {
		t.Fatal("helper subject population", err, p)
	}
	var fact CompositionFieldFileFact
	if err := json.Unmarshal(rows[0], &fact); err != nil || fact.CaseIndex != 0 || fact.EventIndex != 3 || fact.AssignmentIndex != 3 || fact.Category != "value" || fact.Expression != `"helper"` {
		t.Fatal("helper original identity", err, fact)
	}
	q = NativeOriginQuery{Kind: "field-uses", FieldID: "urn:directreturn21917:tag", RootActivityID: "urn:directreturn21917:pick", ActivityID: "urn:directreturn21917:pick", Limit: 1}
	columns, p, err := SelectRecordedFieldLineage(r, q)
	if err != nil || p.SourceRows == nil || *p.SourceRows != 2 || p.TotalRows != 1 || columns.CaseIndex != 2 || columns.SelectedUse != 1 || columns.RootAssignment != 1 || columns.WriteCount != 2 || columns.UseCount != 2 || len(columns.Fields) != 2 || columns.Fields[0] == q.FieldID || columns.Fields[1] != q.FieldID {
		t.Fatal("unsliced field lineage", err, p, columns)
	}
	bridge, p, err := SelectRecordedBridgeLineagePage(r, q)
	if err != nil || len(bridge) != 1 || p.SourceRows == nil || *p.SourceRows != 2 || p.TotalRows != 1 || bridge[0].SelectedUse != 1 || bridge[0].FieldUseCount != 2 || bridge[0].WriteCount != 2 || len(bridge[0].Fields) != 2 {
		t.Fatal("unsliced field/local bridge", err, p)
	}
	t.Log("helper origins20 retained/2 selected;tag use2 retained/1 selected;original use1,assignment1,case2;both field and bridge columns retain complete two-field cases;producer/model0")
}
