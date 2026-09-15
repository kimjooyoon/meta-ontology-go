package replay

func lawsFlowStep23(flow *lawsFlowState) {
	if err := flow.slot11.AddFact(flow.slot05[0]); err != nil {
		{
			flow.result0, flow.result1 = LawObservation{}, err
			flow.done = true
			return
		}
	}
}
