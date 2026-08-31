package replay

func lawsFlowStep14(flow *lawsFlowState) {
	if flow.slot03 != nil {
		{
			flow.result0, flow.result1 = LawObservation{}, flow.slot03
			flow.done = true
			return
		}
	}
}
