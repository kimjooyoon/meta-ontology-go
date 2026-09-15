package replay

func lawsFlowStep04(flow *lawsFlowState) {
	flow.slot05 = flow.slot02.Graph.DeterministicFacts()
}
