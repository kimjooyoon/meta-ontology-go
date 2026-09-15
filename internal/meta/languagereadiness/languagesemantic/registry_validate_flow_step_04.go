package languagesemantic

func registry_validateFlowStep04(flow *registry_validateFlowState) {
	flow.slot01 = make(map[string]struct{}, len(flow.slot00.Cases))
}
