package cache

import (
	"strconv"
)

func evidenceFixture(run string) EvidenceFreshness {
	headSHA := commitFixtureSHA("head")
	return EvidenceFreshness{
		BaseDigest: HashBytes([]byte("base")), HeadDigest: HashBytes([]byte("head")), RunID: run,
		BaseSHA: commitFixtureSHA("base"), HeadSHA: headSHA, CheckoutRef: headSHA,
		Event: "pull_request", EventRef: "refs/pull/8/merge", EventID: "event-" + run, Attempt: 1,
		Jobs:               freshnessJobsFixture(headSHA, run, 1),
		PredecessorDigests: []Digest{HashBytes([]byte("generator")), HashBytes([]byte("semantic"))},
		SourceDigest:       HashBytes([]byte("source")), IRDigest: HashBytes([]byte("ir")),
		PolicyDigest: HashBytes([]byte("policy")), ToolchainDigest: HashBytes([]byte("go1.26.5")),
		TargetDigest: HashBytes([]byte("darwin/arm64")), BundleDigest: HashBytes([]byte("bundle-" + run)),
		EvidenceRefs: []EvidenceRef{
			{Name: "source", Digest: HashBytes([]byte("source-ref"))},
			{Name: "bundle", Digest: HashBytes([]byte("bundle-ref"))},
			{Name: "policy", Digest: HashBytes([]byte("policy"))},
			{Name: "toolchain", Digest: HashBytes([]byte("go1.26.5"))},
		},
	}
}
func commitFixtureSHA(value string) string {
	return string(HashBytes([]byte("commit:" + value)))
}
func freshnessJobsFixture(headSHA, runID string, attempt uint64) map[string]FreshnessJob {
	jobs := make(map[string]FreshnessJob, len(canonicalBenchmarkJobs))
	for index, name := range canonicalBenchmarkJobs {
		jobs[name] = FreshnessJob{
			ID: strconv.Itoa(index + 1), RunID: runID, Attempt: attempt,
			Status: "completed", Conclusion: "success", HeadSHA: headSHA,
		}
	}
	return jobs
}
