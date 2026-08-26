package main

import (
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
	"reflect"
)

func initializeShadowEvidence(output *selectiveCIShadowOutput, inputs *shadowDecodedInputs) {
	output.BaseSourceDigest = inputs.base.Digest
	output.HeadSourceDigest = inputs.head.Digest
	output.RegistryDigest = inputs.planInput.Registry.Digest
	inputs.proof = proofsci.Evaluate(inputs.proofInput)
	output.ProofStatus = string(inputs.proof.Status)
	output.ProofCode = inputs.proof.Code
	inputs.lane = lanesci.Classify(inputs.laneInput)
	output.Lane = shadowLaneReceipt{
		Decision: string(inputs.lane.Decision), Reason: string(inputs.lane.Reason),
		RegistryDigest: inputs.lane.RegistryDigest, BaseSHA: inputs.lane.BaseSHA,
		LaneHeadSHA: inputs.lane.LaneHeadSHA, LaneID: inputs.lane.LaneID,
	}
}
func bindShadowSnapshots(inputs shadowDecodedInputs, output *selectiveCIShadowOutput) (plannersci.SnapshotManifest, plannersci.SnapshotManifest, *shadowEvaluationFailure) {
	baseManifest, err := plannerManifestFromAnalyzerSnapshot(inputs.base)
	if err != nil {
		return plannersci.SnapshotManifest{}, plannersci.SnapshotManifest{}, &shadowEvaluationFailure{"SNAPSHOT_BINDING", "base_manifest", "DERIVATION_FAILED"}
	}
	headManifest, err := plannerManifestFromAnalyzerSnapshot(inputs.head)
	if err != nil {
		return plannersci.SnapshotManifest{}, plannersci.SnapshotManifest{}, &shadowEvaluationFailure{"SNAPSHOT_BINDING", "head_manifest", "DERIVATION_FAILED"}
	}
	output.BaseSemanticDigest = baseManifest.Digest
	output.HeadSemanticDigest = headManifest.Digest
	if !reflect.DeepEqual(inputs.planInput.Base, baseManifest) {
		return plannersci.SnapshotManifest{}, plannersci.SnapshotManifest{}, &shadowEvaluationFailure{"SNAPSHOT_BINDING", "base_manifest", "MANIFEST_MISMATCH"}
	}
	if !reflect.DeepEqual(inputs.planInput.Head, headManifest) {
		return plannersci.SnapshotManifest{}, plannersci.SnapshotManifest{}, &shadowEvaluationFailure{"SNAPSHOT_BINDING", "head_manifest", "MANIFEST_MISMATCH"}
	}
	return baseManifest, headManifest, nil
}
func bindShadowRegistry(inputs shadowDecodedInputs) *shadowEvaluationFailure {
	if rawDigest(inputs.base.RegistryDigest) != inputs.planInput.Registry.Digest {
		return &shadowEvaluationFailure{"REGISTRY_BINDING", "base_snapshot", "REGISTRY_DIGEST_MISMATCH"}
	}
	if rawDigest(inputs.head.RegistryDigest) != inputs.planInput.Registry.Digest {
		return &shadowEvaluationFailure{"REGISTRY_BINDING", "head_snapshot", "REGISTRY_DIGEST_MISMATCH"}
	}
	if inputs.lane.RegistryDigest != inputs.planInput.Registry.Digest {
		return &shadowEvaluationFailure{"REGISTRY_BINDING", "lane", "REGISTRY_DIGEST_MISMATCH"}
	}
	return nil
}
func planShadowInput(inputs shadowDecodedInputs, output *selectiveCIShadowOutput) (plannersci.PlanResult, *shadowEvaluationFailure) {
	plan := plannersci.Plan(inputs.planInput)
	output.PlanDigest = plan.CanonicalDigest
	if plan.Status != plannersci.StatusSelective {
		return plannersci.PlanResult{}, &shadowEvaluationFailure{"PLAN", "planner", plan.ReasonCode}
	}
	return plan, nil
}
