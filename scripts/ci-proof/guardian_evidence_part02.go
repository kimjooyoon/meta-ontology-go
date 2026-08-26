package main

import (
	"encoding/json"
	"fmt"
	"time"
)

func digestGuardianEnvironment(environment guardianEnvironment) string {
	environment.Digest = ""
	data, _ := json.Marshal(environment)
	return digestBytes(data)
}
func digestGuardianInstallationScope(scope guardianInstallationScope) string {
	scope.Digest = ""
	data, _ := json.Marshal(scope)
	return digestBytes(data)
}
func validateGuardianInstallationScopeAt(scope guardianInstallationScope, evidence *guardianEvidence, bundle proofBundle, now time.Time) error {
	if scope.Repository != bundle.Repository || bundle.Repository != guardianInstallationRepository || scope.InstallationID <= 0 || scope.TokenSource != "github_app_installation" || scope.ReadStatus != "verified" || scope.RepositoryCount != 1 || len(scope.Repositories) != 1 || scope.Repositories[0] != guardianInstallationRepository || !scope.ExactMatch || scope.MissingReason != "" || scope.RunID != evidence.RunID || scope.RunAttempt != evidence.RunAttempt || scope.WorkflowSHA != evidence.WorkflowSHA || scope.Digest == "" || scope.Digest != digestGuardianInstallationScope(scope) || !validObserverFreshness(scope.ObservedAt, scope.ValidUntil, now) {
		return fmt.Errorf("guardian installation repository scope evidence is missing, tampered, or unbound")
	}
	return nil
}
func validateGuardianEnvironmentEvidence(environment guardianEnvironment, evidence *guardianEvidence, bundle proofBundle) error {
	return validateGuardianEnvironmentEvidenceAt(environment, evidence, bundle, time.Now().UTC())
}
func validateGuardianEnvironmentEvidenceAt(environment guardianEnvironment, evidence *guardianEvidence, bundle proofBundle, now time.Time) error {
	if environment.Repository != bundle.Repository || environment.Name != "guardian-observer" || environment.TokenSource != "github.token" || environment.ReadStatus != "verified" || !environment.DeploymentBranchPolicy.ProtectedBranches || environment.DeploymentBranchPolicy.CustomBranchPolicies || len(environment.ProtectionRules) != 1 || environment.ProtectionRules[0] != "branch_policy" || environment.WaitTimer != 0 || len(environment.Reviewers) != 0 || environment.MissingReason != "" || environment.RunID != evidence.RunID || environment.RunAttempt != evidence.RunAttempt || environment.WorkflowSHA != evidence.WorkflowSHA || environment.Digest == "" || environment.Digest != digestGuardianEnvironment(environment) || environment.Digest != evidence.ObserverEnvironmentDigest || !validObserverFreshness(environment.ObservedAt, environment.ValidUntil, now) {
		return fmt.Errorf("guardian observer environment evidence is missing, tampered, or unbound")
	}
	return nil
}
