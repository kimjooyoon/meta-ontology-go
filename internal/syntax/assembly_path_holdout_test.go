package syntax

import (
	"strings"
	"testing"
)

func TestTypedPathHoldoutSyntaxRoundTripAndIsolation(t *testing.T) {
	source := strings.Replace(assemblySource, `    attempts "2"`, `    holdout_case "9007199254740993" -> "9007199254740995"
    attempts "2"`, 1)
	testAssemblyParseFormatCloneAndSpans(t, source)
	file, diagnostics := Parse(source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	spec := file.Declarations[1].(*ActivityDecl).Assembly.Spec
	if len(spec.HoldoutCases) != 1 || spec.HoldoutCases[0].Input != 9007199254740993 || spec.HoldoutCases[0].Expected != 9007199254740995 {
		t.Fatal("holdout integer meaning lost", spec.HoldoutCases)
	}
}
