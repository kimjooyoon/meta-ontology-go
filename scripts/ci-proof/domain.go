package main

import (
	"fmt"
)

type domainEvidence struct {
	Schema              string               `json:"schema"`
	Repository          string               `json:"repository"`
	Event               string               `json:"event"`
	BaseRef             string               `json:"base_ref"`
	BaseSHA             string               `json:"base_sha"`
	HeadSHA             string               `json:"head_sha"`
	EventRef            string               `json:"event_ref"`
	CheckoutRef         string               `json:"checkout_ref"`
	RunID               int64                `json:"run_id"`
	RunAttempt          int64                `json:"run_attempt"`
	WorkflowSHA         string               `json:"workflow_sha"`
	CLI                 domainCommand        `json:"cli"`
	Graph               domainCommand        `json:"graph"`
	ObserverReceiptRefs []string             `json:"observer_receipt_refs"`
	ObserverStatus      string               `json:"observer_status"`
	ProtectionStatus    string               `json:"protection_status"`
	ProvenanceStatus    string               `json:"provenance_status"`
	Digests             domainEvidenceDigest `json:"digests"`
	MissingReasons      missingReasons       `json:"missing_reasons"`
}

type domainCommand struct {
	Command      string `json:"command"`
	Fixture      string `json:"fixture,omitempty"`
	Status       string `json:"status"`
	Available    bool   `json:"available"`
	Reason       string `json:"reason,omitempty"`
	Output       string `json:"output,omitempty"`
	OutputSHA256 string `json:"output_sha256"`
}

type domainEvidenceDigest struct {
	SourceSHA256    string `json:"source_sha256"`
	IRSHA256        string `json:"ir_sha256"`
	GeneratedSHA256 string `json:"generated_output_sha256"`
	BundleSHA256    string `json:"bundle_sha256"`
	DomainSHA256    string `json:"domain_sha256"`
}

func validateDomainEvidence(domain domainEvidence, evidence evidenceInput, context contextInput) error {
	if domain.Schema != domainEvidenceSchema || domain.Repository != evidence.Repository || domain.Event != evidence.Event || domain.EventRef != evidence.EventRef || domain.CheckoutRef != evidence.CheckoutRef || domain.BaseRef != evidence.BaseRef || domain.BaseSHA != evidence.BaseSHA || domain.HeadSHA != evidence.HeadSHA || domain.EventRef != context.EventRef || domain.CheckoutRef != context.CheckoutRef || domain.RunID != evidence.RunID || domain.RunAttempt != evidence.Attempt || domain.WorkflowSHA != evidence.WorkflowSHA {
		return fmt.Errorf("domain evidence identity is incomplete or mismatched")
	}
	if domain.CLI.Status != "verified" || !domain.CLI.Available || domain.CLI.Command == "" || domain.CLI.Fixture == "" || domain.CLI.OutputSHA256 != digestBytes([]byte(domain.CLI.Output)) {
		return fmt.Errorf("CLI domain evidence is missing or mismatched")
	}
	if domain.Graph.Status != "deferred" || domain.Graph.Available || domain.Graph.Command == "" {
		return fmt.Errorf("graph domain evidence must remain explicitly deferred")
	}
	if domain.ObserverStatus != "unavailable" || len(domain.ObserverReceiptRefs) != 0 || domain.ProtectionStatus != "unavailable" || domain.ProvenanceStatus != "not_applicable" {
		return fmt.Errorf("domain evidence status is not CI-only and deterministic")
	}
	if err := validateMissingReasons(domain.MissingReasons, domain.ProtectionStatus, domain.ProvenanceStatus); err != nil {
		return err
	}
	if domain.Digests.SourceSHA256 != evidence.Digests.Source || domain.Digests.IRSHA256 != evidence.Digests.IR || domain.Digests.GeneratedSHA256 != evidence.Digests.Generated || domain.Digests.BundleSHA256 != evidence.Digests.Bundle || !validDigest(domain.Digests.DomainSHA256) {
		return fmt.Errorf("domain evidence digests are missing or mismatched")
	}
	if domain.Digests.DomainSHA256 != digestDomainEvidence(domain) {
		return fmt.Errorf("domain evidence digest mismatch")
	}
	return nil
}
