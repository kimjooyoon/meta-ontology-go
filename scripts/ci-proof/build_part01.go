package main

import (
	"sort"
)

func buildProof(inputs proofInputs, digests proofDigests) (proofBundle, provenanceReceipt, error) {
	context := inputs.Context
	rejections := gateRejections(inputs)
	bundle := proofBundle{Schema: proofSchema, Repository: context.Repository, Event: context.Event, PRNumber: context.PRNumber, BaseRef: context.BaseRef, BaseSHA: context.BaseSHA, HeadSHA: context.HeadSHA, Ref: context.Ref, EventRef: context.EventRef, CheckoutRef: context.CheckoutRef, RunID: context.RunID, RunAttempt: context.RunAttempt, WorkflowSHA: context.WorkflowSHA, Jobs: inputs.Jobs, Actors: actorRoles{Actor: context.Actor, Builder: context.Builder, Gate: context.Gate}, BranchProtection: context.BranchProtection, DevBranchProtection: context.DevBranchProtection, DomainEvidence: context.DomainEvidence, GuardianEvidence: context.GuardianEvidence, Scope: scopeResult{Decision: context.ScopeDecision, Status: statusFor(context.ScopeDecision)}, Fixtures: fixtureResult{Paths: context.FixturePaths, Status: context.FixtureStatus, Source: context.SourceStatus, Semantic: context.SemanticStatus, Provenance: context.ProvenanceStatus}, Artifacts: context.Artifacts, Cache: context.Cache, Digests: digests, WriteEffect: context.WriteEffect, Decision: "PASS", NoWrite: context.NoWrite, Rejections: rejections, Predecessors: context.Predecessors, MissingReasons: context.MissingReasons}
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
