package verticalsliceclosureshadow

import (
	languagesemantic "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic"
	languagesyntax "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	toolchainconformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainconformance"
)

func observeBoundary(id string, artifact artifactEnvelope) (int, string) {
	summary := artifact.Summary
	switch id {
	case "syntax":
		return summary.CapabilitySatisfied, status(summary.CapabilitySatisfied == languagesyntax.FixedCapabilityTotal &&
			summary.CapabilityTotal == languagesyntax.FixedCapabilityTotal &&
			summary.CapabilityExecuted == languagesyntax.FixedCapabilityTotal && summary.CapabilityUnresolved == 0 &&
			summary.GovernanceSatisfied == languagesyntax.FixedGovernanceTotal &&
			summary.GovernanceTotal == languagesyntax.FixedGovernanceTotal &&
			summary.GovernanceExecuted == languagesyntax.FixedGovernanceTotal && summary.GovernanceUnresolved == 0 &&
			summary.Satisfied == languagesyntax.FixedTotal && summary.Total == languagesyntax.FixedTotal &&
			summary.Executed == languagesyntax.FixedTotal && summary.NotSatisfied == 0 && summary.Unresolved == 0 &&
			summary.ReadinessBPS == 10000)
	case "semantics":
		return summary.Satisfied, status(summary.Satisfied == languagesemantic.FixedTotal && summary.Total == languagesemantic.FixedTotal &&
			summary.Executed == languagesemantic.FixedTotal && summary.NotSatisfied == 0 && summary.Unresolved == 0 &&
			summary.ReadinessBPS == 10000 && summary.StageOrderViolations == 0 &&
			summary.EffectfulStages == 0 && summary.RegistryDrift == 0)
	case "binding":
		return summary.BoundCoordinates, status(summary.Coordinates == 12 &&
			summary.BoundCoordinates == 12 && summary.Unresolved == 0 &&
			summary.ReadinessCompleted == 21 && summary.ReadinessTotal == 24 &&
			summary.ReadinessBPS == 8750 && summary.SemanticSatisfied == languagesemantic.FixedTotal &&
			summary.SemanticTotal == languagesemantic.FixedTotal && summary.EffectfulStages == 0 &&
			summary.MutationAuthorities == 0)
	case "use-cases":
		return summary.Satisfied, status(summary.Satisfied == 3 && summary.Total == 3 &&
			summary.Executed == 3 && summary.NotSatisfied == 0 &&
			summary.Unresolved == 0 && summary.ReadinessBPS == 10000)
	case "toolchain":
		return summary.CasesSatisfied, status(summary.SurfacesSatisfied == 9 &&
			summary.SurfacesTotal == 9 && len(artifact.Surfaces) == 9 &&
			summary.CasesSatisfied == toolchainconformance.ExpectedCaseCount && summary.CasesTotal == toolchainconformance.ExpectedCaseCount &&
			summary.ExecutedCases == toolchainconformance.ExpectedCaseCount && summary.CaseReadinessBPS == 10000 &&
			summary.IndicatorsSatisfied == summary.IndicatorsTotal &&
			summary.ProofsPassed == summary.ProofsTotal &&
			summary.TamperRejections == 13 && summary.TamperTotal == 13)
	case "release":
		return summary.CasesSatisfied, status(summary.CasesSatisfied == 20 &&
			summary.CasesTotal == 20 && len(artifact.Cases) == 20 &&
			summary.ReadinessBPS == 10000 && summary.PlatformReceipts == 3 &&
			summary.OperatingSystems == 3 && allReleaseCasesExact(artifact.Cases))
	default:
		return 0, StatusBlocked
	}
}

func status(satisfied bool) string {
	if satisfied {
		return StatusSatisfied
	}
	return StatusBlocked
}

func allReleaseCasesExact(cases []artifactCase) bool {
	for _, item := range cases {
		if item.Observed == "" || item.Observed != item.Expected {
			return false
		}
	}
	return true
}
