package languagesyntax

import "testing"

func TestMetaSourceHistoryRequiresAnAppendedExecutableCase(t *testing.T) {
	old := CaseDefinition{ID: "old", Path: "old.gooo"}
	promoted := CaseDefinition{ID: "promoted", Path: "source-B", Kind: KindValid,
		ExpectedDecision: DecisionPass, Scope: ScopeLanguageCapability, MetaOperation: "replay-language-syntax"}
	prior := Registry{Cases: []CaseDefinition{old}, MetaSources: []string{"source-A", "source-B", "source-C"}}
	for _, name := range []string{"promotion", "deletion", "wrong-source", "wrong-scope", "failed-case",
		"wrong-operation", "duplicate-cohort", "reordered", "rewritten-old-case"} {
		t.Run(name, func(t *testing.T) {
			current := Registry{Cases: []CaseDefinition{old, promoted}, MetaSources: []string{"source-A", "source-C"}}
			switch name {
			case "deletion":
				current.Cases = current.Cases[:1]
			case "wrong-source":
				current.Cases[1].Path = "source-D"
			case "wrong-scope":
				current.Cases[1].Scope = ScopeGovernanceObservation
			case "failed-case":
				current.Cases[1].ExpectedDecision = "FAIL_CLOSED"
			case "wrong-operation":
				current.Cases[1].MetaOperation = "unobserved"
			case "duplicate-cohort":
				current.MetaSources = append([]string{}, prior.MetaSources...)
			case "reordered":
				current.MetaSources = []string{"source-C", "source-A"}
			case "rewritten-old-case":
				current.Cases[0].ID = "changed"
			}
			if got := metaSourceHistoryRetainedByCases67143(prior, current); got != (name == "promotion") {
				t.Fatalf("history transition %s: %v", name, got)
			}
		})
	}
}
