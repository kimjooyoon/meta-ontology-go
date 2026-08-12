package cache

import (
	"errors"
	"testing"
)

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
	benchmark.Jobs[canonicalRaceJob] = BenchmarkJob{ID: "6", RunID: benchmark.WorkflowRunID, Attempt: benchmark.Attempt,
		Status: "completed", Conclusion: "success", HeadSHA: key.Digest, HeadCommitSHA: commitFixtureSHA("bench-head")}
	if err := benchmark.Validate(); err != nil {
		t.Fatalf("complete benchmark = %v", err)
	}
	for name, mutate := range map[string]func(*BenchmarkReceipt){
		"base sha":     func(r *BenchmarkReceipt) { r.BaseSHA = "not-a-commit-sha" },
		"head sha":     func(r *BenchmarkReceipt) { r.HeadSHA = commitFixtureSHA("other-head") },
		"event":        func(r *BenchmarkReceipt) { r.Event = "push" },
		"workflow":     func(r *BenchmarkReceipt) { r.Workflow = "" },
		"workflow run": func(r *BenchmarkReceipt) { r.WorkflowRunID = "" },
		"job run": func(r *BenchmarkReceipt) {
			job := r.Jobs[canonicalTestJob]
			job.RunID = "other-run"
			r.Jobs[canonicalTestJob] = job
		},
		"job attempt": func(r *BenchmarkReceipt) {
			job := r.Jobs[canonicalTestJob]
			job.Attempt++
			r.Jobs[canonicalTestJob] = job
		},
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
		jobs[name] = BenchmarkJob{ID: string(rune('1' + index)), RunID: "31560000000", Attempt: 1,
			Status: "completed", Conclusion: "success", HeadSHA: head, HeadCommitSHA: commitHead}
	}
	return jobs
}
