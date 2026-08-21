package main

import (
	"encoding/json"
	"strings"
	"time"
)

func validProof() proofBundle {
	head := strings.Repeat("a", 40)
	jobs := make([]jobInput, len(proofJobs))
	for index, name := range proofJobs {
		jobs[index] = jobInput{ID: int64(index + 1), Name: name, Status: "completed", Conclusion: "success", HeadSHA: head, RunID: 1, RunAttempt: 1}
	}
	bundle := proofBundle{Schema: proofSchema, Repository: "owner/repo", Event: "pull_request", PRNumber: 1, BaseRef: "dev", BaseSHA: strings.Repeat("b", 40), HeadSHA: head, Ref: "refs/pull/1/merge", EventRef: "refs/pull/1/merge", CheckoutRef: head, RunID: 1, RunAttempt: 1, WorkflowSHA: strings.Repeat("c", 40), Jobs: jobs, Actors: actorRoles{Actor: "builder", Builder: "builder", Gate: "CI policy"}, ArtifactProvenance: artifactProvenanceFixture(strings.Repeat("b", 40), head), Scope: scopeResult{Decision: "passed", Status: "verified"}, Fixtures: fixtureResult{Paths: []string{"examples/billing/main.gooo"}, Status: "verified", Source: "verified", Semantic: "verified", Provenance: "verified"}, Artifacts: []artifactInput{{ID: 1, Name: "ci-evidence-1-1", Size: 1, Digest: "sha256:" + strings.Repeat("a", 64), RunID: 1, RunAttempt: 1}}, Cache: cacheInput{Key: "none", Outcome: "not_run", Status: "not_applicable"}, Digests: proofDigests{Source: strings.Repeat("1", 64), Semantic: strings.Repeat("2", 64), Provenance: strings.Repeat("3", 64), Projection: strings.Repeat("4", 64), Build: strings.Repeat("5", 64), Policy: strings.Repeat("6", 64), Schema: strings.Repeat("7", 64), Toolchain: strings.Repeat("8", 64), Target: strings.Repeat("9", 64)}, WriteEffect: "none", Decision: "PASS", NoWrite: true, MissingReasons: missingReasons{Protection: "domain_protection_observer_unavailable"}}
	bundle.BranchProtection = unobservedBranchProtection(bundle)
	bundle.DomainEvidence = validDomainEvidence(bundle)
	payload, _ := json.Marshal(bundle)
	bundle.Digests.Bundle = digestBytes(payload)
	return bundle
}

func validDomainEvidence(bundle proofBundle) domainEvidence {
	domain := domainEvidence{Schema: domainEvidenceSchema, Repository: bundle.Repository, Event: bundle.Event, BaseRef: bundle.BaseRef, BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA, EventRef: bundle.EventRef, CheckoutRef: bundle.CheckoutRef, RunID: bundle.RunID, RunAttempt: bundle.RunAttempt, WorkflowSHA: bundle.WorkflowSHA, CLI: domainCommand{Command: "go run ./cmd/gooo check examples/billing/main.gooo", Fixture: "examples/billing/main.gooo", Status: "verified", Available: true, Output: "ok: examples/billing/main.gooo\n"}, Graph: domainCommand{Command: "go run ./cmd/gooo graph-dump examples/billing/main.gooo", Status: "deferred", Available: false}, ObserverStatus: "unavailable", ProtectionStatus: "unavailable", ProvenanceStatus: "not_applicable", MissingReasons: bundle.MissingReasons, Digests: domainEvidenceDigest{SourceSHA256: bundle.Digests.Source, IRSHA256: bundle.Digests.Semantic, GeneratedSHA256: bundle.Digests.Projection, BundleSHA256: strings.Repeat("a", 64)}}
	domain.CLI.OutputSHA256 = digestBytes([]byte(domain.CLI.Output))
	domain.Digests.DomainSHA256 = digestDomainEvidence(domain)
	return domain
}

func validBranchProtection(bundle proofBundle) branchProtection {
	tokenSource := "not_observed"
	if bundle.BaseRef == "main" {
		tokenSource = "github_app_installation"
	}
	observedAt, validUntil := freshObserverWindow()
	protection := branchProtection{Repository: bundle.Repository, Branch: bundle.BaseRef, PolicySHA: bundle.Digests.Policy, EventRef: bundle.EventRef, CheckoutRef: bundle.CheckoutRef, TokenSource: tokenSource, AppInstallationID: 42, AppSlug: "guardian", ReadStatus: "verified", Exists: true, Strict: true, RequiredChecks: append([]string(nil), proofJobs...), RequiredCheckBindings: requiredCheckBindingsFor(proofJobs), EnforceAdmins: true, RequiredReviews: 0, DismissStaleReviews: false, RequireLastPushApproval: false, LinearHistory: true, BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA, RunID: bundle.RunID, RunAttempt: bundle.RunAttempt, WorkflowSHA: bundle.WorkflowSHA, ObservedAt: observedAt, ValidUntil: validUntil}
	if bundle.BaseRef != "main" {
		protection.ReadStatus = "unavailable"
		protection.Exists = false
		protection.Strict = false
		protection.RequiredChecks = nil
		protection.RequiredCheckBindings = nil
		protection.EnforceAdmins = false
		protection.LinearHistory = false
		protection.ObservedAt = nil
		protection.ValidUntil = nil
		protection.MissingReason = "trusted_guardian_required"
	}
	protection.Digest = digestBranchProtection(protection)
	return protection
}

func freshObserverWindow() (*string, *string) {
	observed := time.Now().UTC().Add(-time.Minute)
	validUntil := observed.Add(guardianObserverFreshnessWindow)
	return new(observed.Format(time.RFC3339Nano)), new(validUntil.Format(time.RFC3339Nano))
}

func unobservedBranchProtection(bundle proofBundle) branchProtection {
	return validBranchProtection(bundle)
}

func validCache() cacheInput {
	semantic := strings.Repeat("1", 64)
	root := "dependency-root"
	return cacheInput{Key: digestBytes([]byte("artifact\x00" + semantic + "\x00" + root)), Outcome: "hit", Status: "verified", ArtifactKind: "artifact", SemanticClosureDigest: semantic, DependencyRoot: root, DirectDependencies: []string{"dep"}, PolicyDigest: strings.Repeat("2", 64), SchemaDigest: strings.Repeat("3", 64), ToolchainDigest: strings.Repeat("4", 64), TargetDigest: strings.Repeat("5", 64), EvidenceRefs: []string{"receipt"}, ProducerHost: "host", ContentSize: 1, HitContentSize: 1, Predecessor: strings.Repeat("b", 40)}
}
