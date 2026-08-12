package cache

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"
)

func cacheReceiptFixture(key Key, run string) CacheReceipt {
	evidence := evidenceFixture(run)
	return CacheReceipt{
		SchemaVersion: cacheReceiptSchemaVersion, CacheKey: key.Digest, Domain: key.Domain,
		KeyVersion: key.Version, HostStage: key.HostStage, ArtifactKind: key.ArtifactKind,
		Projection:            key.Projection,
		SemanticClosureDigest: key.SemanticClosureDigest, DependencyRoot: key.DependencyRoot,
		DirectDependencies: []Digest{HashBytes([]byte("direct"))}, PolicySchemaDigest: key.PolicySchemaDigest,
		Toolchain: key.Toolchain, Target: key.Target, BuildTagsDigest: key.BuildTagsDigest,
		OptionsDigest: key.OptionsDigest,
		ContentDigest: HashBytes([]byte("projection")), Size: int64(len("projection")), Reconstructable: true,
		EvidenceRefs: evidence.EvidenceRefs, ProducerHost: "go-hosted", Status: ReceiptRecomputed,
		Evidence: evidence,
	}
}

func TestCacheReceiptCanonicalizationIsPresentationStable(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	left := cacheReceiptFixture(key, "canonical")
	right := cacheReceiptFixture(key, "canonical")
	right.EvidenceRefs = append([]EvidenceRef(nil), right.EvidenceRefs...)
	slices.Reverse(right.Evidence.PredecessorDigests)
	slices.Reverse(right.Evidence.EvidenceRefs)
	slices.Reverse(right.EvidenceRefs)
	sealedLeft, err := left.Seal()
	if err != nil {
		t.Fatal(err)
	}
	sealedRight, err := right.Seal()
	if err != nil {
		t.Fatal(err)
	}
	if sealedLeft.ReceiptDigest != sealedRight.ReceiptDigest {
		t.Fatalf("presentation changed receipt digest: %s != %s", sealedLeft.ReceiptDigest, sealedRight.ReceiptDigest)
	}
	leftJSON, err := json.Marshal(sealedLeft)
	if err != nil {
		t.Fatal(err)
	}
	rightJSON, err := json.Marshal(sealedRight)
	if err != nil {
		t.Fatal(err)
	}
	if string(leftJSON) != string(rightJSON) {
		t.Fatal("canonical receipt JSON differs by presentation order")
	}
}

func TestCacheReceiptBindsProjectionAndContent(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	receipt := cacheReceiptFixture(key, "binding")
	if err := receipt.ValidateForKey(key); err != nil {
		t.Fatal(err)
	}
	if err := receipt.ValidateForData([]byte("projection")); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*CacheReceipt){
		"domain":     func(r *CacheReceipt) { r.Domain = "other" },
		"version":    func(r *CacheReceipt) { r.KeyVersion = "other" },
		"host stage": func(r *CacheReceipt) { r.HostStage = GoooHostedStage },
		"projection": func(r *CacheReceipt) { r.Projection = "other" },
		"artifact":   func(r *CacheReceipt) { r.ArtifactKind = "other" },
		"options":    func(r *CacheReceipt) { r.OptionsDigest = mustOptionsDigest(map[string]any{"mode": "other"}) },
		"content":    func(r *CacheReceipt) { r.ContentDigest = HashBytes([]byte("other")) },
		"size":       func(r *CacheReceipt) { r.Size++ },
		"rebuild":    func(r *CacheReceipt) { r.Reconstructable = false },
	} {
		t.Run(name, func(t *testing.T) {
			mutated := receipt
			mutate(&mutated)
			if name == "content" || name == "size" {
				if err := mutated.ValidateForData([]byte("projection")); !errors.Is(err, ErrInvalidReceipt) {
					t.Fatalf("mutated data receipt = %v, want ErrInvalidReceipt", err)
				}
				return
			}
			if err := mutated.ValidateForKey(key); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("mutated identity receipt = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}

func TestCacheReceiptC1C4OptionsDigestFailsClosed(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	receipt := cacheReceiptFixture(key, "options")
	missing := receipt
	missing.OptionsDigest = ""
	if err := missing.ValidateForKey(key); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("missing options digest = %v, want ErrInvalidReceipt", err)
	}
	variant := projectionSpec()
	variant.OptionsDigest = mustOptionsDigest(map[string]any{"mode": "other"})
	variantKey, err := NewProjectionKey(variant)
	if err != nil {
		t.Fatal(err)
	}
	if variantKey == key {
		t.Fatal("options mutation retained projection identity")
	}
	receipt.OptionsDigest = variantKey.OptionsDigest
	if err := receipt.ValidateForKey(key); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("mismatched options digest = %v, want ErrInvalidReceipt", err)
	}
}

func TestCacheReceiptC3C5RequiresImmutableEvidenceBundle(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	receipt := cacheReceiptFixture(key, "run-1")
	sealed, err := receipt.Seal()
	if err != nil {
		t.Fatal(err)
	}
	cache, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealed); !errors.Is(err, ErrReceiptReplay) {
		t.Fatalf("replayed receipt = %v, want ErrReceiptReplay", err)
	}
	eventReplay := cacheReceiptFixture(key, "run-2")
	eventReplay.Evidence.EventID = sealed.Evidence.EventID
	sealedEventReplay, err := eventReplay.Seal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(sealedEventReplay); !errors.Is(err, ErrReceiptReplay) {
		t.Fatalf("event/attempt replay = %v, want ErrReceiptReplay", err)
	}
	receipts, err := cache.Receipts()
	if err != nil || len(receipts) != 1 {
		t.Fatalf("receipt log = %d, %v", len(receipts), err)
	}
	missing := receipt
	missing.Evidence.BundleDigest = ""
	assertInvalidSeal(t, "missing bundle", missing)
	missingEvent := receipt
	missingEvent.Evidence.EventID = ""
	assertInvalidSeal(t, "missing event", missingEvent)
	missingAttempt := receipt
	missingAttempt.Evidence.Attempt = 0
	assertInvalidSeal(t, "missing attempt", missingAttempt)
	missingSHA := receipt
	missingSHA.Evidence.HeadSHA = ""
	assertInvalidSeal(t, "missing head SHA", missingSHA)
	missingJobs := receipt
	missingJobs.Evidence.Jobs = nil
	assertInvalidSeal(t, "missing canonical jobs", missingJobs)
	zero := receipt
	zero.Evidence.BundleDigest = Digest("0000000000000000000000000000000000000000000000000000000000000000")
	assertInvalidSeal(t, "zero bundle", zero)
	unknownRef := receipt
	unknownRef.EvidenceRefs = nil
	unknownRef.Evidence.EvidenceRefs = nil
	assertInvalidSeal(t, "missing evidence ref", unknownRef)
	unboundRef := receipt
	unboundRef.EvidenceRefs = append([]EvidenceRef(nil), receipt.EvidenceRefs...)
	unboundRef.EvidenceRefs = append(unboundRef.EvidenceRefs,
		EvidenceRef{Name: "unbound", Digest: HashBytes([]byte("unbound"))})
	unboundRef.Evidence.EvidenceRefs = append([]EvidenceRef(nil), unboundRef.EvidenceRefs...)
	assertInvalidSeal(t, "unbound evidence ref", unboundRef)
	zeroRef := receipt
	zeroRef.EvidenceRefs = []EvidenceRef{{Name: "source", Digest: Digest("0000000000000000000000000000000000000000000000000000000000000000")}}
	zeroRef.Evidence.EvidenceRefs = zeroRef.EvidenceRefs
	assertInvalidSeal(t, "zero evidence ref", zeroRef)
}

func assertInvalidSeal(t *testing.T, name string, receipt CacheReceipt) {
	t.Helper()
	if _, err := receipt.Seal(); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("%s = %v, want ErrInvalidReceipt", name, err)
	}
}

func TestBenchmarkReceiptBindsCanonicalJobs(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	benchmark := BenchmarkReceipt{SchemaVersion: benchmarkReceiptSchemaVersion, Fixture: "partial",
		BaseDigest: key.Digest, HeadDigest: key.Digest, BaseSHA: commitFixtureSHA("bench-base"), HeadSHA: commitFixtureSHA("bench-head"),
		Event: "pull_request", Workflow: "CI [PR authoritative]", WorkflowRunID: "31560000000",
		RunID: "bench-1", EventID: "event-bench-1", Attempt: 1,
		Filesystem: "local", ToolchainDigest: HashBytes([]byte("go1.26.5")), PolicyDigest: HashBytes([]byte("policy")),
		EvidenceRefs: benchmarkEvidenceRefs(), Jobs: benchmarkJobs(key.Digest, commitFixtureSHA("bench-head")), P50Nanoseconds: 10, P95Nanoseconds: 20}
	delete(benchmark.Jobs, canonicalRaceJob)
	if err := benchmark.Validate(); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("incomplete benchmark = %v, want ErrInvalidReceipt", err)
	}
	benchmark.Jobs[canonicalRaceJob] = BenchmarkJob{ID: "6", Status: "completed", Conclusion: "success", HeadSHA: key.Digest, HeadCommitSHA: commitFixtureSHA("bench-head")}
	if err := benchmark.Validate(); err != nil {
		t.Fatalf("complete benchmark = %v", err)
	}
	for name, mutate := range map[string]func(*BenchmarkReceipt){
		"base sha":     func(r *BenchmarkReceipt) { r.BaseSHA = commitFixtureSHA("other-base") },
		"head sha":     func(r *BenchmarkReceipt) { r.HeadSHA = commitFixtureSHA("other-head") },
		"event":        func(r *BenchmarkReceipt) { r.Event = "push" },
		"workflow":     func(r *BenchmarkReceipt) { r.Workflow = "" },
		"workflow run": func(r *BenchmarkReceipt) { r.WorkflowRunID = "" },
		"job status": func(r *BenchmarkReceipt) {
			job := r.Jobs[canonicalTestJob]
			job.Status = ""
			r.Jobs[canonicalTestJob] = job
		},
		"job conclusion": func(r *BenchmarkReceipt) {
			job := r.Jobs[canonicalTestJob]
			job.Conclusion = ""
			r.Jobs[canonicalTestJob] = job
		},
		"job head": func(r *BenchmarkReceipt) {
			job := r.Jobs[canonicalTestJob]
			job.HeadSHA = HashBytes([]byte("other"))
			r.Jobs[canonicalTestJob] = job
		},
		"job commit head": func(r *BenchmarkReceipt) {
			job := r.Jobs[canonicalTestJob]
			job.HeadCommitSHA = commitFixtureSHA("other-head")
			r.Jobs[canonicalTestJob] = job
		},
		"policy ref": func(r *BenchmarkReceipt) { r.EvidenceRefs[0].Digest = HashBytes([]byte("other")) },
	} {
		t.Run(name, func(t *testing.T) {
			mutated := benchmark
			mutate(&mutated)
			if err := mutated.Validate(); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("mutated benchmark = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}

func benchmarkEvidenceRefs() []EvidenceRef {
	return []EvidenceRef{
		{Name: "policy", Digest: HashBytes([]byte("policy"))},
		{Name: "toolchain", Digest: HashBytes([]byte("go1.26.5"))},
	}
}

func benchmarkJobs(head Digest, commitHead string) map[string]BenchmarkJob {
	jobs := make(map[string]BenchmarkJob, len(canonicalBenchmarkJobs))
	for index, name := range canonicalBenchmarkJobs {
		jobs[name] = BenchmarkJob{ID: string(rune('1' + index)), Status: "completed", Conclusion: "success", HeadSHA: head, HeadCommitSHA: commitHead}
	}
	return jobs
}
