package languagesemantic

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic/replay"
	"slices"
)

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
