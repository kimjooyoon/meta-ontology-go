package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicworkflowlineage"
)

func testReadOnlySource(t *testing.T) (string, []byte) {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate read-only observation test source")
	}
	path := filepath.Join(filepath.Dir(filename), "../../examples/ci-effort-observation/main.gooo")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return path, source
}

func testReadOnlyPolicy(t *testing.T) publicworkflowlineage.Policy {
	t.Helper()
	path, source := testReadOnlySource(t)
	policy, err := publicworkflowlineage.Load(path, source)
	if err != nil {
		t.Fatal(err)
	}
	return policy
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
	policy := testReadOnlyPolicy(t)
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
	policy := testReadOnlyPolicy(t)
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
	policy := testReadOnlyPolicy(t)
	input := testReadOnlyInput(policy)
	input.Source = publicworkflowlineage.SourceRun{}

	projection := policy.EvaluateReadOnlyObservation(input)
	if projection.Decision != publicworkflowlineage.DecisionUnknown || projection.LineageState != publicworkflowlineage.StateDirectMissing || projection.Eligibility != publicworkflowlineage.ObservationDenied || projection.Unknown == nil || projection.TimingObservationEligible || projection.OperationObservationEligible {
		t.Fatalf("missing source identity was treated as observable: %+v", projection)
	}
}

func TestLoadBindsReadOnlyPermissionsToSemanticActivities(t *testing.T) {
	policy := testReadOnlyPolicy(t)
	if policy.ReadOnlyPermissions.WorkflowWindow != publicworkflowlineage.ReadOnlyPermission || policy.ReadOnlyPermissions.VerificationRuntime != publicworkflowlineage.ReadOnlyPermission || policy.ReadOnlyPermissions.EvidenceReuse != publicworkflowlineage.ExactSuccessReuse || policy.ReadOnlyPermissions.Promotion != publicworkflowlineage.NoPromotionPermission {
		t.Fatalf("read-only permissions were not loaded from the canonical activities: %+v", policy.ReadOnlyPermissions)
	}
}

func TestLoadRejectsReadOnlyPermissionMarkerDrift(t *testing.T) {
	path, source := testReadOnlySource(t)
	cases := []struct {
		name string
		from string
		to   string
	}{
		{"window-permission-missing", "partial-lineage-observation-permission=READ_ONLY;", ""},
		{"runtime-permission-changed", "observation-permission=READ_ONLY", "observation-permission=WRITE"},
		{"reuse-permission-duplicated", "evidence-reuse-permission=EXACT_SUCCESS_ONLY;", "evidence-reuse-permission=EXACT_SUCCESS_ONLY;evidence-reuse-permission=WRITE;"},
		{"promotion-permission-contradictory", "promotion-permission=NONE", "promotion-permission=NONE;promotion-permission=WRITE"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			mutated := strings.Replace(string(source), test.from, test.to, 1)
			if mutated == string(source) {
				t.Fatal("test marker was not found")
			}
			if _, err := publicworkflowlineage.Load(path, []byte(mutated)); err == nil {
				t.Fatal("permission marker drift was accepted")
			}
		})
	}
}

func TestReadOnlyObservationRejectsIncompleteManualPolicy(t *testing.T) {
	policy := testReadOnlyPolicy(t)
	policy.ReadOnlyPermissions = publicworkflowlineage.ReadOnlyPermissions{}
	input := testReadOnlyInput(policy)
	input.Source.Conclusion = "failure"

	projection := policy.EvaluateReadOnlyObservation(input)
	if projection.Eligibility != publicworkflowlineage.ObservationDenied || projection.TimingObservationEligible || projection.OperationObservationEligible || projection.EvidenceReuseAllowed || projection.PromotionAllowed {
		t.Fatalf("incomplete manual policy received observation authority: %+v", projection)
	}
}
