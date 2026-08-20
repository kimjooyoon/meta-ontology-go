package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
)

func buildProof(inputs proofInputs, digests proofDigests) (proofBundle, provenanceReceipt, error) {
	context := inputs.Context
	rejections := gateRejections(inputs)
	bundle := proofBundle{Schema: proofSchema, Repository: context.Repository, Event: context.Event, PRNumber: context.PRNumber, BaseRef: context.BaseRef, BaseSHA: context.BaseSHA, HeadSHA: context.HeadSHA, Ref: context.Ref, EventRef: context.EventRef, CheckoutRef: context.CheckoutRef, RunID: context.RunID, RunAttempt: context.RunAttempt, WorkflowSHA: context.WorkflowSHA, Scheduler: inputs.Scheduler, Jobs: inputs.Jobs, Actors: actorRoles{Actor: context.Actor, Builder: context.Builder, Gate: context.Gate}, BranchProtection: context.BranchProtection, DevBranchProtection: context.DevBranchProtection, DomainEvidence: context.DomainEvidence, GuardianEvidence: context.GuardianEvidence, Scope: scopeResult{Decision: context.ScopeDecision, Status: statusFor(context.ScopeDecision)}, Fixtures: fixtureResult{Paths: context.FixturePaths, Status: context.FixtureStatus, Source: context.SourceStatus, Semantic: context.SemanticStatus, Provenance: context.ProvenanceStatus}, Artifacts: context.Artifacts, Cache: context.Cache, Digests: digests, WriteEffect: context.WriteEffect, Decision: "PASS", NoWrite: context.NoWrite, Rejections: rejections, Predecessors: context.Predecessors, MissingReasons: context.MissingReasons}
	if len(rejections) > 0 {
		bundle.Decision = "FAIL_CLOSED"
	}
	bundle.PromotionObservation = context.PromotionObservation
	bundle.PromotionAuthorization = promotionAuthorizationFor(bundle)
	payload, err := marshalProof(bundle)
	if err != nil {
		return proofBundle{}, provenanceReceipt{}, err
	}
	bundle.Digests.Bundle = digestBytes(payload)
	if bundle.PromotionAuthorization != nil {
		bundle.PromotionAuthorization.ProofDigest = bundle.Digests.Bundle
	}
	receipt := makeReceipt(bundle, context)
	return bundle, receipt, nil
}

func gateRejections(inputs proofInputs) []string {
	c := inputs.Context
	failures := make([]string, 0)
	if c.ScopeDecision != "passed" {
		failures = append(failures, "scope_not_passed")
	}
	if c.BaseRef == "main" {
		if c.GuardianEvidence == nil || c.BranchProtection.ReadStatus != "verified" || !branchProtectionReadyFor(c.BranchProtection, "main") || c.DevBranchProtection.ReadStatus != "verified" || !branchProtectionReadyFor(c.DevBranchProtection, "dev") {
			failures = append(failures, "main_protection_not_verified")
		}
		if !validPromotionObservationForContext(c) {
			failures = append(failures, "promotion_observation_not_verified")
		}
	}
	for status, value := range map[string]string{"fixture": c.FixtureStatus, "source": c.SourceStatus, "semantic": c.SemanticStatus, "provenance": c.ProvenanceStatus} {
		if value != "verified" {
			failures = append(failures, status+"_evidence_not_verified")
		}
	}
	if c.ArtifactsStatus != "verified" {
		failures = append(failures, "artifact_evidence_not_verified")
	}
	if c.WriteEffect != "none" || !c.NoWrite {
		failures = append(failures, "write_effect_not_none")
	}
	if len(c.FixturePaths) == 0 {
		failures = append(failures, "fixture_inventory_missing")
	}
	if len(c.Artifacts) == 0 {
		failures = append(failures, "artifact_inventory_missing")
	}
	if err := validateCache(c.Cache, inputs.Evidence); err != nil {
		failures = append(failures, "cache_"+err.Error())
	}
	sort.Strings(failures)
	unique := failures[:0]
	for _, failure := range failures {
		if len(unique) == 0 || unique[len(unique)-1] != failure {
			unique = append(unique, failure)
		}
	}
	return unique
}

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
	return provenanceReceipt{Schema: receiptSchema, Operation: "verify", Relation: "conformance", Delta: "ci-policy", AllowedIntent: "verification-only", Locality: "repository", Repository: bundle.Repository, Event: bundle.Event, BaseRef: bundle.BaseRef, BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA, Ref: bundle.Ref, EventRef: bundle.EventRef, CheckoutRef: bundle.CheckoutRef, PRNumber: bundle.PRNumber, RunID: bundle.RunID, RunAttempt: bundle.RunAttempt, WorkflowSHA: bundle.WorkflowSHA, BranchProtection: bundle.BranchProtection, DevBranchProtection: bundle.DevBranchProtection, DomainEvidence: bundle.DomainEvidence, GuardianEvidence: bundle.GuardianEvidence, PromotionObservation: bundle.PromotionObservation, PromotionAuthorization: bundle.PromotionAuthorization, Scheduler: bundle.Scheduler, Jobs: bundle.Jobs, Artifacts: bundle.Artifacts, Digests: receiptDigests{Source: bundle.Digests.Source, IR: bundle.Digests.Semantic, Projection: bundle.Digests.Projection, Build: bundle.Digests.Build, Policy: bundle.Digests.Policy, Schema: bundle.Digests.Schema, Toolchain: bundle.Digests.Toolchain, Target: bundle.Digests.Target, Bundle: bundle.Digests.Bundle}, Cache: cacheReceipt{Key: bundle.Cache.Key, Outcome: bundle.Cache.Outcome}, DiagnosticIDs: context.DiagnosticIDs, RepairIDs: context.RepairIDs, WriteEffect: bundle.WriteEffect, Producer: "go-ci-proof", Role: "Gate", Predecessors: context.Predecessors, Decision: bundle.Decision, MissingReasons: bundle.MissingReasons}
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

func validateCache(cache cacheInput, evidence evidenceInput) error {
	if cache.Status == "not_applicable" {
		if cache.Key != "none" || cache.Outcome != "not_run" {
			return fmt.Errorf("not_applicable_values_invalid")
		}
		return nil
	}
	if cache.Status != "verified" || cache.Key == "" || cache.Outcome != "hit" && cache.Outcome != "miss" {
		return fmt.Errorf("receipt_missing")
	}
	expectedKey := digestBytes([]byte(cache.ArtifactKind + "\x00" + cache.SemanticClosureDigest + "\x00" + cache.DependencyRoot))
	if cache.Key != expectedKey || !validDigest(cache.SemanticClosureDigest) || !validDigest(cache.PolicyDigest) || !validDigest(cache.SchemaDigest) || !validDigest(cache.ToolchainDigest) || !validDigest(cache.TargetDigest) || cache.DependencyRoot == "" || cache.ArtifactKind == "" || cache.ContentSize != cache.HitContentSize || len(cache.DirectDependencies) == 0 || len(cache.EvidenceRefs) == 0 || cache.ProducerHost == "" || cache.Predecessor == evidence.HeadSHA {
		return fmt.Errorf("contract_mismatch")
	}
	return nil
}

func verifyReceipt(filename string, bundle proofBundle) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 || lines[len(lines)-1] == "" {
		return fmt.Errorf("receipt is empty")
	}
	var receipt provenanceReceipt
	decoder := json.NewDecoder(strings.NewReader(lines[len(lines)-1]))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return err
	}
	expectedDigests := receiptDigests{Source: bundle.Digests.Source, IR: bundle.Digests.Semantic, Projection: bundle.Digests.Projection, Build: bundle.Digests.Build, Policy: bundle.Digests.Policy, Schema: bundle.Digests.Schema, Toolchain: bundle.Digests.Toolchain, Target: bundle.Digests.Target, Bundle: bundle.Digests.Bundle}
	if receipt.Schema != receiptSchema || receipt.Operation != "verify" || receipt.Relation != "conformance" || receipt.AllowedIntent != "verification-only" || receipt.Repository != bundle.Repository || receipt.Event != bundle.Event || receipt.BaseRef != bundle.BaseRef || receipt.BaseSHA != bundle.BaseSHA || receipt.HeadSHA != bundle.HeadSHA || receipt.Ref != bundle.Ref || receipt.EventRef != bundle.EventRef || receipt.CheckoutRef != bundle.CheckoutRef || receipt.PRNumber != bundle.PRNumber || receipt.RunID != bundle.RunID || receipt.RunAttempt != bundle.RunAttempt || receipt.WorkflowSHA != bundle.WorkflowSHA || receipt.Decision != bundle.Decision || !reflect.DeepEqual(receipt.BranchProtection, bundle.BranchProtection) || !reflect.DeepEqual(receipt.DevBranchProtection, bundle.DevBranchProtection) || !reflect.DeepEqual(receipt.DomainEvidence, bundle.DomainEvidence) || !reflect.DeepEqual(receipt.GuardianEvidence, bundle.GuardianEvidence) || !reflect.DeepEqual(receipt.PromotionObservation, bundle.PromotionObservation) || !reflect.DeepEqual(receipt.PromotionAuthorization, bundle.PromotionAuthorization) || !reflect.DeepEqual(receipt.Scheduler, bundle.Scheduler) || !reflect.DeepEqual(receipt.Jobs, bundle.Jobs) || receipt.Digests != expectedDigests || receipt.MissingReasons != bundle.MissingReasons || len(receipt.Artifacts) != len(bundle.Artifacts) || receipt.WriteEffect != "none" || receipt.Cache.Key != bundle.Cache.Key || receipt.Cache.Outcome != bundle.Cache.Outcome || !reflect.DeepEqual(receipt.Predecessors, bundle.Predecessors) {
		return fmt.Errorf("provenance receipt does not match proof bundle")
	}
	for index, artifact := range bundle.Artifacts {
		recorded := receipt.Artifacts[index]
		if recorded.ID != artifact.ID || recorded.Name != artifact.Name || recorded.Size != artifact.Size || recorded.Expired != artifact.Expired || recorded.Digest != artifact.Digest || recorded.RunID != artifact.RunID || recorded.RunAttempt != artifact.RunAttempt {
			return fmt.Errorf("provenance receipt artifact inventory mismatch")
		}
	}
	return nil
}
