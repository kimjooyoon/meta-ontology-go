package main

import (
	"errors"
	"reflect"
	"strings"

	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func bindShadowProof(inputs shadowDecodedInputs, baseManifest, headManifest plannersci.SnapshotManifest, plan plannersci.PlanResult) *shadowEvaluationFailure {
	expectedSnapshots := proofsci.SnapshotBinding{
		Base: semantic.SnapshotDigests{Source: rawDigest(inputs.base.Digest), Semantic: baseManifest.Digest},
		Head: semantic.SnapshotDigests{Source: rawDigest(inputs.head.Digest), Semantic: headManifest.Digest},
	}
	expectedSelected := sortedUnion(plan.SelectedCommandIDs, plan.SelectedGuardCommandIDs)
	checks := []struct {
		component string
		reason    string
		failed    bool
	}{
		{"registry_digest", "PROOF_REGISTRY_DIGEST_MISMATCH", inputs.proofInput.RegistryDigest != inputs.planInput.Registry.Digest},
		{"plan_digest", "PROOF_PLAN_DIGEST_MISMATCH", inputs.proofInput.PlanDigest != plan.CanonicalDigest},
		{"changed_root_ids", "PROOF_CHANGED_ROOT_IDS_MISMATCH", !reflect.DeepEqual(sortedSemanticIDs(inputs.proofInput.ChangedRootIDs), plan.ChangedSemanticIDs)},
		{"selected_command_ids", "PROOF_SELECTED_COMMAND_IDS_MISMATCH", !reflect.DeepEqual(sortedSemanticIDs(inputs.proofInput.SelectedCommandIDs), expectedSelected)},
		{"snapshots", "PROOF_SNAPSHOT_BINDING_MISMATCH", inputs.proofInput.Snapshots != expectedSnapshots},
	}
	for _, check := range checks {
		if check.failed {
			return &shadowEvaluationFailure{"PLAN_PROOF_BINDING", check.component, check.reason}
		}
	}
	return nil
}

func gateShadowProof(proof proofsci.Receipt) *shadowEvaluationFailure {
	switch proof.Status {
	case proofsci.FailClosed:
		return &shadowEvaluationFailure{"PROOF_FAIL_CLOSED", "proof", proof.Code}
	case proofsci.Unknown:
		return &shadowEvaluationFailure{"PROOF_UNKNOWN", "proof", proof.Code}
	case proofsci.Verified:
		return nil
	default:
		return &shadowEvaluationFailure{"PROOF_FAIL_CLOSED", "proof", "INVALID_PROOF_STATUS"}
	}
}

func gateShadowLane(lane lanesci.Output) *shadowEvaluationFailure {
	switch lane.Decision {
	case lanesci.DecisionUnknown:
		return &shadowEvaluationFailure{"LANE_UNKNOWN", "lane", string(lane.Reason)}
	case lanesci.DecisionIneligible:
		return &shadowEvaluationFailure{"LANE_INELIGIBLE", "lane", string(lane.Reason)}
	case lanesci.DecisionEligible:
		return nil
	default:
		return &shadowEvaluationFailure{"LANE_UNKNOWN", "lane", "INVALID_LANE_DECISION"}
	}
}

func finishShadowSelection(output selectiveCIShadowOutput, plan plannersci.PlanResult, registry plannersci.Registry) selectiveCIShadowOutput {
	commands, guards, receipts, err := selectedShadowCommands(plan, registry)
	if err != nil {
		return shadowFallback(output, "PLAN", "selected_commands", "MISSING_COMMAND_SPEC")
	}
	output.Status = "SHADOW_SELECTIVE"
	output.Stage = "SELECTIVE"
	output.Component = "all"
	output.Reason = "VERIFIED"
	output.ChangedSemanticIDs = append([]string{}, plan.ChangedSemanticIDs...)
	output.SelectedCommands = commands
	output.SelectedGuards = guards
	output.SelectedWorkIDs = append([]string{}, plan.SelectedWorkIDs...)
	output.ResourceReceipts = receipts
	return sealSelectiveCIShadowOutput(output)
}

func shadowDecodeReason(err error) string {
	var snapshotErr *analyzersci.Error
	if errors.As(err, &snapshotErr) {
		return string(snapshotErr.Code)
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "duplicate"):
		return "DUPLICATE_FIELD"
	case strings.Contains(message, "unknown field") || strings.Contains(message, "unknown"):
		return "UNKNOWN_FIELD"
	case strings.Contains(message, "trailing") || strings.Contains(message, "multiple"):
		return "TRAILING_DATA"
	case strings.Contains(message, "stale") || strings.Contains(message, "mismatch"):
		return "STALE_OR_MISMATCHED"
	default:
		return "MALFORMED"
	}
}
