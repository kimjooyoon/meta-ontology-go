package syntax

import (
	"strings"
	"testing"
)

func TestSourceObserveSyntaxRoundTripAndBoundaries(t *testing.T) {
	source := strings.Replace(assemblySource, `    attempts "2"`, `    observe {
        input "-9007199254740995"
        input "0"
        oracle "기준"
        max_candidates "2"
        max_rounds "1"
        reuse_outputs "true"
        resolve_unique "true"
    }
    attempts "2"`, 1)
	testAssemblyParseFormatCloneAndSpans(t, source)
	file, diagnostics := Parse(source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	o := file.Declarations[1].(*ActivityDecl).Assembly.Spec.Observation
	clone := file.Clone().Declarations[1].(*ActivityDecl).Assembly.Spec.Observation
	clone.Inputs[0], clone.OracleActivity = 7, "changed"
	if o.Inputs[0] != -9007199254740995 || o.OracleActivity != "기준" {
		t.Fatal("source observation shared storage or rounded an integer")
	}
	for _, change := range []struct{ from, to string }{
		{`max_candidates "2"`, `max_candidates "65"`},
		{`max_rounds "1"`, `max_rounds "0"`},
		{`max_rounds "1"`, `max_rounds "9223372036854775807"`},
		{`max_rounds "1"`, ``},
		{`max_rounds "1"`, `max_rounds "1" max_rounds "2"`},
		{`input "0"`, `input "-9007199254740995"`},
		{`input "0"`, `input "9223372036854775808"`},
		{`oracle "기준"`, `oracle ""`},
		{`oracle "기준"`, `oracle "not.a.name"`},
		{`reuse_outputs "true"`, `reuse_outputs "TRUE"`},
		{`resolve_unique "true"`, `resolve_unique true`},
		{`    attempts "2"`, `    holdout_case "0" -> "7" attempts "2"`},
		{`    attempts "2"`, `    observe { input "7" max_candidates "2" max_rounds "1" } attempts "2"`},
	} {
		if _, diagnostics := Parse(strings.Replace(source, change.from, change.to, 1)); !diagnostics.HasErrors() {
			t.Fatalf("invalid observe declaration accepted: %+v", change)
		}
	}
}
