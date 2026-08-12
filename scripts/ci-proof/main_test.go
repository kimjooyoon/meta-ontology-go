package main

import (
	"encoding/json"
	"os"
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
	if err := validateArtifacts([]artifactInput{{ID: 1, Name: "empty", Size: 0, Digest: strings.Repeat("a", 64)}}); err == nil {
		t.Fatal("zero-sized artifact was accepted")
	}
}

func TestCIArtifactDigestMissingFailsClosed(t *testing.T) {
	if err := validateArtifacts([]artifactInput{{ID: 1, Name: "receipt", Size: 1}}); err == nil {
		t.Fatal("artifact without digest was accepted")
	}
}

func TestCIBranchProtectionSnapshotMismatchFailsClosed(t *testing.T) {
	bundle := validProof()
	bundle.BranchProtection.Digest = "mismatch"
	if err := validateProof(bundle); err == nil {
		t.Fatal("unbound branch protection snapshot was accepted")
	}
}

func TestCIBranchProtectionSnapshotRefTamperingFailsClosed(t *testing.T) {
	for _, mutate := range []func(*branchProtection){
		func(snapshot *branchProtection) { snapshot.EventRef = "refs/pull/2/merge" },
		func(snapshot *branchProtection) { snapshot.CheckoutRef = strings.Repeat("b", 40) },
	} {
		bundle := validProof()
		mutate(&bundle.BranchProtection)
		if err := validateProof(bundle); err == nil {
			t.Fatal("tampered branch protection ref was accepted")
		}
	}
}

func TestCIBranchProtectionUnavailableIsNotReady(t *testing.T) {
	bundle := validProof()
	bundle.BranchProtection.Exists = false
	bundle.BranchProtection.ReadStatus = "unavailable"
	if branchProtectionReady(bundle.BranchProtection) {
		t.Fatal("unavailable branch protection snapshot was promotion-ready")
	}
}

func TestCIDomainEvidenceOutputTamperingFailsClosed(t *testing.T) {
	bundle := validProof()
	bundle.DomainEvidence.CLI.Output += "tampered"
	if err := validateProof(bundle); err == nil {
		t.Fatal("tampered CLI domain evidence was accepted")
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

func TestCITerminalJobSnapshotRejectsInProgress(t *testing.T) {
	jobs := make([]jobInput, len(proofJobs))
	head := strings.Repeat("a", 40)
	for index, name := range proofJobs {
		jobs[index] = jobInput{ID: int64(index + 1), Name: name, Status: "completed", Conclusion: "success", HeadSHA: head}
	}
	jobs[len(jobs)-1].Status = "in_progress"
	data, err := json.Marshal(jobs)
	if err != nil {
		t.Fatal(err)
	}
	filename := t.TempDir() + "/jobs.json"
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readJobs(filename); err == nil {
		t.Fatal("in-progress canonical job was accepted")
	}
}

func validProof() proofBundle {
	head := strings.Repeat("a", 40)
	jobs := make([]jobInput, len(proofJobs))
	for index, name := range proofJobs {
		jobs[index] = jobInput{ID: int64(index + 1), Name: name, Conclusion: "success", HeadSHA: head}
	}
	bundle := proofBundle{Schema: proofSchema, Repository: "owner/repo", Event: "pull_request", PRNumber: 1, BaseRef: "integration", BaseSHA: strings.Repeat("b", 40), HeadSHA: head, Ref: "refs/pull/1/merge", EventRef: "refs/pull/1/merge", CheckoutRef: head, RunID: 1, RunAttempt: 1, WorkflowSHA: strings.Repeat("c", 40), Jobs: jobs, Actors: actorRoles{Actor: "builder", Builder: "builder", Guardian: "guardian", Approver: "approver", Gate: "CI policy"}, Scope: scopeResult{Decision: "passed", Status: "verified"}, Fixtures: fixtureResult{Paths: []string{"examples/billing/main.gooo"}, Status: "verified", Source: "verified", Semantic: "verified", Provenance: "verified"}, Artifacts: []artifactInput{{ID: 1, Name: "receipt", Size: 1, Digest: strings.Repeat("a", 64)}}, Cache: cacheInput{Key: "none", Outcome: "not_run", Status: "not_applicable"}, Digests: proofDigests{Source: strings.Repeat("1", 64), Semantic: strings.Repeat("2", 64), Provenance: strings.Repeat("3", 64), Projection: strings.Repeat("4", 64), Build: strings.Repeat("5", 64), Policy: strings.Repeat("6", 64), Schema: strings.Repeat("7", 64), Toolchain: strings.Repeat("8", 64), Target: strings.Repeat("9", 64)}, WriteEffect: "none", Decision: "PASS", NoWrite: true}
	bundle.BranchProtection = validBranchProtection(bundle)
	bundle.DomainEvidence = validDomainEvidence(bundle)
	payload, _ := json.Marshal(bundle)
	bundle.Digests.Bundle = digestBytes(payload)
	return bundle
}

func validDomainEvidence(bundle proofBundle) domainEvidence {
	domain := domainEvidence{Schema: "gooo/domain-evidence/v1", Repository: bundle.Repository, Event: bundle.Event, BaseRef: bundle.BaseRef, BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA, EventRef: bundle.EventRef, CheckoutRef: bundle.CheckoutRef, RunID: bundle.RunID, RunAttempt: bundle.RunAttempt, WorkflowSHA: bundle.WorkflowSHA, CLI: domainCommand{Command: "go run ./cmd/gooo check examples/billing/main.gooo", Fixture: "examples/billing/main.gooo", Status: "verified", Available: true, Output: "ok: examples/billing/main.gooo\n"}, Graph: domainCommand{Command: "go run ./cmd/gooo graph-dump examples/billing/main.gooo", Status: "deferred", Available: false}, ObserverStatus: "unavailable", ProtectionStatus: "unavailable", ApprovalStatus: "unavailable", ProvenanceStatus: "unavailable", Digests: domainEvidenceDigest{SourceSHA256: bundle.Digests.Source, IRSHA256: bundle.Digests.Semantic, GeneratedSHA256: bundle.Digests.Projection, BundleSHA256: strings.Repeat("a", 64)}}
	domain.CLI.OutputSHA256 = digestBytes([]byte(domain.CLI.Output))
	domain.Digests.DomainSHA256 = digestDomainEvidence(domain)
	return domain
}

func validBranchProtection(bundle proofBundle) branchProtection {
	protection := branchProtection{Repository: bundle.Repository, Branch: bundle.BaseRef, PolicySHA: bundle.Digests.Policy, EventRef: bundle.EventRef, CheckoutRef: bundle.CheckoutRef, TokenSource: "github.token", ReadStatus: "verified", Exists: true, Strict: true, RequiredChecks: append([]string(nil), proofJobs...), EnforceAdmins: true, RequiredReviews: 1, DismissStaleReviews: true, RequireLastPushApproval: true, LinearHistory: true, BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA, RunID: bundle.RunID, RunAttempt: bundle.RunAttempt, WorkflowSHA: bundle.WorkflowSHA}
	protection.Digest = digestBranchProtection(protection)
	return protection
}

func validCache() cacheInput {
	semantic := strings.Repeat("1", 64)
	root := "dependency-root"
	return cacheInput{Key: digestBytes([]byte("artifact\x00" + semantic + "\x00" + root)), Outcome: "hit", Status: "verified", ArtifactKind: "artifact", SemanticClosureDigest: semantic, DependencyRoot: root, DirectDependencies: []string{"dep"}, PolicyDigest: strings.Repeat("2", 64), SchemaDigest: strings.Repeat("3", 64), ToolchainDigest: strings.Repeat("4", 64), TargetDigest: strings.Repeat("5", 64), EvidenceRefs: []string{"receipt"}, ProducerHost: "host", ContentSize: 1, HitContentSize: 1, Predecessor: strings.Repeat("b", 40)}
}
