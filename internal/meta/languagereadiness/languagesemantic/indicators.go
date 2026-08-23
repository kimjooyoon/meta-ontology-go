package languagesemantic

import (
	"fmt"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic/replay"
)

func buildReport(_ Registry, cases []CaseResult, source Source, unregistered, missing, registryDrift int) Report {
	summary := summarize(cases, unregistered, missing, registryDrift)
	resolution := ResolutionExact
	if !source.ObservationKnown || summary.Unresolved > 0 || unregistered > 0 || missing > 0 || registryDrift > 0 {
		resolution = ResolutionLower
	}
	indicators := buildIndicators(summary, resolution)
	decision, reason := DecisionPass, "SEMANTIC_MODEL_EXACTLY_PROVEN"
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			decision = DecisionFailClosed
			reason = "SEMANTIC_MODEL_CONFORMANCE_FAILED"
			if resolution == ResolutionLower {
				reason = "SEMANTIC_MODEL_EVIDENCE_UNKNOWN"
			}
			break
		}
	}
	proofs := buildProofs(summary, source, resolution)
	report := Report{
		Schema:             ReportSchema,
		Decision:           decision,
		Resolution:         resolution,
		ReasonCode:         reason,
		Source:             source,
		Summary:            summary,
		Cases:              cases,
		Indicators:         indicators,
		Proofs:             proofs,
		RepositoryWrites:   0,
		MutationAuthorized: false,
	}
	finalizeReport(&report)
	return report
}

func summarize(cases []CaseResult, unregistered, missing, registryDrift int) Summary {
	summary := Summary{Total: FixedTotal, UnregisteredGooo: unregistered, MissingRegistered: missing, RegistryDrift: registryDrift}
	for _, result := range cases {
		switch result.Status {
		case StatusSatisfied:
			summary.Satisfied++
			summary.Executed++
		case StatusNotSatisfied:
			summary.NotSatisfied++
			summary.Executed++
		case StatusUnresolved:
			summary.Unresolved++
		}
		if result.Definition.Kind == CaseSource && result.Evidence.Source != nil {
			observation := result.Evidence.Source
			if result.Status == StatusSatisfied {
				summary.SourceModels++
			}
			if observation.Normalized {
				summary.NormalizedIRs++
			}
			if observation.SemanticReplay {
				summary.SemanticReplays++
			}
			if observation.ProvenanceReplay {
				summary.ProvenanceReplays++
			}
			if observation.EvidenceReplay {
				summary.EvidenceReplays++
			}
			if !slices.Equal(observation.Stages, replay.ExpectedStages) {
				summary.StageOrderViolations++
			}
			summary.EffectfulStages += observation.Effects.Writes + observation.Effects.Network + observation.Effects.Processes
		}
		if result.Status == StatusSatisfied && result.Definition.Kind == CaseLaw {
			switch result.Definition.Law {
			case "PRESENTATION_INVARIANCE":
				summary.PresentationLaws++
			case "CANDIDATE_NON_AUTHORITY":
				summary.CandidateAuthorityLaws++
			case "DETERMINISTIC_AUTHORITY":
				summary.DeterministicAuthorityLaws++
			}
		}
		if result.Status == StatusSatisfied && result.Definition.Kind == CaseUpstreamRejection {
			summary.UpstreamRejections++
		}
	}
	summary.ReadinessBPS = summary.Satisfied * 10000 / summary.Total
	return summary
}

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

func buildProofs(summary Summary, source Source, resolution Resolution) []Proof {
	values := []struct {
		choice, operation, evidence string
		passed                      bool
	}{
		{"FOUNDATION", "bind-versioned-semantic-corpus", fmt.Sprintf("%s|%s|%d|%d|%d", source.RegistryDigest, source.SyntaxArtifactDigest, summary.UnregisteredGooo, summary.MissingRegistered, summary.RegistryDrift), source.ObservationKnown && source.ConceptBound && summary.UnregisteredGooo == 0 && summary.MissingRegistered == 0 && summary.RegistryDrift == 0},
		{"COHERENCE", "replay-normalized-authoritative-meaning", fmt.Sprintf("%d|%d|%d|%d|%d|%d", summary.NormalizedIRs, summary.SemanticReplays, summary.ProvenanceReplays, summary.EvidenceReplays, summary.PresentationLaws, summary.DeterministicAuthorityLaws), summary.NormalizedIRs == expectedSources && summary.SemanticReplays == expectedSources && summary.ProvenanceReplays == expectedSources && summary.EvidenceReplays == expectedSources && summary.PresentationLaws == 1 && summary.DeterministicAuthorityLaws == 1},
		{"REGRESSION", "reject-unknown-effects-and-candidate-authority", fmt.Sprintf("%d|%d|%d|%d", summary.UpstreamRejections, summary.CandidateAuthorityLaws, summary.EffectfulStages, summary.Unresolved), summary.UpstreamRejections == expectedRejections && summary.CandidateAuthorityLaws == 1 && summary.EffectfulStages == 0 && summary.Unresolved == 0 && resolution == ResolutionExact},
	}
	proofs := make([]Proof, 0, len(values))
	for _, value := range values {
		proofs = append(proofs, Proof{Choice: value.choice, MetaOperation: value.operation, EvidenceDigest: semanticHash(value.evidence), Passed: value.passed})
	}
	return proofs
}
