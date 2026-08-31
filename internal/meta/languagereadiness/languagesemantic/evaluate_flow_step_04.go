package languagesemantic

import (
	"os"
)

func evaluateFlowStep04(flow *evaluateFlowState) {
	flow.slot04, flow.slot03 = os.ReadFile(flow.slot00.SyntaxArtifactPath)
}
