package replay

func lawsFlowStep11(flow *lawsFlowState) {
	if err := flow.slot06.Graph.AddNode(flow.slot07); err != nil {
		{
			flow.result0, flow.result1 = LawObservation{}, err
			flow.done = true
			return
		}
	}
}
