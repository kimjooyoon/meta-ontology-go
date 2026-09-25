package languagesemantic

func evaluateFlowStep01(flow *evaluateFlowState) {
	flow.slot01, flow.slot02, flow.slot03 = LoadRegistry(flow.slot00.RegistryPath)
}
