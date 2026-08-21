package main

import (
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
)

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
