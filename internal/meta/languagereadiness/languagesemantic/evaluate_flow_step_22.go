package languagesemantic

func evaluateFlowStep20(flow *evaluateFlowState) {
	for _, item := range flow.slot05.Cases {
		flow.slot16[item.Definition.ID] = item
	}
}
