package languagesemantic

func Evaluate(input Input) (Report, error) {
	flow := &evaluateFlowState{slot00: input}
	for _, step := range evaluateFlowSteps {
		step(flow)
		if flow.done {
			break
		}
	}
	return flow.result0, flow.result1
}
