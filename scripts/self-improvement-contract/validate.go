package main

type stageContract struct {
	ID     string
	Inputs []string
	Output string
}

var loopStages = []stageContract{
	{ID: "meta://activity/measure", Inputs: []string{"gooo://meta/source-tree"}, Output: "gooo://meta/source-metrics"},
	{ID: "meta://activity/plan", Inputs: []string{"gooo://meta/source-metrics", "gooo://meta/proof-choice"}, Output: "gooo://meta/generation-plan"},
	{ID: "meta://activity/evaluate", Inputs: []string{"gooo://meta/execution-manifest"}, Output: "gooo://meta/receipt"},
	{ID: "meta://activity/verify", Inputs: []string{"gooo://meta/receipt"}, Output: "gooo://meta/evidence"},
	{ID: "meta://activity/improve", Inputs: []string{"gooo://meta/evidence"}, Output: "gooo://meta/source-tree"},
}

const executeActivityID = "meta://activity/execute"

func closedLoop(model contractModel) bool {
	for _, stage := range loopStages {
		if !activityMatches(model.Activities[stage.ID], stage.Inputs, stage.Output) {
			return false
		}
	}
	execute, exists := model.Activities[executeActivityID]
	return exists && len(execute.Inputs) > 0 &&
		execute.Inputs[0] == "gooo://meta/generation-plan" &&
		execute.Output == "gooo://meta/execution-manifest"
}

func trilemmaChoice(model contractModel, registry []RegistryBinding) bool {
	choice := model.Activities["meta://activity/choose-proof"]
	if !activityMatches(choice,
		[]string{"proof://foundation", "proof://coherence", "proof://regression"}, "gooo://meta/proof-choice") {
		return false
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
