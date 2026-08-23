package toolchainformatfix

func indicators(summary Summary, resolution Resolution) []Indicator {
	return []Indicator{
		metric("readiness-bps", "OUTCOME", "COHERENCE", summary.ReadinessBPS, 10000, resolution),
		metric("positive-paths", "OUTCOME", "COHERENCE", summary.PositivePaths, FixedPositive, resolution),
		metric("guardrail-rejections", "OUTCOME", "REGRESSION", summary.GuardrailRejections, FixedGuardrails, resolution),
		metric("executed-cases", "DRIVER", "FOUNDATION", summary.Executed, FixedTotal, resolution),
		metric("invocations", "DRIVER", "COHERENCE", summary.Invocations, ExpectedRuns, resolution),
		metric("structured-outputs", "DRIVER", "COHERENCE", summary.StructuredOutputs, ExpectedStructured, resolution),
		metric("structured-plans", "DRIVER", "FOUNDATION", summary.StructuredPlans, ExpectedPlans, resolution),
		metric("in-memory-applications", "DRIVER", "COHERENCE", summary.InMemoryApplications, 1, resolution),
		metric("fixed-points", "DRIVER", "COHERENCE", summary.FixedPoints, 2, resolution),
		metric("deterministic-replays", "DRIVER", "REGRESSION", summary.ReplayMatches, FixedTotal, resolution),
		metric("binary-bindings", "DRIVER", "FOUNDATION", summary.BinaryBindings, 1, resolution),
		metric("unresolved.guardrail", "GUARDRAIL", "FOUNDATION", summary.Unresolved, 0, resolution),
		metric("exit-mismatch.guardrail", "GUARDRAIL", "REGRESSION", summary.ExitMismatches, 0, resolution),
		metric("output-mismatch.guardrail", "GUARDRAIL", "REGRESSION", summary.OutputMismatches, 0, resolution),
		metric("replay-mismatch.guardrail", "GUARDRAIL", "REGRESSION", summary.ReplayMismatches, 0, resolution),
		metric("repository-writes.guardrail", "GUARDRAIL", "REGRESSION", summary.RepositoryWrites, 0, resolution),
		metric("direct-writes.guardrail", "GUARDRAIL", "COHERENCE", summary.DirectWrites, 0, resolution),
		metric("registry-drift.guardrail", "GUARDRAIL", "FOUNDATION", summary.RegistryDrift, 0, resolution),
	}
}

func metric(name, class, proof string, value, target int, resolution Resolution) Indicator {
	satisfied := resolution == ResolutionExact && value == target
	return Indicator{MetricID: "gooo.metric.toolchain.format-fix-" + name + ".v1",
		Class: class, ProofChoice: proof, MetaOperation: "evaluate-toolchain-format-fix",
		Value: value, Target: target, Resolution: resolution, Satisfied: satisfied}
}
