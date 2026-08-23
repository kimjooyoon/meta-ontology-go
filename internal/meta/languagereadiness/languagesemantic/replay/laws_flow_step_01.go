package replay

func lawsFlowStep01(flow *lawsFlowState) {
	flow.slot02, flow.slot03 = flow.slot01.Normalized()
}
