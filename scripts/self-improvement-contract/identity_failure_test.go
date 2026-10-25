package main

import (
	"strings"
	"testing"
)

func TestContractIdentityAndRelationsRemainAuthoritative(t *testing.T) {
	path, source := contractIdentitySource(t)
	for _, change := range []struct{ name, old, replacement string }{
		{"activity-identity", "meta://activity/execute", "meta://activity/unbound-execute"},
		{"entity-identity", "gooo://meta/source-tree", "gooo://meta/other-source-tree"},
		{"ordered-inputs", "Plan(SourceMetrics, ProofChoice)", "Plan(ProofChoice, SourceMetrics)"},
		{"missing-input", "Plan(SourceMetrics, ProofChoice)", "Plan(SourceMetrics)"},
		{"wrong-result", "Improve(Evidence) -> SourceTree", "Improve(Evidence) -> Evidence"},
		{"observation-consumed", "Improve(Evidence)", "Improve(Evidence, ReadOnlyImprovementInput)"},
		{"extra-observer", "Verify(Receipt) -> Evidence", "Verify(Receipt) -> ReadOnlyImprovementInput"},
	} {
		t.Run(change.name, func(t *testing.T) {
			broken := strings.Replace(source, change.old, change.replacement, 1)
			broken = renameContractLabels(t, broken, map[string]string{
				"Execute": "실행", "SourceTree": "소스", "ReadOnlyImprovementInput": "읽기전용",
			})
			report := buildReport(path, []byte(broken), strings.Repeat("b", 40))
			if report.Status != "FAIL" || report.PromotionAuthorized {
				t.Fatal("renaming hid a changed identity or relation", report)
			}
		})
	}
}

func TestContractRejectsDuplicateAuthoredIdentities(t *testing.T) {
	path, source := contractIdentitySource(t)
	for _, extra := range []string{
		`entity OtherTree id "gooo://meta/source-tree"`,
		`activity OtherMeasure(SourceTree) -> SourceMetrics id "meta://activity/measure"`,
	} {
		if _, err := compileContract(path, []byte(source+"\n"+extra+"\n")); err == nil {
			t.Fatal("duplicate authored identity was silently coalesced")
		}
	}
}
