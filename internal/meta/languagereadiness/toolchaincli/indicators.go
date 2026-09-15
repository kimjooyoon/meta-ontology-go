package toolchaincli

func indicators(summary Summary, resolution Resolution, mutationAuthorized bool) []Indicator {
	return []Indicator{
		metric("readiness-bps", "OUTCOME", "COHERENCE", "evaluate-cli-contract", summary.ReadinessBPS, 10000, resolution),
		metric("positive-paths", "OUTCOME", "COHERENCE", "execute-cli-positive-paths", summary.PositivePaths, FixedPositive, resolution),
		metric("guardrail-rejections", "OUTCOME", "REGRESSION", "reject-invalid-cli-invocations", summary.GuardrailRejections, FixedGuardrails, resolution),
		metric("executed-cases", "DRIVER", "FOUNDATION", "execute-fixed-cli-corpus", summary.Executed, FixedTotal, resolution),
		metric("invocations", "DRIVER", "COHERENCE", "invoke-bound-cli-binary", summary.Invocations, ExpectedRuns, resolution),
		metric("declared-commands", "DRIVER", "FOUNDATION", "observe-command-surface", summary.DeclaredCommands, ExpectedCommands, resolution),
		metric("structured-outputs", "DRIVER", "COHERENCE", "decode-cli-json", summary.StructuredOutputs, ExpectedStructured, resolution),
		metric("language-operations", "DRIVER", "COHERENCE", "execute-language-operations", summary.LanguageOperations, ExpectedLanguageOps, resolution),
		metric("deterministic-replays", "DRIVER", "REGRESSION", "replay-cli-observations", summary.ReplayMatches, FixedTotal, resolution),
		metric("binary-bindings", "DRIVER", "FOUNDATION", "bind-cli-binary-digest", summary.BinaryBindings, 1, resolution),
		metric("unresolved.guardrail", "GUARDRAIL", "FOUNDATION", "lower-cli-resolution", summary.Unresolved, 0, resolution),
		metric("exit-mismatch.guardrail", "GUARDRAIL", "REGRESSION", "verify-cli-exit-contract", summary.ExitMismatches, 0, resolution),
		metric("stdout-mismatch.guardrail", "GUARDRAIL", "REGRESSION", "verify-cli-stdout-contract", summary.StdoutMismatches, 0, resolution),
		metric("stderr-mismatch.guardrail", "GUARDRAIL", "REGRESSION", "verify-cli-stderr-contract", summary.StderrMismatches, 0, resolution),
		metric("replay-mismatch.guardrail", "GUARDRAIL", "REGRESSION", "reject-cli-replay-drift", summary.ReplayMismatches, 0, resolution),
		metric("repository-writes.guardrail", "GUARDRAIL", "REGRESSION", "seal-cli-repository-effects", summary.RepositoryWrites, 0, resolution),
		metric("mutation-authority.guardrail", "GUARDRAIL", "REGRESSION", "deny-cli-mutation-authority", boolInt(mutationAuthorized), 0, resolution),
		metric("registry-drift.guardrail", "GUARDRAIL", "FOUNDATION", "bind-cli-case-registry", summary.RegistryDrift, 0, resolution),
	}
}

func metric(suffix, class, proof, operation string, value, target int, resolution Resolution) Indicator {
	return Indicator{MetricID: "gooo.metric.toolchain.cli-" + suffix + ".v1", Class: class,
		ProofChoice: proof, MetaOperation: operation, Producer: "toolchain-cli-witness",
		Consumer: "language-readiness-witness", Value: value, Target: target,
		Resolution: resolution, Satisfied: resolution == ResolutionExact && value == target}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
