package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDomainEvidenceDigestUsesCanonicalNotApplicableReasonShape(t *testing.T) {
	bundle := validProof()
	domain := bundle.DomainEvidence
	domain.Digests.DomainSHA256 = ""
	payload, err := json.Marshal(domain)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), `"provenance":""`) {
		t.Fatalf("empty not-applicable provenance reason changed the canonical domain payload: %s", payload)
	}
	domain.Digests.DomainSHA256 = digestDomainEvidence(domain)
	evidence := evidenceInput{
		Repository: bundle.Repository, Event: bundle.Event, EventRef: bundle.EventRef,
		CheckoutRef: bundle.CheckoutRef, BaseRef: bundle.BaseRef, BaseSHA: bundle.BaseSHA,
		HeadSHA: bundle.HeadSHA, RunID: bundle.RunID, Attempt: bundle.RunAttempt,
		WorkflowSHA: bundle.WorkflowSHA, Digests: evidenceDigests{Source: bundle.Digests.Source, IR: bundle.Digests.Semantic, Generated: bundle.Digests.Projection, Bundle: domain.Digests.BundleSHA256},
	}
	if err := validateDomainEvidence(domain, evidence, contextInput{EventRef: bundle.EventRef, CheckoutRef: bundle.CheckoutRef}); err != nil {
		t.Fatalf("canonical not-applicable domain evidence was rejected: %v", err)
	}
	tampered := domain
	tampered.MissingReasons.Provenance = "invented_provenance_reason"
	tampered.Digests.DomainSHA256 = digestDomainEvidence(tampered)
	if err := validateDomainEvidence(tampered, evidence, contextInput{EventRef: bundle.EventRef, CheckoutRef: bundle.CheckoutRef}); err == nil {
		t.Fatal("not_applicable provenance accepted a fabricated missing reason")
	}
}
