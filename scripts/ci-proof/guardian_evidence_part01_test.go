package main

import (
	"strings"
	"testing"
)

func validGuardianEvidenceFixture() (guardianEvidence, proofBundle) {
	base := strings.Repeat("b", 40)
	head := strings.Repeat("a", 40)
	bundle := proofBundle{Repository: guardianInstallationRepository, PRNumber: 7, Event: "pull_request", BaseRef: "main", BaseSHA: base, HeadSHA: head, RunID: 300, RunAttempt: 1, Digests: proofDigests{Policy: strings.Repeat("6", 64)}, PromotionObservation: &promotionObservation{Action: "synchronize", Topology: guardianTopology{Status: "ahead", AheadBy: 1, BehindBy: 0, MergeBaseSHA: base}}}
	observedAt, validUntil := freshObserverWindow()
	evidence := guardianEvidence{
		Schema: guardianEvidenceSchema, Route: "promotion_main", CheckName: "CI guardian", Repository: guardianInstallationRepository, PRNumber: 7, Action: "synchronize",
		BaseRepo: guardianInstallationRepository, BaseRef: "main", BaseSHA: base, HeadRepo: guardianInstallationRepository, HeadRef: "dev", HeadSHA: head,
		WorkflowRef: guardianInstallationRepository + "/.github/workflows/ci-guardian.yml@refs/heads/dev", WorkflowSHA: head, RuntimeRef: "refs/heads/dev", RuntimeSHA: head,
		EventRef: "refs/heads/dev", DefaultBranch: "dev", ObserverEnvironmentName: "guardian-observer", RunID: 200, RunAttempt: 1, WorkflowID: 12, WorkflowPath: ".github/workflows/ci-guardian.yml",
		RunEvent: "pull_request_target", RunStatus: "completed", RunConclusion: "success", RunCreatedAt: "2026-08-14T00:00:00Z", RunNumber: 12,
		LiveRefsBefore: guardianLiveRefs{DevSHA: head, MainSHA: base}, LiveRefsAfter: guardianLiveRefs{DevSHA: head, MainSHA: base},
		Topology: guardianTopology{Status: "ahead", AheadBy: 1, BehindBy: 0, MergeBaseSHA: base}, ArtifactID: 44, ArtifactName: "ci-guardian-200-1", ArtifactSize: 100,
		ArtifactDigest: "sha256:" + strings.Repeat("1", 64), ManifestBundleSHA: "sha256:" + strings.Repeat("2", 64), GuardianJobID: 55, GuardianJobName: "CI guardian", GuardianJobStatus: "completed", GuardianJobConclusion: "success", GuardianJobHeadSHA: head,
		CheckRunID: 55, CheckRunName: "CI guardian", CheckRunAppID: 15368, CheckRunStatus: "completed", CheckRunConclusion: "success", CheckRunHeadSHA: head, CheckSuiteID: 77,
		Decision: "PASS", HeadBindingStatus: "verified", BundleSHA: "sha256:" + strings.Repeat("2", 64),
	}
	evidence.BranchProtection = branchProtection{Repository: guardianInstallationRepository, Branch: "main", PolicySHA: strings.Repeat("6", 64), EventRef: "refs/heads/dev", CheckoutRef: head, TokenSource: "github_app_installation", AppInstallationID: 42, AppSlug: "guardian", ReadStatus: "verified", Exists: true, Strict: true, RequiredChecks: append(append([]string(nil), proofJobs...), "CI guardian"), RequiredCheckBindings: requiredCheckBindingsFor(append(append([]string(nil), proofJobs...), "CI guardian")), EnforceAdmins: true, RequiredReviews: 0, LinearHistory: true, BaseSHA: base, HeadSHA: head, RunID: evidence.RunID, RunAttempt: evidence.RunAttempt, WorkflowSHA: head, ObservedAt: observedAt, ValidUntil: validUntil}
	evidence.BranchProtection.Digest = digestBranchProtection(evidence.BranchProtection)
	evidence.DevBranchProtection = branchProtection{Repository: guardianInstallationRepository, Branch: "dev", PolicySHA: strings.Repeat("6", 64), EventRef: "refs/heads/dev", CheckoutRef: head, TokenSource: "github_app_installation", AppInstallationID: 42, AppSlug: "guardian", ReadStatus: "verified", Exists: true, Strict: true, RequiredChecks: append(append([]string(nil), proofJobs...), "CI guardian shadow"), RequiredCheckBindings: requiredCheckBindingsFor(append(append([]string(nil), proofJobs...), "CI guardian shadow")), EnforceAdmins: true, RequiredReviews: 0, LinearHistory: true, BaseSHA: base, HeadSHA: head, RunID: evidence.RunID, RunAttempt: evidence.RunAttempt, WorkflowSHA: head, ObservedAt: observedAt, ValidUntil: validUntil}
	evidence.DevBranchProtection.Digest = digestBranchProtection(evidence.DevBranchProtection)
	evidence.ObserverEnvironmentSnapshot = guardianEnvironment{Repository: guardianInstallationRepository, Name: "guardian-observer", DeploymentBranchPolicy: guardianDeploymentBranchPolicy{ProtectedBranches: true, CustomBranchPolicies: false}, ProtectionRules: []string{"branch_policy"}, TokenSource: "github.token", ReadStatus: "verified", WaitTimer: 0, Reviewers: []string{}, RunID: evidence.RunID, RunAttempt: evidence.RunAttempt, WorkflowSHA: head, ObservedAt: observedAt, ValidUntil: validUntil}
	evidence.ObserverEnvironmentSnapshot.Digest = digestGuardianEnvironment(evidence.ObserverEnvironmentSnapshot)
	evidence.ObserverEnvironmentDigest = evidence.ObserverEnvironmentSnapshot.Digest
	evidence.InstallationRepositoryScope = guardianInstallationScope{Repository: guardianInstallationRepository, InstallationID: 42, TokenSource: "github_app_installation", ReadStatus: "verified", RepositoryCount: 1, Repositories: []string{guardianInstallationRepository}, ExactMatch: true, RunID: evidence.RunID, RunAttempt: evidence.RunAttempt, WorkflowSHA: head, ObservedAt: observedAt, ValidUntil: validUntil}
	evidence.InstallationRepositoryScope.Digest = digestGuardianInstallationScope(evidence.InstallationRepositoryScope)
	bundle.BranchProtection = evidence.BranchProtection
	bundle.DevBranchProtection = evidence.DevBranchProtection
	return evidence, bundle
}
func TestGuardianEvidenceRejectsObserverProtectionRuleMutations(t *testing.T) {
	for name, rules := range map[string][]string{
		"empty":     {},
		"duplicate": {"branch_policy", "branch_policy"},
		"wrong":     {"required_reviewers"},
	} {
		t.Run(name, func(t *testing.T) {
			evidence, bundle := validGuardianEvidenceFixture()
			evidence.ObserverEnvironmentSnapshot.ProtectionRules = rules
			evidence.ObserverEnvironmentSnapshot.Digest = digestGuardianEnvironment(evidence.ObserverEnvironmentSnapshot)
			evidence.ObserverEnvironmentDigest = evidence.ObserverEnvironmentSnapshot.Digest
			if err := validateGuardianEvidence(&evidence, bundle); err == nil {
				t.Fatalf("observer protection rule mutation %q was accepted", name)
			}
		})
	}
}
