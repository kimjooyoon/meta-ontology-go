package main

import (
	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
)

type shadowDecodedInputs struct {
	base       analyzersci.Snapshot
	head       analyzersci.Snapshot
	planInput  plannersci.Input
	proofInput proofsci.Input
	laneInput  lanesci.Input
	proof      proofsci.Receipt
	lane       lanesci.Output
}
type shadowEvaluationFailure struct {
	stage     string
	component string
	reason    string
}

func evaluateSelectiveCIShadow(files shadowInputFiles) selectiveCIShadowOutput {
	output := newSelectiveCIShadowOutput()
	inputs, failure := decodeShadowInputs(files)
	if failure != nil {
		return shadowFallback(output, failure.stage, failure.component, failure.reason)
	}
	initializeShadowEvidence(&output, &inputs)
	baseManifest, headManifest, failure := bindShadowSnapshots(inputs, &output)
	if failure != nil {
		return shadowFallback(output, failure.stage, failure.component, failure.reason)
	}
	if failure = bindShadowRegistry(inputs); failure != nil {
		return shadowFallback(output, failure.stage, failure.component, failure.reason)
	}
	plan, failure := planShadowInput(inputs, &output)
	if failure != nil {
		return shadowFallback(output, failure.stage, failure.component, failure.reason)
	}
	failure = bindShadowProof(inputs, baseManifest, headManifest, plan)
	if failure != nil {
		return shadowFallback(output, failure.stage, failure.component, failure.reason)
	}
	if failure = gateShadowProof(inputs.proof); failure != nil {
		return shadowFallback(output, failure.stage, failure.component, failure.reason)
	}
	if failure = gateShadowLane(inputs.lane); failure != nil {
		return shadowFallback(output, failure.stage, failure.component, failure.reason)
	}
	return finishShadowSelection(output, plan, inputs.planInput.Registry)
}
