package guardedcapability

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedpromotion"

func summarize(source Source, coordinates []guardedpromotion.Coordinate) Summary {
	summary := Summary{Total: len(coordinates), RepositoryWrites: source.RepositoryWrites,
		MutationAuthorized: source.MutationAuthorized}
	for _, item := range coordinates {
		switch item.Status {
		case "SATISFIED":
			summary.Satisfied++
			if item.ID == "foundation-provenance" || item.ID == "foundation-authorized-report" {
				summary.FoundationReceipts++
			}
			if item.ID == "guard-implementation-tree" || item.ID == "witness-implementation-tree" {
				summary.EquivalentTrees++
			}
		case "NOT_SATISFIED":
			summary.NotSatisfied++
		default:
			summary.Unresolved++
		}
	}
	if summary.Total > 0 {
		summary.ReadinessBPS = summary.Satisfied * 10000 / summary.Total
	}
	return summary
}

func decide(summary Summary) (string, string, string) {
	if summary.Unresolved > 0 {
		return DecisionFailClosed, ReasonUnknown, ResolutionLower
	}
	if summary.Total != 8 || summary.Satisfied != summary.Total || summary.NotSatisfied != 0 {
		return DecisionFailClosed, ReasonRejected, ResolutionExact
	}
	return DecisionPass, ReasonExact, ResolutionExact
}
