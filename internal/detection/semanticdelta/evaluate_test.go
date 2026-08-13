package semanticdelta

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestEvaluateWritesReportBeforeReturningScopeError(t *testing.T) {
	input := `version semanticdelta/v1
scope id billing://activity/pay-order
delta add-fact gooo:invokes billing://activity/pay-order fraud://activity/check
`
	var output bytes.Buffer
	report, err := Evaluate(strings.NewReader(input), &output, FormatText)
	if !errors.Is(err, ErrScopeViolation) {
		t.Fatalf("Evaluate error = %v, want ErrScopeViolation", err)
	}
	if report.Passes() || !strings.Contains(output.String(), "allowed\tfalse") {
		t.Fatalf("Evaluate report/output = %#v %q", report, output.String())
	}
}

func TestEvaluateAcceptsJSONAndEmitsJSONReport(t *testing.T) {
	request := Request{
		Allowed: Scope{Prefixes: []string{"billing://"}},
		Delta:   Delta{AddedNodes: []Node{{ID: "billing://entity/order", Kind: "Entity"}}},
	}
	input, err := EncodeJSON(request)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	report, err := Evaluate(bytes.NewReader(input), &output, FormatJSON)
	if err != nil || !report.Passes() {
		t.Fatalf("Evaluate = %#v, %v", report, err)
	}
	if !strings.Contains(output.String(), `"allowed": true`) {
		t.Fatalf("JSON report = %q", output.String())
	}
}
