package main

import (
	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
)

func decodeShadowInputs(files shadowInputFiles) (shadowDecodedInputs, *shadowEvaluationFailure) {
	base, err := analyzersci.DecodeSnapshot(files.baseSnapshot)
	if err != nil {
		return shadowDecodedInputs{}, &shadowEvaluationFailure{"INPUT", "base_snapshot", shadowDecodeReason(err)}
	}
	head, err := analyzersci.DecodeSnapshot(files.headSnapshot)
	if err != nil {
		return shadowDecodedInputs{}, &shadowEvaluationFailure{"INPUT", "head_snapshot", shadowDecodeReason(err)}
	}
	planInput, err := plannersci.DecodeJSON(files.planInput)
	if err != nil {
		return shadowDecodedInputs{}, &shadowEvaluationFailure{"INPUT", "plan_input", shadowDecodeReason(err)}
	}
	proofInput, err := proofsci.DecodeInput(files.evidenceInput)
	if err != nil {
		return shadowDecodedInputs{}, &shadowEvaluationFailure{"INPUT", "evidence_input", shadowDecodeReason(err)}
	}
	laneInput, err := lanesci.DecodeJSON(files.laneInput)
	if err != nil {
		return shadowDecodedInputs{}, &shadowEvaluationFailure{"INPUT", "lane_input", shadowDecodeReason(err)}
	}
	return shadowDecodedInputs{base: base, head: head, planInput: planInput, proofInput: proofInput, laneInput: laneInput}, nil
}
