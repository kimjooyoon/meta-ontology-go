package main

import (
	"fmt"
)

func buildFailureManifest(input failureInput, binding failureBinding) (failureManifest, error) {
	if err := validateFailureCatalog(); err != nil {
		return failureManifest{}, err
	}
	if err := validateFailureBranchBinding(binding); err != nil {
		return failureManifest{}, err
	}
	entry, ok := failureCatalog[input.Code]
	if !ok {
		return failureManifest{}, fmt.Errorf("unknown failure code %q", input.Code)
	}
	scope, err := failureScope(binding)
	if err != nil {
		return failureManifest{}, err
	}
	if input.Message == "" || input.Remediation == "" {
		return failureManifest{}, fmt.Errorf("failure message and remediation are required")
	}
	if input.HeadBranch == "" || input.HeadBranch != binding.HeadBranch {
		return failureManifest{}, fmt.Errorf("failure head branch is missing or mismatched")
	}
	manifest := failureManifest{
		Schema: failureSchema, Version: 2, Code: input.Code, FailureCodes: input.FailureCodes, Class: entry.Class, Severity: entry.Severity,
		Scope: scope, BlockingScope: entry.BlockingScope, Parallelizable: entry.Parallelizable,
		SourceCommit: binding.HeadSHA, Repository: binding.Repository, BaseRef: binding.BaseRef, BaseSHA: binding.BaseSHA, HeadSHA: binding.HeadSHA,
		Event: binding.Event, EventRef: binding.EventRef, CheckoutRef: binding.CheckoutRef, PRNumber: binding.PRNumber, RunID: binding.RunID,
		RunAttempt: binding.RunAttempt, WorkflowSHA: binding.WorkflowSHA, Job: input.Job,
		HeadBranch: binding.HeadBranch, CatalogPath: failureCatalogPath, CatalogDigest: failureCatalogDigest,
		CatalogRef: failureCatalogPath + "@" + binding.HeadSHA, CatalogVersion: 2, CatalogSHA256: failureCatalogDigest,
		Rejections: input.Rejections, MissingReasons: input.MissingReasons, Artifacts: input.Artifacts, ProofArtifactRef: input.ProofArtifact,
		ArtifactStatus: input.ArtifactStatus, ArtifactReason: input.ArtifactReason, Message: input.Message, Remediation: input.Remediation,
		TerminalFailures: append([]failureJob(nil), input.TerminalFailures...), TerminalFailureCodes: append([]string(nil), input.TerminalFailureCodes...), TerminalFailureEvidence: append([]terminalFailureEvidence(nil), input.TerminalFailureEvidence...),
		NextOperation: entry.NextOperation,
	}
	if len(manifest.FailureCodes) == 0 {
		manifest.FailureCodes = []string{input.Code}
	}
	manifest.Activity = fmt.Sprintf("urn:gooo:ci-run:%d:%d", binding.RunID, binding.RunAttempt)
	manifest.Agent = "urn:gooo:agent:" + binding.Actor
	manifest.Entity = fmt.Sprintf("urn:gooo:ci-failure:%d:%d:%d:%s", binding.RunID, binding.RunAttempt, input.Job.ID, input.Code)
	manifest.ArtifactURLs = failureArtifactRefs(binding, manifest.Artifacts, manifest.ProofArtifactRef)
	manifest.ArtifactRefs = failureArtifactInputs(manifest.Artifacts, manifest.ProofArtifactRef)
	runRef := fmt.Sprintf("https://github.com/%s/actions/runs/%d", binding.Repository, binding.RunID)
	jobRef := fmt.Sprintf("%s/job/%d", runRef, input.Job.ID)
	sourceRef := fmt.Sprintf("https://github.com/%s/commit/%s", binding.Repository, binding.HeadSHA)
	manifest.EvidenceRefs = failureEvidenceRefs(manifest, runRef, jobRef)
	manifest.Provenance = failureProvenance{
		WasGeneratedBy:    manifest.Activity,
		WasAssociatedWith: manifest.Agent,
		WasDerivedFrom:    []string{runRef, jobRef},
		HadPrimarySource:  append([]string{sourceRef, manifest.CatalogRef, manifest.CatalogSHA256}, manifest.ArtifactURLs...),
	}
	if err := validateFailureManifest(manifest, binding); err != nil {
		return failureManifest{}, err
	}
	return manifest, nil
}
