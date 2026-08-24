package userjourneyscorecard

func (s *inspection) indicators() []Indicator {
	expectedSamples := len(s.contract.Journeys) * s.contract.SamplesPerJourney
	effects := s.repositoryWrites + boolInt(s.upstream.MutationAuthorized)
	return []Indicator{
		metric("functional.cli-contract", "OUTCOME", "FOUNDATION", "consume-toolchain-cli-receipt", boolInt(s.upstreamPassed), 1),
		metric("functional.user-journeys", "OUTCOME", "COHERENCE", "execute-user-positive-paths", int64(s.journeysPassed), 6),
		metric("resource.envelopes", "OUTCOME", "COHERENCE", "reduce-runner-resource-envelopes", int64(s.envelopesPassed), 6),
		metric("profile.samples", "DRIVER", "FOUNDATION", "observe-resource-samples", int64(s.samplesObserved), int64(expectedSamples)),
		metric("profile.output-replay", "DRIVER", "REGRESSION", "compare-user-visible-output", int64(s.outputReplays), 6),
		metric("functional.structured-output", "DRIVER", "COHERENCE", "decode-user-json-output", int64(s.upstream.Summary.StructuredOutputs), 3),
		metric("functional.language-operations", "DRIVER", "COHERENCE", "execute-language-operations", int64(s.upstream.Summary.LanguageOperations), 4),
		metric("functional.declared-commands", "DRIVER", "FOUNDATION", "observe-cli-surface", int64(s.upstream.Summary.DeclaredCommands), 13),
		metric("binding.binary", "DRIVER", "FOUNDATION", "bind-profiled-executable", boolInt(s.binaryBound && s.sourceBound), 1),
		metric("binding.meta-operations", "DRIVER", "FOUNDATION", "join-functional-and-resource-meta-operations", int64(s.metaBindings), 18),
		metric("guardrail.unknown", "GUARDRAIL", "FOUNDATION", "lower-unknown-observation", int64(s.unknowns), 0),
		metric("guardrail.wall", "GUARDRAIL", "REGRESSION", "bound-wall-observations", int64(s.wallViolations), 0),
		metric("guardrail.rss", "GUARDRAIL", "REGRESSION", "bound-rss-observations", int64(s.rssViolations), 0),
		metric("guardrail.binary-size", "GUARDRAIL", "REGRESSION", "bound-binary-size", int64(s.binaryViolations), 0),
		metric("guardrail.effects", "GUARDRAIL", "REGRESSION", "deny-repository-mutation", int64(effects), 0),
	}
}

func metric(id, class, proof, operation string, observed, expected int64) Indicator {
	return Indicator{ID: id, Class: class, ProofChoice: proof, MetaOperation: operation,
		Observed: observed, Expected: expected, Satisfied: observed == expected}
}

func boolInt(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
