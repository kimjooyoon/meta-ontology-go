package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

func buildProof(inputs proofInputs, digests proofDigests) (proofBundle, provenanceReceipt, error) {
	context := inputs.Context
	rejections := gateRejections(inputs)
	bundle := proofBundle{Schema: proofSchema, Repository: context.Repository, Event: context.Event, PRNumber: context.PRNumber, BaseRef: context.BaseRef, BaseSHA: context.BaseSHA, HeadSHA: context.HeadSHA, Ref: context.Ref, RunID: context.RunID, RunAttempt: context.RunAttempt, WorkflowSHA: context.WorkflowSHA, Jobs: inputs.Jobs, Actors: actorRoles{Actor: context.Actor, Builder: context.Builder, Guardian: context.Guardian, Approver: context.Approver, Gate: context.Gate}, Scope: scopeResult{Decision: context.ScopeDecision, Status: statusFor(context.ScopeDecision)}, Fixtures: fixtureResult{Paths: context.FixturePaths, Status: context.FixtureStatus, Source: context.SourceStatus, Semantic: context.SemanticStatus, Provenance: context.ProvenanceStatus}, Artifacts: context.Artifacts, Approvals: context.Approvals, Cache: context.Cache, Digests: digests, WriteEffect: context.WriteEffect, Decision: "PASS", NoWrite: context.NoWrite, Rejections: rejections, Predecessors: context.Predecessors}
	if len(rejections) > 0 {
		bundle.Decision = "FAIL_CLOSED"
	}
	payload, err := marshalProof(bundle)
	if err != nil {
		return proofBundle{}, provenanceReceipt{}, err
	}
	bundle.Digests.Bundle = digestBytes(payload)
	receipt := makeReceipt(bundle, context)
	return bundle, receipt, nil
}

func gateRejections(inputs proofInputs) []string {
	c := inputs.Context
	failures := make([]string, 0)
	if inputs.Governance.Promotion.BranchProtectionRequired && !c.BranchProtected {
		failures = append(failures, "branch_protection_missing")
	}
	if c.Guardian == "" || c.Approver == "" || c.Guardian == c.Approver || c.Guardian == c.Builder || c.Approver == c.Builder || c.Guardian == c.Actor || c.Approver == c.Actor {
		failures = append(failures, "independent_approval_missing_or_overlapping")
	}
	if c.ScopeDecision != "passed" {
		failures = append(failures, "scope_not_passed")
	}
	for status, value := range map[string]string{"fixture": c.FixtureStatus, "source": c.SourceStatus, "semantic": c.SemanticStatus, "provenance": c.ProvenanceStatus, "artifacts": c.ArtifactsStatus, "approvals": c.ApprovalsStatus} {
		if value != "verified" {
			failures = append(failures, status+"_evidence_not_verified")
		}
	}
	if c.WriteEffect != "none" || !c.NoWrite {
		failures = append(failures, "write_effect_not_none")
	}
	if len(c.FixturePaths) == 0 || len(c.Artifacts) == 0 {
		failures = append(failures, "fixture_or_artifact_inventory_missing")
	}
	if err := validateCache(c.Cache, inputs.Evidence); err != nil {
		failures = append(failures, "cache_"+err.Error())
	}
	sort.Strings(failures)
	return failures
}

func statusFor(decision string) string {
	if decision == "passed" {
		return "verified"
	}
	return "rejected"
}

func makeReceipt(bundle proofBundle, context contextInput) provenanceReceipt {
	return provenanceReceipt{Schema: receiptSchema, Operation: "verify", Relation: "conformance", Delta: "ci-policy", AllowedIntent: "verification-only", Locality: "repository", Event: bundle.Event, BaseRef: bundle.BaseRef, BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA, Ref: bundle.Ref, PRNumber: bundle.PRNumber, RunID: bundle.RunID, RunAttempt: bundle.RunAttempt, WorkflowSHA: bundle.WorkflowSHA, Jobs: bundle.Jobs, Digests: receiptDigests{Source: bundle.Digests.Source, IR: bundle.Digests.Semantic, Projection: bundle.Digests.Projection, Build: bundle.Digests.Build, Policy: bundle.Digests.Policy, Schema: bundle.Digests.Schema, Toolchain: bundle.Digests.Toolchain, Target: bundle.Digests.Target, Bundle: bundle.Digests.Bundle}, Cache: cacheReceipt{Key: bundle.Cache.Key, Outcome: bundle.Cache.Outcome}, DiagnosticIDs: context.DiagnosticIDs, RepairIDs: context.RepairIDs, WriteEffect: bundle.WriteEffect, Producer: "go-ci-proof", Role: "Gate", Predecessors: context.Predecessors, Decision: bundle.Decision}
}

func marshalProof(bundle proofBundle) ([]byte, error) {
	bundle.Digests.Bundle = ""
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
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &receipt); err != nil {
		return err
	}
	if receipt.Schema != receiptSchema || receipt.Operation != "verify" || receipt.Relation != "conformance" || receipt.AllowedIntent != "verification-only" || receipt.HeadSHA != bundle.HeadSHA || receipt.RunID != bundle.RunID || receipt.Decision != bundle.Decision || len(receipt.Jobs) != len(proofJobs) || receipt.WriteEffect != "none" || receipt.Cache.Key != bundle.Cache.Key || receipt.Cache.Outcome != bundle.Cache.Outcome {
		return fmt.Errorf("provenance receipt does not match proof bundle")
	}
	return nil
}
