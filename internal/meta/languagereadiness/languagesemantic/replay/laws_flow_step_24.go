package replay

func lawsFlowStep24(flow *lawsFlowState) {
	if err := flow.slot11.Validate(); err != nil {
		{
			flow.result0, flow.result1 = LawObservation{}, err
			flow.done = true
			return
		}
	}
}
