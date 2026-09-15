package guardedcapability

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedpromotion"

func indicators(source Source, summary Summary, coordinates []guardedpromotion.Coordinate) []guardedpromotion.Indicator {
	passed := summary.Satisfied == summary.Total && summary.Total == 8
	return []guardedpromotion.Indicator{
		metric("capability-readiness-bps", "OUTCOME", "COHERENCE", boolInt(passed)*10000, 10000),
		metric("foundation-receipts", "DRIVER", "FOUNDATION", summary.FoundationReceipts, 2),
		metric("implementation-tree-equivalence-bps", "DRIVER", "COHERENCE", summary.EquivalentTrees*5000, 10000),
		metric("foundation-ancestor-bps", "DRIVER", "COHERENCE", coordinateBPS(coordinates, "foundation-ancestor"), 10000),
		metric("unresolved-evidence", "GUARDRAIL", "FOUNDATION", summary.Unresolved, 0),
		metric("implementation-tree-drift", "GUARDRAIL", "REGRESSION", 2-summary.EquivalentTrees, 0),
		metric("observer-writes", "GUARDRAIL", "REGRESSION", source.RepositoryWrites, 0),
		metric("mutation-authority", "GUARDRAIL", "REGRESSION", boolInt(source.MutationAuthorized), 0),
	}
}

func metric(id, class, proof string, value, target int) guardedpromotion.Indicator {
	return guardedpromotion.Indicator{
		MetricID: "gooo.metric.language.guarded-capability-" + id + ".v1",
		Class:    class, ProofChoice: proof,
		Producer: "internal/meta/languagereadiness/guardedcapability",
		Consumer: "language-readiness", MetaOperation: "bind-guarded-capability-foundation",
		Value: value, Target: target, Satisfied: value == target,
	}
}

func coordinateBPS(coordinates []guardedpromotion.Coordinate, id string) int {
	for _, item := range coordinates {
		if item.ID == id && item.Status == "SATISFIED" {
			return 10000
		}
	}
	return 0
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
