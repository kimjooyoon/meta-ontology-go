package languagesemantic

import (
	"path/filepath"
)

func evaluateFlowStep08(flow *evaluateFlowState) {
	flow.slot06, flow.slot03 = filepath.Abs(flow.slot00.Root)
}
