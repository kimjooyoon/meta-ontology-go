package cache

import (
	"errors"
	"testing"
)

func TestBenchmarkReceiptBindsCanonicalJobs(t *testing.T) {
	benchmark := validBenchmarkReceipt(t)
	assertBenchmarkReceiptIdentity(t, benchmark)
	for name, mutate := range benchmarkReceiptMutations() {
		t.Run(name, func(t *testing.T) {
			mutated := cloneBenchmarkReceipt(benchmark)
			mutate(&mutated)
			if err := mutated.Validate(); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("mutated benchmark = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}
func validBenchmarkReceipt(t *testing.T) BenchmarkReceipt {
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
	benchmark.Jobs[canonicalRaceJob] = BenchmarkJob{ID: "6", RunID: benchmark.WorkflowRunID, Attempt: benchmark.Attempt,
		Status: "completed", Conclusion: "success", HeadSHA: key.Digest, HeadCommitSHA: commitFixtureSHA("bench-head")}
	if err := benchmark.Validate(); err != nil {
		t.Fatalf("complete benchmark = %v", err)
	}
	return benchmark
}
func assertBenchmarkReceiptIdentity(t *testing.T, benchmark BenchmarkReceipt) {
	firstDigest, err := DigestOf(benchmark)
	if err != nil {
		t.Fatal(err)
	}
	reordered := cloneBenchmarkReceipt(benchmark)
	reordered.Jobs = make(map[string]BenchmarkJob, len(benchmark.Jobs))
	for index := len(canonicalBenchmarkJobs) - 1; index >= 0; index-- {
		name := canonicalBenchmarkJobs[index]
		reordered.Jobs[name] = benchmark.Jobs[name]
	}
	reorderedDigest, err := DigestOf(reordered)
	if err != nil {
		t.Fatal(err)
	}
	if reorderedDigest != firstDigest {
		t.Fatalf("job map insertion order changed receipt digest: %s != %s", reorderedDigest, firstDigest)
	}
	mutatedIdentity := cloneBenchmarkReceipt(benchmark)
	mutatedIdentity.Attempt++
	mutatedDigest, err := DigestOf(mutatedIdentity)
	if err != nil {
		t.Fatal(err)
	}
	if mutatedDigest == firstDigest {
		t.Fatal("receipt identity ignored attempt mutation")
	}
}
