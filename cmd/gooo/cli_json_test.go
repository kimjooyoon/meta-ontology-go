package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRunCheckJSONDiagnosticsAreStableAndMachineReadable(t *testing.T) {
	args := []string{"--json", "broken.gooo"}
	var firstOut, firstErr bytes.Buffer
	firstCode := runCheck(args, fixtureReader{source: "package billing\nentity Broken id \"x\" @"}, SyntaxSourceParser{}, &firstOut, &firstErr)
	var secondOut, secondErr bytes.Buffer
	secondCode := runCheck(args, fixtureReader{source: "package billing\nentity Broken id \"x\" @"}, SyntaxSourceParser{}, &secondOut, &secondErr)
	if firstCode != exitFailure || secondCode != exitFailure || !bytes.Equal(firstOut.Bytes(), secondOut.Bytes()) || firstErr.Len() != 0 || secondErr.Len() != 0 {
		t.Fatalf("JSON diagnostics were not deterministic: first=(%d,%q,%q), second=(%d,%q,%q)", firstCode, firstOut.String(), firstErr.String(), secondCode, secondOut.String(), secondErr.String())
	}
	var report jsonReport
	if err := json.Unmarshal(firstOut.Bytes(), &report); err != nil {
		t.Fatalf("JSON diagnostic report did not decode: %v", err)
	}
	if report.SchemaVersion != diagnosticSchemaVersion || report.Command != "check" || report.Status != "error" || len(report.Diagnostics) != 2 {
		t.Fatalf("unexpected diagnostic report: %#v", report)
	}
	if report.Diagnostics[0].Code != "parse.expected-namespace" || report.Diagnostics[1].Code != "lex.unexpected-character" {
		t.Fatalf("diagnostic order/codes = %#v", report.Diagnostics)
	}
}

func TestRunCheckSemanticModeReportsLoweringDiagnosticsAsJSON(t *testing.T) {
	source := `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Missing) -> Order
`
	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"--semantic", "--json", "fixture.gooo"}, fixtureReader{source: source}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitFailure || stderr.Len() != 0 {
		t.Fatalf("semantic JSON check = %d, stderr=%q", code, stderr.String())
	}
	var report jsonReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "error" || len(report.Diagnostics) != 1 || report.Diagnostics[0].Code != "semantic.lowering" {
		t.Fatalf("unexpected semantic report: %#v", report)
	}
}
