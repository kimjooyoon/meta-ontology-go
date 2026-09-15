package replay

func lawsFlowStep15(flow *lawsFlowState) {
	flow.slot09, flow.slot03 = flow.slot08.Normalized()
}
