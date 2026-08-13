package main

import (
	"strings"
	"testing"
)

func validGuardianEvidenceFixture() (guardianEvidence, proofBundle) {
	base := strings.Repeat("b", 40)
	head := strings.Repeat("a", 40)
	bundle := proofBundle{Repository: "owner/repo", PRNumber: 7, BaseRef: "main", BaseSHA: base, HeadSHA: head, RunID: 300, RunAttempt: 1, Digests: proofDigests{Policy: strings.Repeat("6", 64)}, PromotionObservation: &promotionObservation{Action: "synchronize", Topology: guardianTopology{Status: "ahead", AheadBy: 1, BehindBy: 0, MergeBaseSHA: base}}}
	evidence := guardianEvidence{
		Schema: guardianEvidenceSchema, Route: "promotion_main", CheckName: "CI guardian", Repository: "owner/repo", PRNumber: 7, Action: "synchronize",
		BaseRepo: "owner/repo", BaseRef: "main", BaseSHA: base, HeadRepo: "owner/repo", HeadRef: "dev", HeadSHA: head,
		WorkflowRef: "owner/repo/.github/workflows/ci-guardian.yml@refs/heads/dev", WorkflowSHA: head, RuntimeRef: "refs/heads/dev", RuntimeSHA: head,
		EventRef: "refs/heads/dev", DefaultBranch: "dev", ObserverEnvironmentName: "guardian-observer", RunID: 200, RunAttempt: 1, WorkflowID: 12, WorkflowPath: ".github/workflows/ci-guardian.yml",
		RunEvent: "pull_request_target", RunStatus: "completed", RunConclusion: "success", RunCreatedAt: "2026-08-14T00:00:00Z", RunNumber: 12,
		LiveRefsBefore: guardianLiveRefs{DevSHA: head, MainSHA: base}, LiveRefsAfter: guardianLiveRefs{DevSHA: head, MainSHA: base},
		Topology: guardianTopology{Status: "ahead", AheadBy: 1, BehindBy: 0, MergeBaseSHA: base}, ArtifactID: 44, ArtifactName: "ci-guardian-200-1", ArtifactSize: 100,
		ArtifactDigest: "sha256:" + strings.Repeat("1", 64), ManifestBundleSHA: "sha256:" + strings.Repeat("2", 64), GuardianJobID: 55, GuardianJobName: "CI guardian", GuardianJobStatus: "completed", GuardianJobConclusion: "success", GuardianJobHeadSHA: head,
		CheckRunID: 55, CheckRunName: "CI guardian", CheckRunAppID: 15368, CheckRunStatus: "completed", CheckRunConclusion: "success", CheckRunHeadSHA: head, CheckSuiteID: 77,
		Decision: "PASS", HeadBindingStatus: "verified", BundleSHA: "sha256:" + strings.Repeat("2", 64),
	}
	evidence.BranchProtection = branchProtection{Repository: "owner/repo", Branch: "main", PolicySHA: strings.Repeat("6", 64), EventRef: "refs/heads/dev", CheckoutRef: head, TokenSource: "github_app_installation", AppInstallationID: 42, AppSlug: "guardian", ReadStatus: "verified", Exists: true, Strict: true, RequiredChecks: append(append([]string(nil), proofJobs...), "CI guardian"), RequiredCheckBindings: requiredCheckBindingsFor(append(append([]string(nil), proofJobs...), "CI guardian")), EnforceAdmins: true, RequiredReviews: 0, LinearHistory: true, BaseSHA: base, HeadSHA: head, RunID: evidence.RunID, RunAttempt: evidence.RunAttempt, WorkflowSHA: head}
	evidence.BranchProtection.Digest = digestBranchProtection(evidence.BranchProtection)
	evidence.DevBranchProtection = branchProtection{Repository: "owner/repo", Branch: "dev", PolicySHA: strings.Repeat("6", 64), EventRef: "refs/heads/dev", CheckoutRef: head, TokenSource: "github_app_installation", AppInstallationID: 42, AppSlug: "guardian", ReadStatus: "verified", Exists: true, Strict: true, RequiredChecks: append(append([]string(nil), proofJobs...), "CI guardian shadow"), RequiredCheckBindings: requiredCheckBindingsFor(append(append([]string(nil), proofJobs...), "CI guardian shadow")), EnforceAdmins: true, RequiredReviews: 0, LinearHistory: true, BaseSHA: base, HeadSHA: head, RunID: evidence.RunID, RunAttempt: evidence.RunAttempt, WorkflowSHA: head}
	evidence.DevBranchProtection.Digest = digestBranchProtection(evidence.DevBranchProtection)
	evidence.ObserverEnvironmentSnapshot = guardianEnvironment{Repository: "owner/repo", Name: "guardian-observer", DeploymentBranchPolicy: guardianDeploymentBranchPolicy{ProtectedBranches: true, CustomBranchPolicies: false}, ProtectionRules: []string{"branch_policy"}, TokenSource: "github.token", ReadStatus: "verified", WaitTimer: 0, Reviewers: []string{}, RunID: evidence.RunID, RunAttempt: evidence.RunAttempt, WorkflowSHA: head}
	evidence.ObserverEnvironmentSnapshot.Digest = digestGuardianEnvironment(evidence.ObserverEnvironmentSnapshot)
	evidence.ObserverEnvironmentDigest = evidence.ObserverEnvironmentSnapshot.Digest
	bundle.BranchProtection = evidence.BranchProtection
	bundle.DevBranchProtection = evidence.DevBranchProtection
	return evidence, bundle
}

func TestGuardianEvidenceKeepsIndependentRunTuple(t *testing.T) {
	evidence, bundle := validGuardianEvidenceFixture()
	if evidence.RunID == bundle.RunID {
		t.Fatal("fixture must prove Guardian and proof runs are independent")
	}
	if err := validateGuardianEvidence(&evidence, bundle); err != nil {
		t.Fatalf("valid independent Guardian evidence rejected: %v", err)
	}
}

func TestGuardianEvidenceRejectsCrossBindingTamper(t *testing.T) {
	mutations := []func(*guardianEvidence){
		func(e *guardianEvidence) { e.PRNumber++ },
		func(e *guardianEvidence) { e.HeadSHA = strings.Repeat("c", 40) },
		func(e *guardianEvidence) { e.ArtifactName = "ci-guardian-999-1" },
		func(e *guardianEvidence) { e.CheckRunAppID = 1 },
		func(e *guardianEvidence) { e.CheckSuiteID = 0 },
		func(e *guardianEvidence) { e.GuardianJobConclusion = "failure" },
		func(e *guardianEvidence) { e.HeadBindingStatus = "CI-GUARDIAN-HEAD-BINDING-UNVERIFIED" },
		func(e *guardianEvidence) { e.Schema = "gooo/ci-guardian/v2" },
		func(e *guardianEvidence) { e.Action = "closed" },
		func(e *guardianEvidence) { e.RunCreatedAt = "not-a-timestamp" },
		func(e *guardianEvidence) { e.CheckRunID = 66 },
	}
	for index, mutate := range mutations {
		evidence, bundle := validGuardianEvidenceFixture()
		mutate(&evidence)
		if err := validateGuardianEvidence(&evidence, bundle); err == nil {
			t.Fatalf("tampered Guardian evidence mutation %d was accepted", index)
		}
	}
}

func TestGuardianEvidenceRejectsUnavailableOrNonPromotionTopology(t *testing.T) {
	evidence, bundle := validGuardianEvidenceFixture()
	evidence.Topology.Status = "identical"
	if err := validateGuardianEvidence(&evidence, bundle); err == nil {
		t.Fatal("identical main/dev topology was accepted for promotion")
	}
}
