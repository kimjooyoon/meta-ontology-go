package replay

func lawsFlowStep13(flow *lawsFlowState) {
	flow.slot08, flow.slot03 = structureOnly(flow.slot02)
}
