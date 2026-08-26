package languagesemantic

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesemantic/replay"
)

func evaluateFlowStep18(flow *evaluateFlowState) {
	if flow.slot13 == nil {
		flow.slot15 = fmt.Errorf("no source model contains a deterministic fact")
	} else {
		flow.slot14, flow.slot15 = replay.ObserveLaws(flow.slot13.Path, flow.slot13.IR)
	}
}
