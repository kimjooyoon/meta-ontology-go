package main

import (
	"strings"
	"testing"
	"time"
)

func validGuardianEvidenceFixture() (guardianEvidence, proofBundle) {
	base := strings.Repeat("b", 40)
	head := strings.Repeat("a", 40)
	bundle := proofBundle{Repository: guardianInstallationRepository, PRNumber: 7, BaseRef: "main", BaseSHA: base, HeadSHA: head, RunID: 300, RunAttempt: 1, Digests: proofDigests{Policy: strings.Repeat("6", 64)}, PromotionObservation: &promotionObservation{Action: "synchronize", Topology: guardianTopology{Status: "ahead", AheadBy: 1, BehindBy: 0, MergeBaseSHA: base}}}
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

func TestGuardianEvidenceRejectsInstallationScopeMutations(t *testing.T) {
	mutations := map[string]func(*guardianEvidence){
		"wrong repository": func(e *guardianEvidence) {
			e.InstallationRepositoryScope.Repository = "owner/repo"
			e.InstallationRepositoryScope.Repositories = []string{"owner/repo"}
			e.InstallationRepositoryScope.Digest = digestGuardianInstallationScope(e.InstallationRepositoryScope)
		},
		"wrong repository list": func(e *guardianEvidence) {
			e.InstallationRepositoryScope.Repositories = []string{"kimjooyoon/other"}
			e.InstallationRepositoryScope.Digest = digestGuardianInstallationScope(e.InstallationRepositoryScope)
		},
		"wrong token source": func(e *guardianEvidence) {
			e.InstallationRepositoryScope.TokenSource = "github.token"
			e.InstallationRepositoryScope.Digest = digestGuardianInstallationScope(e.InstallationRepositoryScope)
		},
		"wrong run binding": func(e *guardianEvidence) {
			e.InstallationRepositoryScope.RunID++
			e.InstallationRepositoryScope.Digest = digestGuardianInstallationScope(e.InstallationRepositoryScope)
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			evidence, bundle := validGuardianEvidenceFixture()
			mutate(&evidence)
			if err := validateGuardianEvidence(&evidence, bundle); err == nil {
				t.Fatalf("installation scope mutation %q was accepted", name)
			}
		})
	}
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

func TestGuardianEvidenceRejectsMissingFutureAndExpiredObserverFreshness(t *testing.T) {
	now := time.Date(2026, time.August, 14, 0, 0, 0, 0, time.UTC)
	setFreshness := func(evidence *guardianEvidence, observedAt, validUntil *string) {
		evidence.BranchProtection.ObservedAt = observedAt
		evidence.BranchProtection.ValidUntil = validUntil
		evidence.BranchProtection.Digest = digestBranchProtection(evidence.BranchProtection)
		evidence.DevBranchProtection.ObservedAt = observedAt
		evidence.DevBranchProtection.ValidUntil = validUntil
		evidence.DevBranchProtection.Digest = digestBranchProtection(evidence.DevBranchProtection)
		evidence.ObserverEnvironmentSnapshot.ObservedAt = observedAt
		evidence.ObserverEnvironmentSnapshot.ValidUntil = validUntil
		evidence.ObserverEnvironmentSnapshot.Digest = digestGuardianEnvironment(evidence.ObserverEnvironmentSnapshot)
		evidence.ObserverEnvironmentDigest = evidence.ObserverEnvironmentSnapshot.Digest
		evidence.InstallationRepositoryScope.ObservedAt = observedAt
		evidence.InstallationRepositoryScope.ValidUntil = validUntil
		evidence.InstallationRepositoryScope.Digest = digestGuardianInstallationScope(evidence.InstallationRepositoryScope)
	}
	for name, window := range map[string][2]*string{
		"missing": {nil, stringPointer("2026-08-14T00:10:00Z")},
		"future":  {stringPointer("2026-08-14T00:01:00Z"), stringPointer("2026-08-14T00:11:00Z")},
		"expired": {stringPointer("2026-08-13T23:40:00Z"), stringPointer("2026-08-13T23:50:00Z")},
	} {
		t.Run(name, func(t *testing.T) {
			evidence, bundle := validGuardianEvidenceFixture()
			setFreshness(&evidence, window[0], window[1])
			bundle.BranchProtection = evidence.BranchProtection
			bundle.DevBranchProtection = evidence.DevBranchProtection
			if err := validateGuardianEvidenceAt(&evidence, bundle, now); err == nil {
				t.Fatalf("%s observer freshness was accepted", name)
			}
		})
	}
	evidence, bundle := validGuardianEvidenceFixture()
	setFreshness(&evidence, stringPointer("2026-08-13T23:55:00Z"), stringPointer("2026-08-14T00:05:00Z"))
	bundle.BranchProtection = evidence.BranchProtection
	bundle.DevBranchProtection = evidence.DevBranchProtection
	if err := validateGuardianEvidenceAt(&evidence, bundle, now); err != nil {
		t.Fatalf("valid observer freshness was rejected: %v", err)
	}
}
