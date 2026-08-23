package replay

func lawsFlowStep12(flow *lawsFlowState) {
	if err := flow.slot06.Validate(); err != nil {
		{
			flow.result0, flow.result1 = LawObservation{}, err
			flow.done = true
			return
		}
	}
}
