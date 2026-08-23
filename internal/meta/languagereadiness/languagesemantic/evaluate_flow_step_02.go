package languagesemantic

func evaluateFlowStep02(flow *evaluateFlowState) {
	if flow.slot03 != nil {
		{
			flow.result0, flow.result1 = Report{}, flow.slot03
			flow.done = true
			return
		}
	}
}
