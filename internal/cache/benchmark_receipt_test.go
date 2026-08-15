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

func benchmarkReceiptMutations() map[string]func(*BenchmarkReceipt) {
	return map[string]func(*BenchmarkReceipt){
		"base sha":       func(r *BenchmarkReceipt) { r.BaseSHA = "not-a-commit-sha" },
		"head sha":       func(r *BenchmarkReceipt) { r.HeadSHA = commitFixtureSHA("other-head") },
		"event":          func(r *BenchmarkReceipt) { r.Event = "push" },
		"workflow":       func(r *BenchmarkReceipt) { r.Workflow = "" },
		"workflow run":   func(r *BenchmarkReceipt) { r.WorkflowRunID = "" },
		"job run":        mutateBenchmarkJob(func(job *BenchmarkJob) { job.RunID = "other-run" }),
		"job attempt":    mutateBenchmarkJob(func(job *BenchmarkJob) { job.Attempt++ }),
		"job status":     mutateBenchmarkJob(func(job *BenchmarkJob) { job.Status = "" }),
		"queued job":     mutateBenchmarkJob(func(job *BenchmarkJob) { job.Status = "queued" }),
		"job conclusion": mutateBenchmarkJob(func(job *BenchmarkJob) { job.Conclusion = "" }),
		"failed job":     mutateBenchmarkJob(func(job *BenchmarkJob) { job.Conclusion = "failure" }),
		"job head":       mutateBenchmarkJob(func(job *BenchmarkJob) { job.HeadSHA = HashBytes([]byte("other")) }),
		"job commit head": mutateBenchmarkJob(func(job *BenchmarkJob) {
			job.HeadCommitSHA = commitFixtureSHA("other-head")
		}),
		"policy ref": func(r *BenchmarkReceipt) { r.EvidenceRefs[0].Digest = HashBytes([]byte("other")) },
	}
}

func mutateBenchmarkJob(mutate func(*BenchmarkJob)) func(*BenchmarkReceipt) {
	return func(receipt *BenchmarkReceipt) {
		job := receipt.Jobs[canonicalTestJob]
		mutate(&job)
		receipt.Jobs[canonicalTestJob] = job
	}
}

func cloneBenchmarkReceipt(receipt BenchmarkReceipt) BenchmarkReceipt {
	receipt.Jobs = make(map[string]BenchmarkJob, len(receipt.Jobs))
	for name, job := range receipt.Jobs {
		receipt.Jobs[name] = job
	}
	receipt.EvidenceRefs = append([]EvidenceRef(nil), receipt.EvidenceRefs...)
	return receipt
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
		jobs[name] = BenchmarkJob{ID: string(rune('1' + index)), RunID: "31560000000", Attempt: 1,
			Status: "completed", Conclusion: "success", HeadSHA: head, HeadCommitSHA: commitHead}
	}
	return jobs
}
