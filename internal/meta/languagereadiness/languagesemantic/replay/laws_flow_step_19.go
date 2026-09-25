package replay

func lawsFlowStep19(flow *lawsFlowState) {
	if err := flow.slot09.AddCandidate(flow.slot10); err != nil {
		{
			flow.result0, flow.result1 = LawObservation{}, err
			flow.done = true
			return
		}
	}
}
