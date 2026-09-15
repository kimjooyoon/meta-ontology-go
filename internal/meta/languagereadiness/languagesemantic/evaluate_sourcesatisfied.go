package languagesemantic

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic/replay"
	"slices"
)

func sourceSatisfied(observation replay.Observation) bool {
	return observation.Normalized && observation.CanonicalReplay && observation.SemanticReplay && observation.ProvenanceReplay && observation.EvidenceReplay &&
		slices.Equal(observation.Stages, replay.ExpectedStages) && observation.Effects.Writes == 0 && observation.Effects.Network == 0 && observation.Effects.Processes == 0
}
