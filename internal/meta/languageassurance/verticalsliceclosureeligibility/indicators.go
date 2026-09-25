package verticalsliceclosureeligibility

func buildIndicators(report Report) []Indicator {
	summary := report.Summary
	assuranceExact, shadowExact := artifactValue(report.Artifacts, AssuranceName), artifactValue(report.Artifacts, ShadowName)
	return []Indicator{
		indicator("gooo.metric.capability.vertical-slice-eligibility.v1", "OUTCOME", "COHERENCE",
			MetaOperation, "paths", "GREATER_OR_EQUAL", report.Resolution, summary.EligiblePaths, 1),
		indicator("gooo.metric.evidence.vertical-slice-assurance-capsule.v1", "DRIVER", "FOUNDATION",
			"consume-fixed-assurance", "capsules", "GREATER_OR_EQUAL", report.Resolution, assuranceExact, 1),
		indicator("gooo.metric.evidence.vertical-slice-shadow-capsule.v1", "DRIVER", "FOUNDATION",
			"consume-merged-vertical-shadow", "capsules", "GREATER_OR_EQUAL", report.Resolution, shadowExact, 1),
		indicator("gooo.metric.evidence.vertical-slice-boundaries.v1", "DRIVER", "COHERENCE",
			"bind-vertical-slice-boundaries", "boundaries", "GREATER_OR_EQUAL", report.Resolution,
			summary.BoundariesSatisfied, 6),
		indicator("gooo.metric.evidence.vertical-slice-links.v1", "DRIVER", "COHERENCE",
			"bind-vertical-slice-links", "links", "GREATER_OR_EQUAL", report.Resolution,
			summary.LinksSatisfied, 12),
		indicator("gooo.metric.epistemic.vertical-slice-eligibility-unknown.v1", "GUARDRAIL", "REGRESSION",
			"preserve-unknown-decision", "paths", "LESS_OR_EQUAL", ResolutionExact, summary.UnknownPaths, 0),
		indicator("gooo.metric.effects.vertical-slice-eligibility-writes.v1", "GUARDRAIL", "FOUNDATION",
			"preserve-read-only-eligibility", "writes", "LESS_OR_EQUAL", ResolutionExact,
			summary.ObservedRepositoryWrites, 0),
		indicator("gooo.metric.capability.vertical-slice-eligibility-applied.v1", "GUARDRAIL", "FOUNDATION",
			"deny-eligibility-side-effects", "transitions", "LESS_OR_EQUAL", ResolutionExact,
			report.PromotionApplied, 0),
	}
}

func artifactValue(values []ArtifactBinding, name string) int {
	for _, value := range values {
		if value.Name == name && value.Exact {
			return 1
		}
	}
	return 0
}

func indicator(id, class, proof, operation, unit, relation, resolution string, value, target int) Indicator {
	satisfied := value <= target
	if relation == "GREATER_OR_EQUAL" {
		satisfied = value >= target
	}
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "verticalsliceclosureeligibility.Evaluate",
		Consumer: "language-assurance-activation-gate", MetaOperation: operation,
		Unit: unit, Relation: relation, Resolution: resolution,
		Value: value, Target: target, Satisfied: satisfied}
}
