package languagesemantic

import (
	"encoding/json"
)

func evaluateFlowStep06(flow *evaluateFlowState) {
	if err := json.Unmarshal(flow.slot04, &flow.slot05); err != nil {
		{
			flow.result0, flow.result1 = unresolvedReport(flow.slot01, flow.slot00.ExpectedHeadSHA, digestBytes(flow.slot02), "syntax artifact invalid: "+err.Error()), nil
			flow.done = true
			return
		}
	}
}
