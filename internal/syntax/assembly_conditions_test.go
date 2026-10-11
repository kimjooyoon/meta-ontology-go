package syntax

import (
	"os"
	"strings"
	"testing"
)

func TestConditionCaseSyntaxRoundTripAndClone(t *testing.T) {
	raw, err := os.ReadFile("../../examples/body-codegen/source-condition-cases.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	source := strings.TrimSpace(strings.Split(string(raw), "activity Main")[0])
	testAssemblyParseFormatCloneAndSpans(t, source)
	f, diagnostics := Parse(source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	original := f.Declarations[1].(*ActivityDecl).Assembly.Spec.ConditionCases
	clone := f.Clone().Declarations[1].(*ActivityDecl).Assembly.Spec.ConditionCases
	clone[0].Input = 1
	clone[0].Expected = false
	if original[0].Input != -9007199254740995 || !original[0].Expected {
		t.Fatal("condition cases shared clone storage or rounded input")
	}
	for _, change := range []struct{ from, to string }{
		{`-> "true"`, `-> "TRUE"`},
		{`-> "true"`, `-> true`},
		{`-> "true"`, `-> "1"`},
		{`input "0"`, `input "9223372036854775808"`},
		{`condition_case "comparison"`, `condition_case "missing"`},
		{`input "0"`, `input "-9007199254740995"`},
	} {
		if _, diagnostics := Parse(strings.Replace(source, change.from, change.to, 1)); !diagnostics.HasErrors() {
			t.Fatalf("accepted invalid condition syntax: %+v", change)
		}
	}
}
