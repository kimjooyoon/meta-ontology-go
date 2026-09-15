package languagesemantic

func registry_validateFlowStep06(flow *registry_validateFlowState) {
	for _, definition := range flow.slot00.Cases {
		if registryValidateDefinitionStep(flow, definition) {
			return
		}
	}
}
