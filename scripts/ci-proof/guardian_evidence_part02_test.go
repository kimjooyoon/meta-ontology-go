package main

import (
	"strings"
	"testing"
)

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
