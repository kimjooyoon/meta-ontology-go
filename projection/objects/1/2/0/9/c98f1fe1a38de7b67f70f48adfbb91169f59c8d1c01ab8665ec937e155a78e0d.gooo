package linecaps

import (
	"strings"
	"testing"
)

func TestRefactorCandidateFindingForSimpleReturns(t *testing.T) {
	source := `
package p

func ReturnDirect(v int) int {
	return v
}

func ReturnWrapped(v int) int {
	return len(v)
}

func ReturnVariable(v int) int {
	value := v
	return value
}

func NotRefactorable(v int) int {
	if v > 0 {
		return v
	}
	return 0
}
`
	findings, err := AnalyzeSource("fixture.go", []byte(source), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	refactorRules := filterRules(findings, []Rule{RuleRefactorReturn, RuleRefactorAssign})
	if len(refactorRules) != 3 {
		t.Fatalf("unexpected refactor findings: %#v", findings)
	}
	for _, finding := range refactorRules {
		if finding.Rule == RuleRefactorReturn && !strings.Contains(finding.Detail, "single return") {
			t.Fatalf("unexpected refactor-return detail: %#v", finding)
		}
	}
	if !hasFinding(findings, RuleRefactorAssign) {
		t.Fatalf("expected assignment refactor finding: %#v", findings)
	}
	if !hasFinding(findings, RuleRefactorReturn) {
		t.Fatalf("expected return refactor findings: %#v", findings)
	}
}
func TestRefactorCandidateFindingSkipsLargeFunctions(t *testing.T) {
	source := `
package p

func Long(v int) int {
	for i := 0; i < 3; i++ {
		v += i
	}
	return v
}
`
	findings, err := AnalyzeSource("fixture.go", []byte(source), DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if hasFinding(findings, RuleRefactorReturn) || hasFinding(findings, RuleRefactorAssign) {
		t.Fatalf("non-trivial function should not be refactor candidate: %#v", findings)
	}
}
func sourceWithFunctionLines(lines int) string {
	return "package p\n\nfunc F() {\n" + strings.Repeat("\t_ = 1\n", lines-4) + "}\n"
}
