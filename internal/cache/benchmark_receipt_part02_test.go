package cache

import "maps"

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
	maps.Copy(receipt.Jobs, receipt.Jobs)
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
