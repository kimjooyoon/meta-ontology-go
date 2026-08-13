package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const failureSchema = "gooo/ci-failure/v1"

const failureCatalogPath = "scripts/ci-proof/docs/failure-reasons.md"

type failureCatalogEntry struct {
	Class           string
	Severity        string
	BlockingScope   string
	Parallelizable  bool
	HandoffRequired bool
}

var failureCatalog = map[string]failureCatalogEntry{
	"CI-TEST-001":       {Class: "test", Severity: "error", BlockingScope: "local", Parallelizable: true, HandoffRequired: false},
	"CI-SCOPE-001":      {Class: "scope", Severity: "error", BlockingScope: "global", Parallelizable: false, HandoffRequired: false},
	"CI-CONTRACT-001":   {Class: "contract", Severity: "critical", BlockingScope: "global", Parallelizable: false, HandoffRequired: true},
	"CI-DEPENDENCY-001": {Class: "dependency", Severity: "warning", BlockingScope: "local", Parallelizable: true, HandoffRequired: true},
	"CI-GATE-001":       {Class: "gate", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true},
	"CI-ARTIFACT-001":   {Class: "artifact", Severity: "error", BlockingScope: "global", Parallelizable: false, HandoffRequired: false},
	"CI-FRESHNESS-001":  {Class: "freshness", Severity: "error", BlockingScope: "global", Parallelizable: false, HandoffRequired: false},
	"CI-PROVENANCE-001": {Class: "provenance", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true},
	"CI-OWNERSHIP-001":  {Class: "ownership", Severity: "blocked", BlockingScope: "local", Parallelizable: true, HandoffRequired: true},
}

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
	Code        string     `json:"code"`
	Message     string     `json:"message"`
	Remediation string     `json:"remediation"`
	Job         failureJob `json:"job"`
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
}

type failureProvenance struct {
	WasGeneratedBy    string   `json:"wasGeneratedBy"`
	WasAssociatedWith string   `json:"wasAssociatedWith"`
	WasDerivedFrom    []string `json:"wasDerivedFrom"`
	HadPrimarySource  []string `json:"hadPrimarySource"`
}

type failureManifest struct {
	Schema          string            `json:"schema"`
	Version         int               `json:"version"`
	Code            string            `json:"code"`
	Class           string            `json:"class"`
	Severity        string            `json:"severity"`
	Scope           string            `json:"scope"`
	BlockingScope   string            `json:"blocking_scope"`
	Parallelizable  bool              `json:"parallelizable"`
	SourceCommit    string            `json:"source_commit"`
	Repository      string            `json:"repository"`
	BaseRef         string            `json:"base_ref"`
	BaseSHA         string            `json:"base_sha"`
	HeadSHA         string            `json:"head_sha"`
	Event           string            `json:"event"`
	EventRef        string            `json:"event_ref"`
	CheckoutRef     string            `json:"checkout_ref"`
	PRNumber        int64             `json:"pr_number"`
	RunID           int64             `json:"run_id"`
	RunAttempt      int64             `json:"run_attempt"`
	WorkflowSHA     string            `json:"workflow_sha"`
	Job             failureJob        `json:"job"`
	Activity        string            `json:"activity"`
	Agent           string            `json:"agent"`
	Entity          string            `json:"entity"`
	Provenance      failureProvenance `json:"provenance"`
	EvidenceRefs    []string          `json:"evidence_refs"`
	CatalogPath     string            `json:"catalog_path"`
	Message         string            `json:"message"`
	Remediation     string            `json:"remediation"`
	HandoffRequired bool              `json:"handoff_required"`
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
	manifest := failureManifest{
		Schema: failureSchema, Version: 1, Code: input.Code, Class: entry.Class, Severity: entry.Severity,
		Scope: scope, BlockingScope: entry.BlockingScope, Parallelizable: entry.Parallelizable,
		SourceCommit: binding.HeadSHA, Repository: binding.Repository, BaseRef: binding.BaseRef, BaseSHA: binding.BaseSHA, HeadSHA: binding.HeadSHA,
		Event: binding.Event, EventRef: binding.EventRef, CheckoutRef: binding.CheckoutRef, PRNumber: binding.PRNumber, RunID: binding.RunID,
		RunAttempt: binding.RunAttempt, WorkflowSHA: binding.WorkflowSHA, Job: input.Job,
		CatalogPath: failureCatalogPath, Message: input.Message, Remediation: input.Remediation,
		HandoffRequired: entry.HandoffRequired,
	}
	manifest.Activity = fmt.Sprintf("urn:gooo:ci-run:%d:%d", binding.RunID, binding.RunAttempt)
	manifest.Agent = "urn:gooo:agent:" + binding.Actor
	manifest.Entity = fmt.Sprintf("urn:gooo:ci-failure:%d:%d:%d:%s", binding.RunID, binding.RunAttempt, input.Job.ID, input.Code)
	runRef := fmt.Sprintf("https://github.com/%s/actions/runs/%d", binding.Repository, binding.RunID)
	jobRef := fmt.Sprintf("%s/job/%d", runRef, input.Job.ID)
	sourceRef := fmt.Sprintf("https://github.com/%s/commit/%s", binding.Repository, binding.HeadSHA)
	manifest.EvidenceRefs = []string{runRef, jobRef, failureCatalogPath}
	manifest.Provenance = failureProvenance{
		WasGeneratedBy:    manifest.Activity,
		WasAssociatedWith: manifest.Agent,
		WasDerivedFrom:    []string{runRef, jobRef},
		HadPrimarySource:  []string{sourceRef, failureCatalogPath},
	}
	if err := validateFailureManifest(manifest, binding); err != nil {
		return failureManifest{}, err
	}
	return manifest, nil
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
	if manifest.Scope != scope || manifest.Class != entry.Class || manifest.Severity != entry.Severity || manifest.BlockingScope != entry.BlockingScope || manifest.Parallelizable != entry.Parallelizable || manifest.HandoffRequired != entry.HandoffRequired {
		return fmt.Errorf("failure classification does not match catalog")
	}
	if manifest.SourceCommit != binding.HeadSHA || manifest.Repository != binding.Repository || manifest.BaseRef != binding.BaseRef || manifest.BaseSHA != binding.BaseSHA || manifest.HeadSHA != binding.HeadSHA || manifest.Event != binding.Event || manifest.EventRef != binding.EventRef || manifest.CheckoutRef != binding.CheckoutRef || manifest.PRNumber != binding.PRNumber || manifest.RunID != binding.RunID || manifest.RunAttempt != binding.RunAttempt || manifest.WorkflowSHA != binding.WorkflowSHA {
		return fmt.Errorf("failure manifest tuple is stale or mismatched")
	}
	if manifest.Repository == "" || manifest.BaseRef == "" || manifest.CatalogPath != failureCatalogPath || !validSHA(manifest.SourceCommit) || !validSHA(manifest.BaseSHA) || !validSHA(manifest.HeadSHA) || !validSHA(manifest.WorkflowSHA) || manifest.BaseSHA == manifest.HeadSHA || !validEventRef(manifest.Event, manifest.EventRef) || manifest.CheckoutRef != manifest.HeadSHA || manifest.RunID <= 0 || manifest.RunAttempt <= 0 || manifest.PRNumber < 0 || manifest.Activity == "" || manifest.Agent == "" || manifest.Entity == "" || manifest.Message == "" || manifest.Remediation == "" || containsUnknown(manifest.Message) || containsUnknown(manifest.Remediation) {
		return fmt.Errorf("failure manifest has incomplete or unknown values")
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
	if !sameStrings(manifest.EvidenceRefs, []string{expected.WasDerivedFrom[0], expected.WasDerivedFrom[1], failureCatalogPath}) {
		return fmt.Errorf("failure evidence references are incomplete or mismatched")
	}
	return nil
}
