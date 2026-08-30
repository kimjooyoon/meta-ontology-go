package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// promotionOperatorReady is a pure, non-mutating predicate for a future
// external fast-forward operator. It never writes refs or protection.
func promotionOperatorReady(bundle proofBundle) bool {
	if bundle.Decision != "PASS" || bundle.Event != "pull_request" || bundle.BaseRef != "main" || bundle.PRNumber <= 0 || !validSHA(bundle.BaseSHA) || !validSHA(bundle.HeadSHA) || bundle.Repository == "" {
		return false
	}
	if validatePromotionObservation(bundle.PromotionObservation, bundle) != nil || validatePromotionAuthorization(bundle) != nil || !promotionProofCoreReady(bundle) {
		return false
	}
	return bundle.PromotionAuthorization.Decision == "PASS" && bundle.PromotionAuthorization.Code == nil && bundle.PromotionAuthorization.ProofDigest == bundle.Digests.Bundle
}
func statusFor(decision string) string {
	if decision == "passed" {
		return "verified"
	}
	return "rejected"
}
func makeReceipt(bundle proofBundle, context contextInput) provenanceReceipt {
	return provenanceReceipt{Schema: receiptSchema, Operation: "verify", Relation: "conformance", Delta: "ci-policy", AllowedIntent: "verification-only", Locality: "repository", Repository: bundle.Repository, Event: bundle.Event, BaseRef: bundle.BaseRef, BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA, Ref: bundle.Ref, EventRef: bundle.EventRef, CheckoutRef: bundle.CheckoutRef, PRNumber: bundle.PRNumber, RunID: bundle.RunID, RunAttempt: bundle.RunAttempt, WorkflowSHA: bundle.WorkflowSHA, BranchProtection: bundle.BranchProtection, DevBranchProtection: bundle.DevBranchProtection, DomainEvidence: bundle.DomainEvidence, ArtifactProvenance: bundle.ArtifactProvenance, GuardianEvidence: bundle.GuardianEvidence, FoundationPromotion: bundle.FoundationPromotion, PromotionObservation: bundle.PromotionObservation, PromotionAuthorization: bundle.PromotionAuthorization, Jobs: bundle.Jobs, Artifacts: bundle.Artifacts, Digests: receiptDigests{Source: bundle.Digests.Source, IR: bundle.Digests.Semantic, Projection: bundle.Digests.Projection, Build: bundle.Digests.Build, Policy: bundle.Digests.Policy, Schema: bundle.Digests.Schema, Toolchain: bundle.Digests.Toolchain, Target: bundle.Digests.Target, Bundle: bundle.Digests.Bundle}, Cache: cacheReceipt{Key: bundle.Cache.Key, Outcome: bundle.Cache.Outcome}, DiagnosticIDs: context.DiagnosticIDs, RepairIDs: context.RepairIDs, WriteEffect: bundle.WriteEffect, Producer: "go-ci-proof", Role: "Gate", Predecessors: context.Predecessors, Decision: bundle.Decision, MissingReasons: bundle.MissingReasons}
}
func marshalProof(bundle proofBundle) ([]byte, error) {
	bundle.Digests.Bundle = ""
	if bundle.PromotionAuthorization != nil {
		authorization := *bundle.PromotionAuthorization
		authorization.ProofDigest = ""
		bundle.PromotionAuthorization = &authorization
	}
	data, err := json.Marshal(bundle)
	if err != nil {
		return nil, fmt.Errorf("marshal CI proof: %w", err)
	}
	return data, nil
}
func writeOutputs(output string, receiptPath string, bundle proofBundle, receipt provenanceReceipt) error {
	encoded, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(output, append(encoded, '\n'), 0o644); err != nil {
		return err
	}
	line, err := json.Marshal(receipt)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(receiptPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(line, '\n'))
	return err
}

func digestDomainEvidence(domain domainEvidence) string {
	domain.Digests.DomainSHA256 = ""
	data, _ := json.Marshal(domain)
	return digestBytes(data)
}
