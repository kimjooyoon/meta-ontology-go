package replay

func lawsFlowStep07(flow *lawsFlowState) {
	flow.slot06, flow.slot03 = flow.slot02.Normalized()
}
