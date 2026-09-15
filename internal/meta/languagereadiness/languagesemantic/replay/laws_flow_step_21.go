package replay

func lawsFlowStep21(flow *lawsFlowState) {
	flow.slot11, flow.slot03 = flow.slot08.Normalized()
}
