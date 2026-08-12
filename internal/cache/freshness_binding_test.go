package cache

import (
	"strconv"
	"testing"
)

func evidenceFixture(run string) EvidenceFreshness {
	headSHA := commitFixtureSHA("head")
	return EvidenceFreshness{
		BaseDigest: HashBytes([]byte("base")), HeadDigest: HashBytes([]byte("head")), RunID: run,
		BaseSHA: commitFixtureSHA("base"), HeadSHA: headSHA,
		Event: "pull_request", EventID: "event-" + run, Attempt: 1,
		Jobs:               freshnessJobsFixture(headSHA),
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

func freshnessJobsFixture(headSHA string) map[string]FreshnessJob {
	jobs := make(map[string]FreshnessJob, len(canonicalBenchmarkJobs))
	for index, name := range canonicalBenchmarkJobs {
		jobs[name] = FreshnessJob{
			ID: strconv.Itoa(index + 1), Status: "completed", Conclusion: "success", HeadSHA: headSHA,
		}
	}
	return jobs
}

func TestEvidenceFreshnessC4RejectsStaleAndReplayTuples(t *testing.T) {
	current := evidenceFixture("run-current")
	for name, mutate := range map[string]func(*EvidenceFreshness){
		"base":       func(e *EvidenceFreshness) { e.BaseDigest = HashBytes([]byte("new-base")) },
		"head":       func(e *EvidenceFreshness) { e.HeadDigest = HashBytes([]byte("new-head")) },
		"base sha":   func(e *EvidenceFreshness) { e.BaseSHA = commitFixtureSHA("new-base") },
		"head sha":   func(e *EvidenceFreshness) { e.HeadSHA = commitFixtureSHA("new-head") },
		"run":        func(e *EvidenceFreshness) { e.RunID = "run-other" },
		"event kind": func(e *EvidenceFreshness) { e.Event = "push" },
		"event":      func(e *EvidenceFreshness) { e.EventID = "event-other" },
		"attempt":    func(e *EvidenceFreshness) { e.Attempt++ },
		"job head": func(e *EvidenceFreshness) {
			jobs := copyFreshnessJobs(e.Jobs)
			job := jobs[canonicalTestJob]
			job.HeadSHA = commitFixtureSHA("other-head")
			jobs[canonicalTestJob] = job
			e.Jobs = jobs
		},
		"job status": func(e *EvidenceFreshness) {
			jobs := copyFreshnessJobs(e.Jobs)
			job := jobs[canonicalTestJob]
			job.Status = "in_progress"
			jobs[canonicalTestJob] = job
			e.Jobs = jobs
		},
		"job conclusion": func(e *EvidenceFreshness) {
			jobs := copyFreshnessJobs(e.Jobs)
			job := jobs[canonicalTestJob]
			job.Conclusion = "failure"
			jobs[canonicalTestJob] = job
			e.Jobs = jobs
		},
		"duplicate job id": func(e *EvidenceFreshness) {
			jobs := copyFreshnessJobs(e.Jobs)
			job := jobs[canonicalTestJob]
			job.ID = jobs[canonicalCIPolicyJob].ID
			jobs[canonicalTestJob] = job
			e.Jobs = jobs
		},
		"prior": func(e *EvidenceFreshness) { e.PredecessorDigests[0] = HashBytes([]byte("other")) },
	} {
		t.Run(name, func(t *testing.T) {
			stale := canonicalEvidence(current)
			mutate(&stale)
			if stale.Matches(current) {
				t.Fatal("stale evidence matched current tuple")
			}
		})
	}
}

func TestProjectionIdentityRequiresTypedFreshness(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	evidence := evidenceFixture("projection-hit")
	identity := ProjectionIdentity{
		SemanticClosureDigest: key.SemanticClosureDigest, SourceDigest: evidence.SourceDigest,
		IRDigest: evidence.IRDigest, OptionsDigest: key.OptionsDigest,
		Toolchain: key.Toolchain, ToolchainDigest: evidence.ToolchainDigest,
	}
	if err := identity.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*ProjectionIdentity){
		"source":    func(i *ProjectionIdentity) { i.SourceDigest = HashBytes([]byte("other-source")) },
		"IR":        func(i *ProjectionIdentity) { i.IRDigest = HashBytes([]byte("other-ir")) },
		"options":   func(i *ProjectionIdentity) { i.OptionsDigest = HashBytes([]byte("other-options")) },
		"toolchain": func(i *ProjectionIdentity) { i.Toolchain = "go1.26.6" },
	} {
		t.Run(name, func(t *testing.T) {
			mutated := identity
			mutate(&mutated)
			if mutated.matchesKey(key) && mutated.matchesEvidence(evidence) {
				t.Fatal("identity mutation retained both key and evidence identity")
			}
		})
	}
}
