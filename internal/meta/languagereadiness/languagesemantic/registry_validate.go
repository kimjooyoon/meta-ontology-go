package languagesemantic

func (registry Registry) Validate() error {
	flow := &registry_validateFlowState{slot00: registry}
	for _, step := range registry_validateFlowSteps {
		step(flow)
		if flow.done {
			break
		}
	}
	return flow.result0
}
