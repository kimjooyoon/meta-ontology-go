package main

import (
	"bytes"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"html"
	"strings"
	"testing"
)

func TestConstructionRecordHoldoutReportPreservesSeparateCounts(t *testing.T) {
	initial := &bodycodegen.RecordHoldoutReceipt{Activity: "Choose", Passed: 0, Total: 2, FieldsPassed: 1, FieldsTotal: 4, AccuracyPercent: new(float64(0))}
	final := *initial
	final.Passed = 1
	final.FieldsPassed = 3
	final.AccuracyPercent = new(float64(50))
	output := bodyConstructOutput{Construction: bodyexecution.JointConstruction{
		Initial:             bodyexecution.Composition{Preparations: []bodyexecution.CompositionStep{{Generation: bodycodegen.Result{Report: bodycodegen.Report{Activity: "Choose", RecordAssembly: &bodycodegen.RecordAssemblyReceipt{Holdout: initial}}}}}},
		FinalRecordHoldouts: []bodyexecution.JointRecordHoldout{{Evaluation: &final, CallerExposure: bodyexecution.RecordHoldoutCallerExposure{ObservedInputs: []string{"[5]"}, OtherInputs: []string{"[7]"}}}}}}
	before, _ := json.Marshal(output)
	for _, format := range []string{"text", "markdown"} {
		var out bytes.Buffer
		if err := writeBodyConstructResult(&out, output, format); err != nil {
			t.Fatal(err)
		}
		visible := html.UnescapeString(out.String())
		for _, want := range []string{"Initial record body holdout", "Final selected record holdout", "cases 0/2 (0.00%); fields 1/4 (25.00%)", "cases 1/2 (50.00%); fields 3/4 (75.00%)", "1 observed; 1 other; 0 unknown"} {
			if !strings.Contains(visible, want) {
				t.Fatal(format, want, visible)
			}
		}
	}
	after, _ := json.Marshal(output)
	if !bytes.Equal(before, after) {
		t.Fatal("report rewrote observations")
	}
	final.AccuracyPercent = nil
	final.Error = "division by zero"
	if got := constructionRecordHoldoutShare(&final); !strings.Contains(got, "unavailable") || strings.Contains(got, "50.00%") {
		t.Fatal(got)
	}
}
