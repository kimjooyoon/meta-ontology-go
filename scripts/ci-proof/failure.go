package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const failureSchema = "gooo/ci-failure/v1"

const failureCatalogPath = "scripts/ci-proof/docs/failure-reasons.md"

type failureJob struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"head_sha"`
	RunID      int64  `json:"run_id"`
	RunAttempt int64  `json:"run_attempt"`
}

type failureInput struct {
	Code                 string          `json:"code"`
	FailureCodes         []string        `json:"failure_codes"`
	Message              string          `json:"message"`
	Remediation          string          `json:"remediation"`
	OwnerBranch          string          `json:"owner_branch"`
	Rejections           []string        `json:"rejections"`
	MissingReasons       missingReasons  `json:"missing_reasons"`
	Artifacts            []artifactInput `json:"artifacts"`
	ProofArtifact        *artifactInput  `json:"proof_artifact"`
	ArtifactStatus       string          `json:"artifact_status"`
	ArtifactReason       string          `json:"artifact_reason"`
	TerminalFailures     []failureJob    `json:"terminal_failures"`
	TerminalFailureCodes []string        `json:"terminal_failure_codes"`
	Job                  failureJob      `json:"job"`
}

type failureBinding struct {
	Repository  string
	Event       string
	EventRef    string
	CheckoutRef string
	BaseRef     string
	BaseSHA     string
	HeadSHA     string
	WorkflowSHA string
	PRNumber    int64
	RunID       int64
	RunAttempt  int64
	Actor       string
	OwnerBranch string
}

type failureProvenance struct {
	WasGeneratedBy    string   `json:"wasGeneratedBy"`
	WasAssociatedWith string   `json:"wasAssociatedWith"`
	WasDerivedFrom    []string `json:"wasDerivedFrom"`
	HadPrimarySource  []string `json:"hadPrimarySource"`
}

type failureManifest struct {
	Schema               string            `json:"schema"`
	Version              int               `json:"version"`
	Code                 string            `json:"code"`
	FailureCodes         []string          `json:"failure_codes"`
	Class                string            `json:"class"`
	Severity             string            `json:"severity"`
	Scope                string            `json:"scope"`
	BlockingScope        string            `json:"blocking_scope"`
	Parallelizable       bool              `json:"parallelizable"`
	SourceCommit         string            `json:"source_commit"`
	Repository           string            `json:"repository"`
	BaseRef              string            `json:"base_ref"`
	BaseSHA              string            `json:"base_sha"`
	HeadSHA              string            `json:"head_sha"`
	Event                string            `json:"event"`
	EventRef             string            `json:"event_ref"`
	CheckoutRef          string            `json:"checkout_ref"`
	PRNumber             int64             `json:"pr_number"`
	RunID                int64             `json:"run_id"`
	RunAttempt           int64             `json:"run_attempt"`
	WorkflowSHA          string            `json:"workflow_sha"`
	Job                  failureJob        `json:"job"`
	Activity             string            `json:"activity"`
	Agent                string            `json:"agent"`
	OwnerBranch          string            `json:"owner_branch"`
	OwnerRef             string            `json:"owner_ref"`
	Entity               string            `json:"entity"`
	Provenance           failureProvenance `json:"provenance"`
	EvidenceRefs         []string          `json:"evidence_refs"`
	CatalogPath          string            `json:"catalog_path"`
	CatalogDigest        string            `json:"catalog_digest"`
	CatalogRef           string            `json:"catalog_ref"`
	CatalogVersion       int               `json:"catalog_version"`
	CatalogSHA256        string            `json:"catalog_sha256"`
	Rejections           []string          `json:"rejections"`
	MissingReasons       missingReasons    `json:"missing_reasons"`
	Artifacts            []artifactInput   `json:"artifacts"`
	ProofArtifactRef     *artifactInput    `json:"proof_artifact_ref,omitempty"`
	ArtifactURLs         []string          `json:"artifact_urls"`
	ArtifactRefs         []artifactInput   `json:"artifact_refs"`
	ArtifactStatus       string            `json:"artifact_status"`
	ArtifactReason       string            `json:"artifact_reason"`
	TerminalFailures     []failureJob      `json:"terminal_failures"`
	TerminalFailureCodes []string          `json:"terminal_failure_codes"`
	Message              string            `json:"message"`
	Remediation          string            `json:"remediation"`
	HandoffOwner         string            `json:"handoff_owner"`
	HandoffRequired      bool              `json:"handoff_required"`
}

func writeFailureManifest(inputPath, outputPath string) error {
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read failure input: %w", err)
	}
	var input failureInput
	if err := json.Unmarshal(inputData, &input); err != nil {
		return fmt.Errorf("parse failure input: %w", err)
	}
	binding, err := readFailureBinding()
	if err != nil {
		return err
	}
	manifest, err := buildFailureManifest(input, binding)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal failure manifest: %w", err)
	}
	if err := os.WriteFile(outputPath, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write failure manifest: %w", err)
	}
	return nil
}

func buildFailureManifest(input failureInput, binding failureBinding) (failureManifest, error) {
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
	if input.OwnerBranch == "" || input.OwnerBranch != binding.OwnerBranch {
		return failureManifest{}, fmt.Errorf("failure owner branch is missing or mismatched")
	}
	manifest := failureManifest{
		Schema: failureSchema, Version: 1, Code: input.Code, FailureCodes: input.FailureCodes, Class: entry.Class, Severity: entry.Severity,
		Scope: scope, BlockingScope: entry.BlockingScope, Parallelizable: entry.Parallelizable,
		SourceCommit: binding.HeadSHA, Repository: binding.Repository, BaseRef: binding.BaseRef, BaseSHA: binding.BaseSHA, HeadSHA: binding.HeadSHA,
		Event: binding.Event, EventRef: binding.EventRef, CheckoutRef: binding.CheckoutRef, PRNumber: binding.PRNumber, RunID: binding.RunID,
		RunAttempt: binding.RunAttempt, WorkflowSHA: binding.WorkflowSHA, Job: input.Job,
		OwnerBranch: binding.OwnerBranch, OwnerRef: failureOwnerRef(binding), CatalogPath: failureCatalogPath, CatalogDigest: failureCatalogDigest,
		CatalogRef: failureCatalogPath + "@" + binding.HeadSHA, CatalogVersion: 1, CatalogSHA256: failureCatalogDigest,
		Rejections: input.Rejections, MissingReasons: input.MissingReasons, Artifacts: input.Artifacts, ProofArtifactRef: input.ProofArtifact,
		ArtifactStatus: input.ArtifactStatus, ArtifactReason: input.ArtifactReason, Message: input.Message, Remediation: input.Remediation,
		TerminalFailures: append([]failureJob(nil), input.TerminalFailures...), TerminalFailureCodes: append([]string(nil), input.TerminalFailureCodes...),
		HandoffRequired: entry.HandoffRequired, HandoffOwner: entry.Owner,
	}
	if len(manifest.FailureCodes) == 0 {
		manifest.FailureCodes = []string{input.Code}
	}
	manifest.Activity = fmt.Sprintf("urn:gooo:ci-run:%d:%d", binding.RunID, binding.RunAttempt)
	manifest.Agent = "urn:gooo:agent:" + binding.Actor
	manifest.Entity = fmt.Sprintf("urn:gooo:ci-failure:%d:%d:%d:%s", binding.RunID, binding.RunAttempt, input.Job.ID, input.Code)
	manifest.ArtifactURLs = failureArtifactRefs(binding, manifest.Artifacts, manifest.ProofArtifactRef)
	manifest.ArtifactRefs = append([]artifactInput(nil), manifest.Artifacts...)
	runRef := fmt.Sprintf("https://github.com/%s/actions/runs/%d", binding.Repository, binding.RunID)
	jobRef := fmt.Sprintf("%s/job/%d", runRef, input.Job.ID)
	sourceRef := fmt.Sprintf("https://github.com/%s/commit/%s", binding.Repository, binding.HeadSHA)
	manifest.EvidenceRefs = failureEvidenceRefs(manifest, runRef, jobRef)
	manifest.Provenance = failureProvenance{
		WasGeneratedBy:    manifest.Activity,
		WasAssociatedWith: manifest.Agent,
		WasDerivedFrom:    []string{runRef, jobRef},
		HadPrimarySource:  append([]string{sourceRef, manifest.OwnerRef, manifest.CatalogRef, manifest.CatalogSHA256}, manifest.ArtifactURLs...),
	}
	if err := validateFailureManifest(manifest, binding); err != nil {
		return failureManifest{}, err
	}
	return manifest, nil
}

func failureOwnerRef(binding failureBinding) string {
	return fmt.Sprintf("https://github.com/%s/blob/%s/.github/ci-governance.json", binding.Repository, binding.HeadSHA)
}

func failureArtifactRefs(binding failureBinding, artifacts []artifactInput, proofArtifact *artifactInput) []string {
	refs := make([]string, 0, len(artifacts)+1)
	for _, artifact := range artifacts {
		refs = append(refs, fmt.Sprintf("https://github.com/%s/actions/runs/%d/artifacts/%d", binding.Repository, binding.RunID, artifact.ID))
	}
	if proofArtifact != nil {
		refs = append(refs, fmt.Sprintf("https://github.com/%s/actions/runs/%d/artifacts/%d", binding.Repository, binding.RunID, proofArtifact.ID))
	}
	return refs
}

func failureEvidenceRefs(manifest failureManifest, runRef, jobRef string) []string {
	refs := []string{runRef, jobRef, manifest.OwnerRef}
	refs = append(refs, manifest.ArtifactURLs...)
	return append(refs, manifest.CatalogRef, manifest.CatalogSHA256)
}

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
	if manifest.SourceCommit != binding.HeadSHA || manifest.Repository != binding.Repository || manifest.BaseRef != binding.BaseRef || manifest.BaseSHA != binding.BaseSHA || manifest.HeadSHA != binding.HeadSHA || manifest.Event != binding.Event || manifest.EventRef != binding.EventRef || manifest.CheckoutRef != binding.CheckoutRef || manifest.PRNumber != binding.PRNumber || manifest.RunID != binding.RunID || manifest.RunAttempt != binding.RunAttempt || manifest.WorkflowSHA != binding.WorkflowSHA || manifest.OwnerBranch != binding.OwnerBranch || manifest.OwnerRef != failureOwnerRef(binding) || !sameArtifactInputs(manifest.ArtifactRefs, manifest.Artifacts) || !sameFailureJobs(manifest.TerminalFailures, manifest.Job, manifest.TerminalFailureCodes) {
		return fmt.Errorf("failure manifest tuple is stale or mismatched")
	}
	if manifest.Repository == "" || manifest.BaseRef == "" || manifest.OwnerBranch == "" || containsUnknown(manifest.OwnerBranch) || manifest.CatalogPath != failureCatalogPath || manifest.CatalogDigest != failureCatalogDigest || manifest.CatalogRef != failureCatalogPath+"@"+binding.HeadSHA || manifest.CatalogVersion != 1 || manifest.CatalogSHA256 != failureCatalogDigest || !validSHA(manifest.SourceCommit) || !validSHA(manifest.BaseSHA) || !validSHA(manifest.HeadSHA) || !validSHA(manifest.WorkflowSHA) || manifest.BaseSHA == manifest.HeadSHA || !validEventRef(manifest.Event, manifest.EventRef) || manifest.CheckoutRef != manifest.HeadSHA || manifest.RunID <= 0 || manifest.RunAttempt <= 0 || manifest.PRNumber < 0 || manifest.Activity == "" || manifest.Agent == "" || manifest.Entity == "" || manifest.Message == "" || manifest.Remediation == "" || containsUnknown(manifest.Message) || containsUnknown(manifest.Remediation) {
		return fmt.Errorf("failure manifest has incomplete or unknown values")
	}
	if err := validateFailureCodes(manifest.FailureCodes, manifest.Code); err != nil {
		return err
	}
	if err := validateTerminalFailureMapping(manifest); err != nil {
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
