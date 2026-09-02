package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func bindClaimDependencyEdges(report claimDependencyReport, parsed []parsedClaimDependencyActivity, producers map[string]*syntax.ActivityDecl) (claimDependencyReport, bool) {
	for _, item := range parsed {
		activity := item.declaration
		if item.program.role == claimDependencyRootRole {
			for _, input := range activity.Inputs {
				if producer, found := producers[input.Name]; found {
					report = refuteClaimDependencies(report, "FOUNDATION", "VERIFY_RECOVERABLE_ROOT", "RECOVERABLE_ROOT_HAS_CLAIM_PREDECESSOR", "REMOVE_PREDECESSOR_OR_SELECT_COHERENCE", claimDependencyRegression, []string{producer.Name, activity.Name})
					return report, false
				}
			}
			continue
		}
		if len(activity.Inputs) == 0 {
			report = refuteClaimDependencies(report, "CLAIM_DEPENDENCY", "BIND_INPUT_PRODUCERS", "TYPED_DEPENDENCY_INPUT_MISSING", "DECLARE_DEPENDENCY_INPUT", claimDependencyRegression, []string{activity.Name})
			return report, false
		}
		for _, input := range activity.Inputs {
			report.Summary.DependencyInputs++
			producer, found := producers[input.Name]
			if !found {
				report.Gaps = append(report.Gaps, claimDependencyGap{
					Activity: activity.Name, InputEntity: input.Name, Stage: "DEPENDENCY_DISCOVERY", Step: "BIND_INPUT_PRODUCER",
					Reason: "CLAIM_INPUT_PRODUCER_UNAVAILABLE", UnknownClass: "DIRECT_MISSING", NextOperation: "DECLARE_INPUT_PRODUCER",
				})
				continue
			}
			edgeOrdinal := len(report.Edges) + 1
			report.Edges = append(report.Edges, claimDependencyEdge{
				Ordinal: edgeOrdinal, ID: fmt.Sprintf("gooo://claim-dependency/edge/%s/%03d", report.Subject.Namespace, edgeOrdinal),
				Kind: item.program.kind, Label: item.program.label, FromActivity: producer.Name, ToActivity: activity.Name, ViaEntity: input.Name,
			})
		}
	}
	report.Summary.TypedEdges = len(report.Edges)
	report.Summary.UnresolvedInputs = len(report.Gaps)
	countClaimDependencyKinds(&report)
	if len(report.Gaps) == 0 {
		return report, true
	}
	blocked := make([]string, 0, len(report.Gaps))
	for _, gap := range report.Gaps {
		blocked = append(blocked, "entity:"+gap.InputEntity)
	}
	report = unknownClaimDependencies(report, "DEPENDENCY_DISCOVERY", "BIND_INPUT_PRODUCER", "CLAIM_INPUT_PRODUCER_UNAVAILABLE", "DIRECT_MISSING", "DECLARE_INPUT_PRODUCER", claimDependencyCoherence, blocked)
	return report, false
}
