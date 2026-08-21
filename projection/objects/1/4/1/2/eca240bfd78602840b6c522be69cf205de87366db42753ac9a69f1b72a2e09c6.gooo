package main

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"testing"
)

func TestInspectAndAnalyzeValidateLowererResults(t *testing.T) {
	invalidLowerer := func(*syntax.File) (semantic.IR, error) { return semantic.IR{}, nil }
	var inspectOut, inspectErr bytes.Buffer
	inspectCode := runInspectWithLowerer([]string{"fixture.gooo"}, fixtureReader{source: sourceOrderA}, SyntaxSourceParser{}, &inspectOut, &inspectErr, invalidLowerer)
	if inspectCode != exitFailure || inspectOut.Len() != 0 || !bytes.Contains(inspectErr.Bytes(), []byte("semantic.invalid")) {
		t.Fatalf("inspect accepted invalid IR = code %d, stdout=%q, stderr=%q", inspectCode, inspectOut.String(), inspectErr.String())
	}

	var analyzeOut, analyzeErr bytes.Buffer
	analyzeCode := runAnalyzeWithLowerer([]string{"fixture.gooo"}, fixtureReader{source: sourceOrderA}, SyntaxSourceParser{}, &analyzeOut, &analyzeErr, invalidLowerer)
	if analyzeCode != exitFailure || analyzeErr.Len() != 0 {
		t.Fatalf("analyze accepted invalid IR = code %d, stderr=%q", analyzeCode, analyzeErr.String())
	}
	plan := decodeFixPlan(t, analyzeOut.Bytes())
	if plan.Status != fixPlanSemanticInvalid || len(plan.Diagnostics) != 1 || plan.Diagnostics[0].Code != "semantic.lowering" {
		t.Fatalf("invalid IR analyze plan = %#v", plan)
	}
}
func checkFixture(t *testing.T, source string) (stdout, stderr string, code int) {
	t.Helper()
	var out, err bytes.Buffer
	code = runCheck([]string{"--semantic", "billing.gooo"}, fixtureReader{source: source}, SyntaxSourceParser{}, &out, &err)
	return out.String(), err.String(), code
}
