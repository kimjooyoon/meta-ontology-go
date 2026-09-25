package languagesemantic

func evaluateFlowStep05(flow *evaluateFlowState) {
	if flow.slot03 != nil {
		{
			flow.result0, flow.result1 = unresolvedReport(flow.slot01, flow.slot00.ExpectedHeadSHA, digestBytes(flow.slot02), "syntax artifact unavailable: "+flow.slot03.Error()), nil
			flow.done = true
			return
		}
	}
}
