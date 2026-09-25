package languagesemantic

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic/replay"
)

func evaluateFlowStep15(flow *evaluateFlowState) {
	flow.slot12 = make(map[string]replay.Observation, expectedSources)
}
