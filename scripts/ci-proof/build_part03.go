package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

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
	if receipt.Schema != receiptSchema || receipt.Operation != "verify" || receipt.Relation != "conformance" || receipt.AllowedIntent != "verification-only" || receipt.Repository != bundle.Repository || receipt.Event != bundle.Event || receipt.BaseRef != bundle.BaseRef || receipt.BaseSHA != bundle.BaseSHA || receipt.HeadSHA != bundle.HeadSHA || receipt.Ref != bundle.Ref || receipt.EventRef != bundle.EventRef || receipt.CheckoutRef != bundle.CheckoutRef || receipt.PRNumber != bundle.PRNumber || receipt.RunID != bundle.RunID || receipt.RunAttempt != bundle.RunAttempt || receipt.WorkflowSHA != bundle.WorkflowSHA || receipt.Decision != bundle.Decision || !reflect.DeepEqual(receipt.BranchProtection, bundle.BranchProtection) || !reflect.DeepEqual(receipt.DevBranchProtection, bundle.DevBranchProtection) || !reflect.DeepEqual(receipt.DomainEvidence, bundle.DomainEvidence) || !reflect.DeepEqual(receipt.ArtifactProvenance, bundle.ArtifactProvenance) || !reflect.DeepEqual(receipt.GuardianEvidence, bundle.GuardianEvidence) || !reflect.DeepEqual(receipt.FoundationPromotion, bundle.FoundationPromotion) || !reflect.DeepEqual(receipt.PromotionObservation, bundle.PromotionObservation) || !reflect.DeepEqual(receipt.PromotionAuthorization, bundle.PromotionAuthorization) || !reflect.DeepEqual(receipt.Jobs, bundle.Jobs) || receipt.Digests != expectedDigests || receipt.MissingReasons != bundle.MissingReasons || len(receipt.Artifacts) != len(bundle.Artifacts) || receipt.WriteEffect != "none" || receipt.Cache.Key != bundle.Cache.Key || receipt.Cache.Outcome != bundle.Cache.Outcome || !reflect.DeepEqual(receipt.Predecessors, bundle.Predecessors) {
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
