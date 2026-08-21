package main

import (
	"fmt"
)

func validateFailureManifest(manifest failureManifest, binding failureBinding) error {
	entry, ok := failureCatalog[manifest.Code]
	if !ok || manifest.Schema != failureSchema || manifest.Version != 1 {
		return fmt.Errorf("failure manifest schema or code is invalid")
	}
	scope, err := failureScope(binding)
	if err != nil {
		return err
	}
	if manifest.Scope != scope || manifest.Class != entry.Class || manifest.Severity != entry.Severity || manifest.BlockingScope != entry.BlockingScope || manifest.Parallelizable != entry.Parallelizable || manifest.HandoffRequired != entry.HandoffRequired || manifest.HandoffOwner != entry.Owner {
		return fmt.Errorf("failure classification does not match catalog")
	}
	if manifest.SourceCommit != binding.HeadSHA || manifest.Repository != binding.Repository || manifest.BaseRef != binding.BaseRef || manifest.BaseSHA != binding.BaseSHA || manifest.HeadSHA != binding.HeadSHA || manifest.Event != binding.Event || manifest.EventRef != binding.EventRef || manifest.CheckoutRef != binding.CheckoutRef || manifest.PRNumber != binding.PRNumber || manifest.RunID != binding.RunID || manifest.RunAttempt != binding.RunAttempt || manifest.WorkflowSHA != binding.WorkflowSHA || manifest.OwnerBranch != binding.OwnerBranch || manifest.OwnerRef != failureOwnerRef(binding) || !sameArtifactInputs(manifest.ArtifactRefs, failureArtifactInputs(manifest.Artifacts, manifest.ProofArtifactRef)) || !sameFailureJobs(manifest.TerminalFailures, manifest.Job, manifest.TerminalFailureCodes) {
		return fmt.Errorf("failure manifest tuple is stale or mismatched")
	}
	if manifest.Repository == "" || manifest.BaseRef == "" || manifest.OwnerBranch == "" || containsUnknown(manifest.OwnerBranch) || manifest.CatalogPath != failureCatalogPath || manifest.CatalogDigest != failureCatalogDigest || manifest.CatalogRef != failureCatalogPath+"@"+binding.HeadSHA || manifest.CatalogVersion != 1 || manifest.CatalogSHA256 != failureCatalogDigest || !validSHA(manifest.SourceCommit) || !validSHA(manifest.BaseSHA) || !validSHA(manifest.HeadSHA) || !validSHA(manifest.WorkflowSHA) || manifest.BaseSHA == manifest.HeadSHA || !validEventRef(manifest.Event, manifest.EventRef) || manifest.CheckoutRef != manifest.HeadSHA || manifest.RunID <= 0 || manifest.RunAttempt <= 0 || manifest.PRNumber < 0 || manifest.Activity == "" || manifest.Agent == "" || manifest.Entity == "" || manifest.Message == "" || manifest.Remediation == "" || containsUnknown(manifest.Message) || containsUnknown(manifest.Remediation) {
		return fmt.Errorf("failure manifest has incomplete or unknown values")
	}
	if err := validateFailureCodes(manifest.FailureCodes, manifest.Code); err != nil {
		return err
	}
	if err := validateFailureCatalog(); err != nil {
		return err
	}
	if err := validateFailureOwnerBinding(binding); err != nil {
		return err
	}
	if err := validateTerminalFailureMapping(manifest, binding); err != nil {
		return err
	}
	if err := validateFailureEvidence(manifest); err != nil {
		return err
	}
	if binding.Actor == "" || containsUnknown(binding.Actor) || manifest.Activity != fmt.Sprintf("urn:gooo:ci-run:%d:%d", binding.RunID, binding.RunAttempt) || manifest.Agent != "urn:gooo:agent:"+binding.Actor || manifest.Entity != fmt.Sprintf("urn:gooo:ci-failure:%d:%d:%d:%s", binding.RunID, binding.RunAttempt, manifest.Job.ID, manifest.Code) {
		return fmt.Errorf("failure activity, agent, or entity is not derived from the exact tuple")
	}
	if err := validateFailureJob(manifest.Job, binding); err != nil {
		return err
	}
	expected := buildFailureProvenance(manifest, binding)
	if manifest.Provenance.WasGeneratedBy != expected.WasGeneratedBy || manifest.Provenance.WasAssociatedWith != expected.WasAssociatedWith || !sameStrings(manifest.Provenance.WasDerivedFrom, expected.WasDerivedFrom) || !sameStrings(manifest.Provenance.HadPrimarySource, expected.HadPrimarySource) {
		return fmt.Errorf("failure provenance relations are incomplete or mismatched")
	}
	if !sameStrings(manifest.EvidenceRefs, failureEvidenceRefs(manifest, expected.WasDerivedFrom[0], expected.WasDerivedFrom[1])) {
		return fmt.Errorf("failure evidence references are incomplete or mismatched")
	}
	if !sameStrings(manifest.ArtifactURLs, failureArtifactRefs(binding, manifest.Artifacts, manifest.ProofArtifactRef)) {
		return fmt.Errorf("failure artifact URLs are incomplete or mismatched")
	}
	return nil
}
