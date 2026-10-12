package syntax

import (
	"fmt"
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

func TestObserveIntegerLimitsAreCheckedBeforeNarrowing(t *testing.T) {
	for _, field := range []struct {
		name  string
		limit int64
	}{{"max_candidates", 64}, {"max_rounds", 8}} {
		for _, value := range []int64{-9223372036854775808, -1, 0, 1, field.limit, field.limit + 1, 4294967297, 4294967296 + field.limit, 9223372036854775807} {
			candidates, rounds := int64(2), int64(1)
			if field.name == "max_candidates" {
				candidates = value
			} else {
				rounds = value
			}
			observation := fmt.Sprintf(`observe { input "3" max_candidates "%d" max_rounds "%d" } attempts "2"`, candidates, rounds)
			source := strings.Replace(assemblySource, `attempts "2"`, observation, 1)
			file, diagnostics := Parse(source)
			valid := value >= 1 && value <= field.limit
			if diagnostics.HasErrors() == valid {
				t.Fatalf("%s=%d: valid=%t diagnostics=%v", field.name, value, valid, diagnostics)
			}
			if valid {
				o := file.Declarations[1].(*ActivityDecl).Assembly.Spec.Observation
				if o.MaxCandidates != int(candidates) || o.MaxRounds != int(rounds) {
					t.Fatal(o)
				}
			}
		}
	}
}
