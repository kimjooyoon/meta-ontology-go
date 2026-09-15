package languagesemantic

func buildIndicators(summary Summary, resolution Resolution) []Indicator {
	type metric struct {
		id, class, proof, operation string
		value, target               int
	}
	metrics := []metric{
		{"semantic-model-readiness-bps", "OUTCOME", "COHERENCE", "prove-staged-semantic-model", summary.ReadinessBPS, 10000},
		{"semantic-model-executed-cases", "DRIVER", "FOUNDATION", "bind-versioned-semantic-corpus", summary.Executed, FixedTotal},
		{"semantic-model-source-models", "DRIVER", "FOUNDATION", "lower-normalize-replay", summary.SourceModels, expectedSources},
		{"semantic-model-normalized-irs", "DRIVER", "COHERENCE", "lower-normalize-replay", summary.NormalizedIRs, expectedSources},
		{"semantic-model-semantic-replays", "DRIVER", "COHERENCE", "replay-authoritative-meaning", summary.SemanticReplays, expectedSources},
		{"semantic-model-provenance-replays", "DRIVER", "COHERENCE", "replay-provenance", summary.ProvenanceReplays, expectedSources},
		{"semantic-model-evidence-replays", "DRIVER", "COHERENCE", "replay-exact-evidence", summary.EvidenceReplays, expectedSources},
		{"semantic-model-presentation-laws", "DRIVER", "COHERENCE", "prove-presentation-invariance", summary.PresentationLaws, 1},
		{"semantic-model-candidate-authority-laws", "DRIVER", "REGRESSION", "prove-candidate-non-authority", summary.CandidateAuthorityLaws, 1},
		{"semantic-model-deterministic-authority-laws", "DRIVER", "COHERENCE", "prove-deterministic-authority", summary.DeterministicAuthorityLaws, 1},
		{"semantic-model-upstream-rejections", "DRIVER", "REGRESSION", "inherit-fail-closed-syntax", summary.UpstreamRejections, expectedRejections},
		{"semantic-model-unregistered-gooo.guardrail", "GUARDRAIL", "FOUNDATION", "bind-complete-source-set", summary.UnregisteredGooo, 0},
		{"semantic-model-missing-registered.guardrail", "GUARDRAIL", "FOUNDATION", "bind-complete-source-set", summary.MissingRegistered, 0},
		{"semantic-model-unresolved.guardrail", "GUARDRAIL", "FOUNDATION", "lower-semantic-resolution", summary.Unresolved, 0},
		{"semantic-model-stage-order.guardrail", "GUARDRAIL", "COHERENCE", "enforce-stage-order", summary.StageOrderViolations, 0},
		{"semantic-model-effects.guardrail", "GUARDRAIL", "REGRESSION", "seal-effect-receipts", summary.EffectfulStages, 0},
		{"semantic-model-repository-writes.guardrail", "GUARDRAIL", "REGRESSION", "preserve-read-only-observer", 0, 0},
		{"semantic-model-mutation-authority.guardrail", "GUARDRAIL", "REGRESSION", "deny-observer-mutation-authority", 0, 0},
		{"semantic-model-registry-drift.guardrail", "GUARDRAIL", "FOUNDATION", "bind-versioned-semantic-corpus", summary.RegistryDrift, 0},
	}
	indicators := make([]Indicator, 0, len(metrics))
	for _, item := range metrics {
		indicators = append(indicators, Indicator{
			MetricID:      "gooo.metric.language." + item.id + ".v1",
			Class:         item.class,
			ProofChoice:   item.proof,
			Producer:      "languagesemantic.Evaluate",
			Consumer:      "self-improvement-cycle",
			MetaOperation: item.operation,
			Resolution:    resolution,
			Value:         item.value,
			Target:        item.target,
			Satisfied:     item.value == item.target,
		})
	}
	return indicators
}
