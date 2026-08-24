package externalecosystemconformance

func makeIndicators(report Report) []Indicator {
	s := report.Summary
	return []Indicator{
		metric("gooo.metric.ecosystem.reference-bound.v1", "OUTCOME", "COHERENCE", "bind-external-reference", "capabilities", "GREATER_OR_EQUAL", s.BoundCapabilities, 8, report.Resolution),
		metric("gooo.metric.ecosystem.commit-pin.v1", "DRIVER", "FOUNDATION", "verify-immutable-commit", "commits", "GREATER_OR_EQUAL", s.CommitExact, 1, report.Resolution),
		metric("gooo.metric.ecosystem.document-exactness.v1", "DRIVER", "FOUNDATION", "verify-reference-documents", "documents", "GREATER_OR_EQUAL", s.DocumentsExact, 2, report.Resolution),
		metric("gooo.metric.ecosystem.module-contract.v1", "DRIVER", "FOUNDATION", "verify-reference-module", "modules", "GREATER_OR_EQUAL", s.ModuleExact, 1, report.Resolution),
		metric("gooo.metric.ecosystem.capability-mapping.v1", "DRIVER", "COHERENCE", "map-reference-capabilities", "capabilities", "GREATER_OR_EQUAL", s.CapabilityMappings, 8, report.Resolution),
		metric("gooo.metric.ecosystem.unknown.v1", "GUARDRAIL", "REGRESSION", "preserve-unknown-reference", "paths", "LESS_OR_EQUAL", s.UnknownPaths, 0, report.Resolution),
		metric("gooo.metric.ecosystem.external-execution.v1", "GUARDRAIL", "FOUNDATION", "deny-external-execution", "executions", "LESS_OR_EQUAL", s.ObservedExecutions, 0, report.Resolution),
		metric("gooo.metric.ecosystem.repository-writes.v1", "GUARDRAIL", "FOUNDATION", "preserve-read-only-reference", "writes", "LESS_OR_EQUAL", s.ObservedWrites, 0, report.Resolution),
		metric("gooo.metric.ecosystem.official-mutations.v1", "GUARDRAIL", "COHERENCE", "preserve-official-denominator", "mutations", "LESS_OR_EQUAL", s.OfficialMutations, 0, report.Resolution),
	}
}

func metric(id, class, proof, operation, unit, relation string, value, target int, resolution string) Indicator {
	satisfied := value >= target
	if relation == "LESS_OR_EQUAL" {
		satisfied = value <= target
	}
	return Indicator{
		MetricID: id, Class: class, ProofChoice: proof,
		Producer: "externalecosystemconformance.Evaluate",
		Consumer: "external-ecosystem-reference-gate",
		MetaOperation: operation, Unit: unit, Relation: relation,
		Resolution: resolution, Value: value, Target: target, Satisfied: satisfied,
	}
}
