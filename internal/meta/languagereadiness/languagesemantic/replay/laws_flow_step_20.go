package replay

func lawsFlowStep20(flow *lawsFlowState) {
	if err := flow.slot09.Validate(); err != nil {
		{
			flow.result0, flow.result1 = LawObservation{}, err
			flow.done = true
			return
		}
	}
}
