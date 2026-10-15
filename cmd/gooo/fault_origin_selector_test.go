package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func savedFaultOriginRuntime(t *testing.T, member string) bodyexecution.CompositionRuntime {
	t.Helper()
	raw, err := os.ReadFile("../../docs/observations/local-division-origins/first-observations.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	var saved struct {
		Runtime bodyexecution.CompositionRuntime
	}
	if err := json.Unmarshal(readNamedInputArchive(t, raw)[member], &saved); err != nil {
		t.Fatal(err)
	}
	return saved.Runtime
}

func TestFaultOriginSelectionPreservesRecordedOperations(t *testing.T) {
	runtime := savedFaultOriginRuntime(t, "repair-outputs/local-division-faults.json")
	query := bodyexecution.NativeOriginQuery{Kind: "faults", Limit: 1, ActivityID: "urn:localdivision:divide"}
	first, report, err := bodyexecution.SelectRecordedNativeOrigins(runtime, query)
	if err != nil || report == nil || len(first) != 1 || report.TotalRows != 2 || report.NextOffset == nil || *report.NextOffset != 1 || report.SourceRows == nil || *report.SourceRows != 2 {
		t.Fatal("first fault page", err, report)
	}
	query.Offset = 1
	second, last, err := bodyexecution.SelectRecordedNativeOrigins(runtime, query)
	if err != nil || len(second) != 1 || last.Truncated || last.NextOffset != nil || last.ProducerExecutions != 0 || last.ProducerModelCalls != 0 {
		t.Fatal("second fault page", err, last)
	}
	for i, raw := range append(first, second...) {
		var fact bodyexecution.CompositionFaultFact
		if err := json.Unmarshal(raw, &fact); err != nil {
			t.Fatal(err)
		}
		if fact.CaseIndex != i+2 || fact.Kind != "ZERO_DIVISOR" || fact.Right != 0 || fact.ActivityID != query.ActivityID || fact.RootActivityID != "urn:localdivision:main" {
			t.Fatal("fault identity", fact)
		}
		if i == 0 && (fact.Variable != "amount" || fact.SourceMapping != "MAPPED" || fact.SourceRelation != "selected_local_rhs") || i == 1 && (fact.Assignment != "" || fact.SourceMapping != "UNMAPPED" || fact.SourceStart != -1) {
			t.Fatal("compound and inner source mixed", fact)
		}
	}
}

func TestFaultOriginSelectionRetainsMixedAndEmptyTables(t *testing.T) {
	runtime := savedFaultOriginRuntime(t, "mixed-outputs/mixed-division-faults.json")
	query := bodyexecution.NativeOriginQuery{Kind: "faults", Limit: 8}
	rows, report, err := bodyexecution.SelectRecordedNativeOrigins(runtime, query)
	if err != nil || len(rows) != 2 || report.TotalRows != 2 {
		t.Fatal("mixed rows", err)
	}
	for i, row := range rows {
		var fact bodyexecution.CompositionFaultFact
		if err := json.Unmarshal(row, &fact); err != nil || fact.Assignment != "/=" || (i == 0) != (fact.Variable == "") {
			t.Fatal("field and local identity", i, err, fact)
		}
	}
	before, _ := json.Marshal(runtime)
	query.RootActivityID = "urn:not-recorded"
	rows, report, err = bodyexecution.SelectRecordedNativeOriginPage(runtime, query)
	if err != nil || len(rows) != 0 || report.TotalRows != 0 || report.SourceRows == nil || *report.SourceRows != 2 {
		t.Fatal("empty selected table", err, report)
	}
	after, _ := json.Marshal(runtime)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("selection changed producer history")
	}
}
