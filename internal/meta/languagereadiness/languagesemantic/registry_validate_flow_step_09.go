package languagesemantic

func registry_validateFlowStep09(flow *registry_validateFlowState) {
	for name := range knownLaws {
		flow.slot05 = append(flow.slot05, name)
	}
}
