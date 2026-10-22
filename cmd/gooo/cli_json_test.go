package main

import (
	"bytes"
	"encoding/json"
	"strings"
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
	span := report.Diagnostics[0].Span
	if span.File != "fixture.gooo" || span.Start.Line != 4 || span.Start.Column != 19 || span.End.Line != 4 || span.End.Column != 26 {
		t.Fatalf("unresolved input span = %#v, want authored Missing reference", span)
	}
}

func TestRunCheckSemanticOutputPointsToResultReference(t *testing.T) {
	source := "package billing\nnamespace billing\nentity Order id \"billing://order\"\nactivity Pay(Order) -> Missing\n"
	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"--semantic", "--json", "fixture.gooo"}, fixtureReader{source: source}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitFailure || stderr.Len() != 0 {
		t.Fatalf("semantic output check = %d, stderr=%q", code, stderr.String())
	}
	var report jsonReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v", report.Diagnostics)
	}
	span := report.Diagnostics[0].Span
	if span.Start.Line != 4 || span.Start.Column != 24 || span.End.Line != 4 || span.End.Column != 31 {
		t.Fatalf("unresolved output span = %#v, want authored Missing reference", span)
	}
}

func TestSemanticReferenceSpanIsSharedByTextCheckAndJSONQuery(t *testing.T) {
	source := "package billing\nnamespace billing\nentity Order id \"billing://order\"\nactivity Pay(Missing) -> Order\n"
	reader := fixtureReader{source: source}
	var checkOut, checkErr bytes.Buffer
	if code := runCheck([]string{"--semantic", "fixture.gooo"}, reader, SyntaxSourceParser{}, &checkOut, &checkErr); code != exitFailure {
		t.Fatalf("text check exit = %d", code)
	}
	if !strings.Contains(checkErr.String(), "fixture.gooo:4:14-4:21:") {
		t.Fatalf("text check span = %q", checkErr.String())
	}
	var queryOut, queryErr bytes.Buffer
	if code := runQuery([]string{"--json", "fixture.gooo"}, reader, SyntaxSourceParser{}, &queryOut, &queryErr); code != exitFailure || queryErr.Len() != 0 {
		t.Fatalf("JSON query exit = %d, stderr = %q", code, queryErr.String())
	}
	var report jsonReport
	if err := json.Unmarshal(queryOut.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Diagnostics) != 1 || report.Diagnostics[0].Span.Start.Line != 4 || report.Diagnostics[0].Span.Start.Column != 14 {
		t.Fatalf("query diagnostics = %#v", report.Diagnostics)
	}
}
