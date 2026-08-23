package languagesemantic

func evaluateFlowStep03(flow *evaluateFlowState) {
	if err := validateHeadSHA(flow.slot00.ExpectedHeadSHA); err != nil {
		{
			flow.result0, flow.result1 = Report{}, err
			flow.done = true
			return
		}
	}
}
