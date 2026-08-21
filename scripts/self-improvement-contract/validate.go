package main

type stageContract struct {
	Name   string
	Inputs []string
	Output string
}

var loopStages = []stageContract{
	{Name: "Measure", Inputs: []string{"SourceTree"}, Output: "SourceMetrics"},
	{Name: "Plan", Inputs: []string{"SourceMetrics", "ProofChoice"}, Output: "GenerationPlan"},
	{Name: "Evaluate", Inputs: []string{"ExecutionManifest"}, Output: "Receipt"},
	{Name: "Verify", Inputs: []string{"Receipt"}, Output: "Evidence"},
	{Name: "Improve", Inputs: []string{"Evidence"}, Output: "SourceTree"},
}

func closedLoop(model contractModel) bool {
	for _, stage := range loopStages {
		if !activityMatches(model.Activities[stage.Name], stage.Inputs, stage.Output) {
			return false
		}
	}
	execute, exists := model.Activities["Execute"]
	return exists && len(execute.Inputs) > 0 &&
		execute.Inputs[0] == "GenerationPlan" &&
		execute.Output == "ExecutionManifest"
}

func trilemmaChoice(model contractModel, registry []RegistryBinding) bool {
	choice := model.Activities["ChooseProof"]
	if !activityMatches(choice,
		[]string{"Foundation", "Coherence", "Regression"}, "ProofChoice") {
		return false
	}
	expected := map[string]string{
		"Foundation": "proof://foundation",
		"Coherence":  "proof://coherence",
		"Regression": "proof://regression",
		"ProofChoice": "gooo://meta/proof-choice",
	}
	for name, id := range expected {
		if model.Entities[name] != id {
			return false
		}
	}
	for _, binding := range registry {
		if model.EntityByID[binding.ProofEntityID] == "" {
			return false
		}
	}
	return true
}

func activityMatches(got activitySignature, inputs []string, output string) bool {
	if got.Output != output || len(got.Inputs) != len(inputs) {
		return false
	}
	for index := range inputs {
		if got.Inputs[index] != inputs[index] {
			return false
		}
	}
	return true
}
