package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCIBranchProtectionUnavailableIsNotReady(t *testing.T) {
	bundle := validProof()
	bundle.BranchProtection.Exists = false
	bundle.BranchProtection.ReadStatus = "unavailable"
	bundle.BranchProtection.MissingReason = "branch_protection_token_unavailable"
	if branchProtectionReady(bundle.BranchProtection) {
		t.Fatal("unavailable branch protection snapshot was promotion-ready")
	}
}
func TestCIMissingReasonIsRequiredForUnavailableEvidence(t *testing.T) {
	bundle := validProof()
	bundle.MissingReasons.Protection = ""
	bundle.DomainEvidence.MissingReasons.Protection = ""
	if err := validateProof(bundle); err == nil {
		t.Fatal("unavailable protection without a reason was accepted")
	}
}
func TestCIDomainEvidenceOutputTamperingFailsClosed(t *testing.T) {
	bundle := validProof()
	bundle.DomainEvidence.CLI.Output += "tampered"
	if err := validateProof(bundle); err == nil {
		t.Fatal("tampered CLI domain evidence was accepted")
	}
}
func TestCIDomainEvidenceCanonicalDigestOmitsDeferredFixture(t *testing.T) {
	bundle := validProof()
	data, err := json.Marshal(bundle.DomainEvidence)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"graph":{"command":"go run ./cmd/gooo graph-dump examples/billing/main.gooo","fixture"`) {
		t.Fatal("deferred graph fixture was serialized despite being unavailable")
	}
	if err := validateProof(bundle); err != nil {
		t.Fatalf("canonical deferred graph evidence was rejected: %v", err)
	}
}
func TestCIRefSeparationRejectsCheckoutMismatch(t *testing.T) {
	bundle := validProof()
	bundle.CheckoutRef = strings.Repeat("b", 40)
	if err := validateProof(bundle); err == nil {
		t.Fatal("checkout ref mismatch was accepted")
	}
}
func TestCIRefSeparationRejectsEventRefMismatch(t *testing.T) {
	bundle := validProof()
	bundle.EventRef = "refs/pull/2/merge"
	if err := validateProof(bundle); err == nil {
		t.Fatal("event ref mismatch was accepted")
	}
}
