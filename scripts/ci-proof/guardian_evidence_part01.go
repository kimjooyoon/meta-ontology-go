package main

import (
	"fmt"
	"reflect"
	"time"
)

const guardianEvidenceSchema = "gooo/ci-guardian-evidence/v1"
const guardianInstallationRepository = "kimjooyoon/meta-ontology-go"

var guardianEvidenceActions = map[string]bool{
	"opened":           true,
	"synchronize":      true,
	"reopened":         true,
	"ready_for_review": true,
}

func validateGuardianEvidence(evidence *guardianEvidence, bundle proofBundle) error {
	return validateGuardianEvidenceAt(evidence, bundle, time.Now().UTC())
}
func validateGuardianEvidenceAt(evidence *guardianEvidence, bundle proofBundle, now time.Time) error {
	if bundle.BaseRef != "main" {
		if evidence != nil {
			return fmt.Errorf("guardian evidence is not allowed on a dev feature proof")
		}
		return nil
	}
	if evidence == nil || evidence.Schema != guardianEvidenceSchema || evidence.Route != "promotion_main" || evidence.CheckName != "CI guardian" || evidence.Repository != bundle.Repository || evidence.PRNumber != bundle.PRNumber || !guardianEvidenceActions[evidence.Action] || evidence.BaseRepo != bundle.Repository || evidence.BaseRef != "main" || evidence.BaseSHA != bundle.BaseSHA || evidence.HeadRepo != bundle.Repository || evidence.HeadRef != "dev" || evidence.HeadSHA != bundle.HeadSHA || evidence.WorkflowSHA != bundle.HeadSHA || evidence.RuntimeSHA != evidence.WorkflowSHA || evidence.WorkflowRef != bundle.Repository+"/.github/workflows/ci-guardian.yml@refs/heads/dev" || evidence.RuntimeRef != "refs/heads/dev" || evidence.EventRef != "refs/heads/dev" || evidence.DefaultBranch != "dev" || evidence.HeadBindingStatus != "verified" || evidence.RunID <= 0 || evidence.RunAttempt <= 0 {
		return fmt.Errorf("guardian promotion evidence identity is incomplete or mismatched")
	}
	if evidence.ObserverEnvironmentName != "guardian-observer" {
		return fmt.Errorf("guardian promotion evidence is not bound to the protected observer environment")
	}
	if bundle.PromotionObservation == nil || bundle.PromotionObservation.Action != evidence.Action || bundle.PromotionObservation.Topology != evidence.Topology {
		return fmt.Errorf("guardian promotion evidence is not cross-bound to the live observation")
	}
	if !validSHA(evidence.BaseSHA) || !validSHA(evidence.HeadSHA) || !validSHA(evidence.WorkflowSHA) || evidence.LiveRefsBefore.MainSHA != evidence.BaseSHA || evidence.LiveRefsAfter.MainSHA != evidence.BaseSHA || evidence.LiveRefsBefore.DevSHA != evidence.HeadSHA || evidence.LiveRefsAfter.DevSHA != evidence.HeadSHA || evidence.Topology.Status != "ahead" || evidence.Topology.AheadBy <= 0 || evidence.Topology.BehindBy != 0 || evidence.Topology.MergeBaseSHA != evidence.BaseSHA {
		return fmt.Errorf("guardian promotion topology is not exact")
	}
	if _, err := time.Parse(time.RFC3339, evidence.RunCreatedAt); err != nil {
		return fmt.Errorf("guardian observer run creation time is not RFC3339")
	}
	if evidence.WorkflowID <= 0 || evidence.WorkflowPath != ".github/workflows/ci-guardian.yml" || evidence.RunEvent != "pull_request_target" || evidence.RunStatus != "completed" || evidence.RunConclusion != "success" || evidence.RunNumber <= 0 || evidence.GuardianJobID <= 0 || evidence.GuardianJobName != "CI guardian" || evidence.GuardianJobStatus != "completed" || evidence.GuardianJobConclusion != "success" || evidence.GuardianJobHeadSHA != evidence.HeadSHA || evidence.CheckRunID <= 0 || evidence.GuardianJobID != evidence.CheckRunID || evidence.CheckRunName != "CI guardian" || evidence.CheckRunAppID != 15368 || evidence.CheckRunStatus != "completed" || evidence.CheckRunConclusion != "success" || evidence.CheckRunHeadSHA != evidence.HeadSHA || evidence.CheckSuiteID <= 0 {
		return fmt.Errorf("guardian observer run, job, or check identity is incomplete")
	}
	if err := validateBranchProtectionAt(evidence.BranchProtection, evidenceInput{Repository: bundle.Repository, BaseSHA: evidence.BaseSHA, HeadSHA: evidence.HeadSHA, Digests: evidenceDigests{Policy: bundle.Digests.Policy}}, contextInput{BaseRef: "main"}, now); err != nil {
		return fmt.Errorf("guardian branch protection evidence is invalid: %w", err)
	}
	if err := validateTrustedBranchProtectionAt(evidence.DevBranchProtection, evidenceInput{Repository: bundle.Repository, BaseSHA: evidence.BaseSHA, HeadSHA: evidence.HeadSHA, EventRef: evidence.EventRef, CheckoutRef: evidence.WorkflowSHA, RunID: evidence.RunID, Attempt: evidence.RunAttempt, WorkflowSHA: evidence.WorkflowSHA, Digests: evidenceDigests{Policy: bundle.Digests.Policy}}, "dev", now); err != nil {
		return fmt.Errorf("guardian dev branch protection evidence is invalid: %w", err)
	}
	if err := validateGuardianEnvironmentEvidenceAt(evidence.ObserverEnvironmentSnapshot, evidence, bundle, now); err != nil {
		return err
	}
	if err := validateGuardianInstallationScopeAt(evidence.InstallationRepositoryScope, evidence, bundle, now); err != nil {
		return err
	}
	if !reflect.DeepEqual(bundle.BranchProtection, evidence.BranchProtection) {
		return fmt.Errorf("proof branch protection is not the exact Guardian snapshot")
	}
	if !reflect.DeepEqual(bundle.DevBranchProtection, evidence.DevBranchProtection) {
		return fmt.Errorf("proof dev branch protection is not the exact Guardian snapshot")
	}
	if evidence.BranchProtection.EventRef != evidence.RuntimeRef || evidence.BranchProtection.CheckoutRef != evidence.WorkflowSHA || evidence.BranchProtection.RunID != evidence.RunID || evidence.BranchProtection.RunAttempt != evidence.RunAttempt || evidence.BranchProtection.WorkflowSHA != evidence.WorkflowSHA {
		return fmt.Errorf("guardian branch protection evidence is not bound to its observer run")
	}
	expectedArtifactName := fmt.Sprintf("ci-guardian-%d-%d", evidence.RunID, evidence.RunAttempt)
	if evidence.ArtifactID <= 0 || evidence.ArtifactName != expectedArtifactName || evidence.ArtifactSize <= 0 || evidence.ArtifactExpired || !validArtifactDigest(evidence.ArtifactDigest) || !validArtifactDigest(evidence.ManifestBundleSHA) || evidence.BundleSHA != evidence.ManifestBundleSHA || evidence.Decision != "PASS" || evidence.Code != nil {
		return fmt.Errorf("guardian promotion artifact is missing, stale, or not PASS")
	}
	return nil
}
