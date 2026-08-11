package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProofBundleValidatesAndPreservesReceiptSchema(t *testing.T) {
	bundle := validProof()
	if err := validateProof(bundle); err != nil {
		t.Fatal(err)
	}
	receipt := makeReceipt(bundle, contextInput{})
	if receipt.Schema != "gooo/provenance-receipt/v1" || receipt.Relation != "conformance" {
		t.Fatal("receipt schema or relation changed")
	}
}

func TestCICacheC1KeyMutationFailsClosed(t *testing.T) {
	cache := validCache()
	cache.Key = "mutated"
	if err := validateCache(cache, evidenceInput{HeadSHA: strings.Repeat("a", 40)}); err == nil {
		t.Fatal("cache key mutation was accepted")
	}
}

func TestCICacheC2ContentSizeMismatchFailsClosed(t *testing.T) {
	cache := validCache()
	cache.HitContentSize++
	if err := validateCache(cache, evidenceInput{HeadSHA: strings.Repeat("a", 40)}); err == nil {
		t.Fatal("cache content-size mismatch was accepted")
	}
}

func TestCICacheC3UnknownDependencyFailsClosed(t *testing.T) {
	cache := validCache()
	cache.DirectDependencies = nil
	if err := validateCache(cache, evidenceInput{HeadSHA: strings.Repeat("a", 40)}); err == nil {
		t.Fatal("unknown dependency evidence was accepted")
	}
}

func TestCICacheC4ReplayPredecessorFailsClosed(t *testing.T) {
	cache := validCache()
	cache.Predecessor = strings.Repeat("a", 40)
	if err := validateCache(cache, evidenceInput{HeadSHA: strings.Repeat("a", 40)}); err == nil {
		t.Fatal("replayed predecessor was accepted")
	}
}

func TestCICacheC5MissingArtifactFailsClosed(t *testing.T) {
	if err := validateArtifacts([]artifactInput{{ID: 1, Name: "empty", Size: 0}}); err == nil {
		t.Fatal("zero-sized artifact was accepted")
	}
}

func validProof() proofBundle {
	head := strings.Repeat("a", 40)
	jobs := make([]jobInput, len(proofJobs))
	for index, name := range proofJobs {
		jobs[index] = jobInput{ID: int64(index + 1), Name: name, Conclusion: "success", HeadSHA: head}
	}
	bundle := proofBundle{Schema: proofSchema, Repository: "owner/repo", Event: "pull_request", PRNumber: 1, BaseRef: "integration", BaseSHA: strings.Repeat("b", 40), HeadSHA: head, Ref: "refs/pull/1/merge", RunID: 1, RunAttempt: 1, WorkflowSHA: strings.Repeat("c", 40), Jobs: jobs, Actors: actorRoles{Actor: "builder", Builder: "builder", Guardian: "guardian", Approver: "approver", Gate: "CI policy"}, Scope: scopeResult{Decision: "passed", Status: "verified"}, Fixtures: fixtureResult{Paths: []string{"examples/billing/main.gooo"}, Status: "verified", Source: "verified", Semantic: "verified", Provenance: "verified"}, Artifacts: []artifactInput{{ID: 1, Name: "receipt", Size: 1}}, Cache: cacheInput{Key: "none", Outcome: "not_run", Status: "not_applicable"}, Digests: proofDigests{Source: strings.Repeat("1", 64), Semantic: strings.Repeat("2", 64), Provenance: strings.Repeat("3", 64), Projection: strings.Repeat("4", 64), Build: strings.Repeat("5", 64), Policy: strings.Repeat("6", 64), Schema: strings.Repeat("7", 64), Toolchain: strings.Repeat("8", 64), Target: strings.Repeat("9", 64)}, WriteEffect: "none", Decision: "PASS", NoWrite: true}
	payload, _ := json.Marshal(bundle)
	bundle.Digests.Bundle = digestBytes(payload)
	return bundle
}

func validCache() cacheInput {
	semantic := strings.Repeat("1", 64)
	root := "dependency-root"
	return cacheInput{Key: digestBytes([]byte("artifact\x00" + semantic + "\x00" + root)), Outcome: "hit", Status: "verified", ArtifactKind: "artifact", SemanticClosureDigest: semantic, DependencyRoot: root, DirectDependencies: []string{"dep"}, PolicyDigest: strings.Repeat("2", 64), SchemaDigest: strings.Repeat("3", 64), ToolchainDigest: strings.Repeat("4", 64), TargetDigest: strings.Repeat("5", 64), EvidenceRefs: []string{"receipt"}, ProducerHost: "host", ContentSize: 1, HitContentSize: 1, Predecessor: strings.Repeat("b", 40)}
}
