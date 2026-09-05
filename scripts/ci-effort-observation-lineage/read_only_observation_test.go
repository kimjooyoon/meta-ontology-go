package main

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicworkflowlineage"
)

func testReadOnlyPolicy() publicworkflowlineage.Policy {
	return publicworkflowlineage.Policy{
		SourceWorkflow:          ".github/workflows/ci.yml",
		ConsumerWorkflow:        "CI effort observation",
		Repository:              "kimjooyoon/meta-ontology-go",
		SourceAPIKey:            "workflow_run.id",
		ArtifactSubjectBinding:  "ci-evidence.head_sha",
		ReadOnlyPermissions:     publicworkflowlineage.ReadOnlyPermissions{
			WorkflowWindow:      publicworkflowlineage.ReadOnlyPermission,
			VerificationRuntime: publicworkflowlineage.ReadOnlyPermission,
			EvidenceReuse:       publicworkflowlineage.ExactSuccessReuse,
			Promotion:           publicworkflowlineage.NoPromotionPermission,
		},
	}
}

func testReadOnlyInput(policy publicworkflowlineage.Policy) publicworkflowlineage.Input {
	return fixture(policy, publicworkflowlineage.CaseSpec{
		ID:             "EXACT_TRIGGER_ARTIFACT_MATCH",
		SourceSubject:  "446c8451b231ed08087945ac1a7f705bea7225be",
		SourceRunID:    33859268902,
		SourceRefState: publicworkflowlineage.RefStateValue,
	}, 0)
}

func TestReadOnlyObservationAllowsExactSourceFailureWithoutAuthority(t *testing.T) {
	policy := testReadOnlyPolicy()
	input := testReadOnlyInput(policy)
	input.Source.Conclusion = "failure"

	strict := publicworkflowlineage.Evaluate(input)
	projection := policy.EvaluateReadOnlyObservation(input)
	if strict.Decision != publicworkflowlineage.DecisionRefuted || strict.LineageState != publicworkflowlineage.StateMismatch || strict.MismatchDetected || !strict.ProductFailureKept {
		t.Fatalf("strict source failure semantics changed: %+v", strict)
	}
	if projection.Schema != publicworkflowlineage.ObservationSchema || projection.Eligibility != publicworkflowlineage.ObservationAllowed || projection.Decision != publicworkflowlineage.DecisionRefuted || projection.LineageState != publicworkflowlineage.StateMismatch || !projection.ExactSourceIdentity || !projection.TimingObservationEligible || !projection.OperationObservationEligible || !projection.SourceFailureKept || projection.EvidenceReuseAllowed || projection.PromotionAllowed {
		t.Fatalf("source failure observation projection is unsafe or unavailable: %+v", projection)
	}
}

func TestReadOnlyObservationRejectsCurrentHeadSubstitution(t *testing.T) {
	policy := testReadOnlyPolicy()
	input := testReadOnlyInput(policy)
	input.Source.Conclusion = "failure"
	input.Trigger.CandidateSubjectSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	strict := publicworkflowlineage.Evaluate(input)
	projection := policy.EvaluateReadOnlyObservation(input)
	if strict.Decision != publicworkflowlineage.DecisionRefuted || strict.LineageState != publicworkflowlineage.StateCurrentDevFallback || !strict.FallbackAttempted || !strict.FallbackRejected || strict.ProductFailureKept {
		t.Fatalf("current-head substitution was not rejected: %+v", strict)
	}
	if projection.Eligibility != publicworkflowlineage.ObservationDenied || projection.TimingObservationEligible || projection.OperationObservationEligible || projection.EvidenceReuseAllowed || projection.PromotionAllowed {
		t.Fatalf("mismatched identity received observation authority: %+v", projection)
	}
}

func TestReadOnlyObservationKeepsMissingIdentityUnknown(t *testing.T) {
	policy := testReadOnlyPolicy()
	input := testReadOnlyInput(policy)
	input.Source = publicworkflowlineage.SourceRun{}

	projection := policy.EvaluateReadOnlyObservation(input)
	if projection.Decision != publicworkflowlineage.DecisionUnknown || projection.LineageState != publicworkflowlineage.StateDirectMissing || projection.Eligibility != publicworkflowlineage.ObservationDenied || projection.Unknown == nil || projection.TimingObservationEligible || projection.OperationObservationEligible {
		t.Fatalf("missing source identity was treated as observable: %+v", projection)
	}
}
