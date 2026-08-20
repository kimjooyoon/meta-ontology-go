package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvidenceRejectsMissingCanonicalJob(t *testing.T) {
	jobs := validJobs()
	jobs = jobs[:len(jobs)-1]
	if _, err := normalizeJobs(jobs, validScheduler(), strings.Repeat("a", 40), 1, 1); err == nil {
		t.Fatal("missing canonical job was accepted")
	}
}

func TestEvidenceRejectsMismatchedJobHead(t *testing.T) {
	jobs := validJobs()
	jobs[0].HeadSHA = strings.Repeat("b", 40)
	if _, err := normalizeJobs(jobs, validScheduler(), strings.Repeat("a", 40), 1, 1); err == nil {
		t.Fatal("mismatched canonical job head was accepted")
	}
}

func TestEvidenceRejectsMismatchedJobRun(t *testing.T) {
	jobs := validJobs()
	jobs[0].RunID = 2
	if _, err := normalizeJobs(jobs, validScheduler(), strings.Repeat("a", 40), 1, 1); err == nil {
		t.Fatal("mismatched canonical job run was accepted")
	}
}

func TestEvidenceRejectsMismatchedJobAttempt(t *testing.T) {
	jobs := validJobs()
	jobs[0].RunAttempt = 2
	if _, err := normalizeJobs(jobs, validScheduler(), strings.Repeat("a", 40), 1, 1); err == nil {
		t.Fatal("mismatched canonical job attempt was accepted")
	}
}

func TestEvidenceRejectsDuplicateCanonicalJob(t *testing.T) {
	jobs := validJobs()
	jobs = append(jobs, jobs[0])
	if _, err := normalizeJobs(jobs, validScheduler(), strings.Repeat("a", 40), 1, 1); err == nil {
		t.Fatal("duplicate canonical job was accepted")
	}
}

func TestEvidenceRawLagNeedsSeparateSchedulerBinding(t *testing.T) {
	jobs := validJobs()
	jobs[0].Status = stringPointer("in_progress")
	jobs[0].Conclusion = nil
	if _, err := normalizeJobs(jobs, nil, strings.Repeat("a", 40), 1, 1); err == nil {
		t.Fatal("raw in-progress/null API observation was accepted without scheduler evidence")
	}
	normalized, err := normalizeJobs(jobs, validScheduler(), strings.Repeat("a", 40), 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if normalized[0].ObservationState != observerLag || normalized[0].Status == nil || *normalized[0].Status != "in_progress" || normalized[0].Conclusion != nil {
		t.Fatalf("raw API lag was not retained distinctly: %+v", normalized[0])
	}
}

func TestEvidenceRejectsEmptyPolicyConclusion(t *testing.T) {
	jobs := validJobs()
	jobs[len(jobs)-1].Conclusion = stringPointer("")
	if _, err := normalizeJobs(jobs, validScheduler(), strings.Repeat("a", 40), 1, 1); err == nil {
		t.Fatal("empty policy conclusion was accepted")
	}
}

func TestEvidenceRejectsMalformedRawStatus(t *testing.T) {
	jobs := validJobs()
	jobs[0].Status = stringPointer("finished")
	if _, err := normalizeJobs(jobs, validScheduler(), strings.Repeat("a", 40), 1, 1); err == nil {
		t.Fatal("arbitrary raw API status was accepted")
	}
}

func TestEvidencePR248HistoricalObserverLagFixture(t *testing.T) {
	apiData, err := os.ReadFile(filepath.Join("testdata", "pr248-jobs.json"))
	if err != nil {
		t.Fatal(err)
	}
	schedulerData, err := os.ReadFile(filepath.Join("testdata", "pr248-scheduler.json"))
	if err != nil {
		t.Fatal(err)
	}
	var apiJobs []apiJob
	if err := json.Unmarshal(apiData, &apiJobs); err != nil {
		t.Fatal(err)
	}
	var scheduler []schedulerEvidence
	if err := json.Unmarshal(schedulerData, &scheduler); err != nil {
		t.Fatal(err)
	}
	const head = "da5125fd247fe65fd7335dacb1c77529e989e6ba"
	jobs, err := normalizeJobs(apiJobs, scheduler, head, 31937165616, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != len(canonicalJobs) {
		t.Fatalf("fixture normalized %d jobs, want %d", len(jobs), len(canonicalJobs))
	}
	var gofmt jobEvidence
	for _, job := range jobs {
		if job.Name == "gofmt" {
			gofmt = job
		}
	}
	if gofmt.ID != 95140792870 || gofmt.ObservationState != observerLag || gofmt.Status == nil || *gofmt.Status != "in_progress" || gofmt.Conclusion != nil || gofmt.CompletedAt != nil {
		t.Fatalf("historical #248 raw lag was not retained: %+v", gofmt)
	}
	if apiJobs[0].RunAttempt != 1 {
		t.Fatalf("historical #248 fixture did not bind run_attempt from the REST job object: %+v", apiJobs[0])
	}
}

func TestEvidenceAcceptsLagOnlyWithSchedulerBinding(t *testing.T) {
	jobs := validJobs()
	jobs[len(jobs)-1].Status = stringPointer("in_progress")
	jobs[len(jobs)-1].Conclusion = nil
	if _, err := normalizeJobs(jobs, validScheduler(), strings.Repeat("a", 40), 1, 1); err != nil {
		t.Fatal(err)
	}
}

func TestEvidenceRejectsBundleDigestMismatch(t *testing.T) {
	bundle := validEvidence()
	bundle.Digests.BundleSHA256 = strings.Repeat("b", 64)
	if err := validateEvidence(bundle); err == nil {
		t.Fatal("mismatched bundle digest was accepted")
	}
}

func validJobs() []apiJob {
	head := strings.Repeat("a", 40)
	jobs := make([]apiJob, len(canonicalJobs))
	for index, name := range canonicalJobs {
		jobs[index] = apiJob{ID: int64(index + 1), Name: name, Status: stringPointer("completed"), Conclusion: stringPointer("success"), HeadSHA: head, RunID: 1, RunAttempt: 1, CompletedAt: stringPointer("2026-08-16T08:44:18Z")}
	}
	return jobs
}

func validScheduler() []schedulerEvidence {
	sources := []string{"needs.format.result", "needs.vet.result", "needs.test.result", "needs.race.result", "needs.semantic.result", "needs.policy.result"}
	head := strings.Repeat("a", 40)
	results := make([]schedulerEvidence, len(canonicalJobs))
	for index, name := range canonicalJobs {
		results[index] = schedulerEvidence{Name: name, Source: sources[index], Result: "success", HeadSHA: head, RunID: 1, RunAttempt: 1}
	}
	return results
}

func stringPointer(value string) *string {
	return &value
}

func validEvidence() evidence {
	head := strings.Repeat("a", 40)
	jobs := make([]jobEvidence, len(canonicalJobs))
	for index, name := range canonicalJobs {
		jobs[index] = jobEvidence{ID: int64(index + 1), Name: name, Status: stringPointer("completed"), Conclusion: stringPointer("success"), HeadSHA: head, RunID: 1, RunAttempt: 1, CompletedAt: stringPointer("2026-08-16T08:44:18Z"), ObservationState: apiTerminalSuccess}
	}
	bundle := evidence{Schema: evidenceSchema, Repository: "owner/repo", Event: "pull_request", EventRef: "refs/pull/1/merge", CheckoutRef: head, BaseRef: "dev", BaseSHA: strings.Repeat("b", 40), HeadSHA: head, RunID: 1, RunAttempt: 1, WorkflowSHA: strings.Repeat("c", 40), Toolchain: "go1.26.5", SlotPreservation: true, NoWriteOutsideGenerated: true, Scheduler: validScheduler(), Jobs: jobs, Digests: digests{SourceSHA256: strings.Repeat("1", 64), IRSHA256: strings.Repeat("2", 64), GeneratorFixtureSHA256: strings.Repeat("3", 64), GeneratedOutputSHA256: strings.Repeat("4", 64), SourceMapSHA256: strings.Repeat("5", 64), PolicySHA256: strings.Repeat("6", 64), ToolchainSHA256: strings.Repeat("7", 64)}}
	payload, _ := marshalWithoutBundleDigest(bundle)
	bundle.Digests.BundleSHA256 = digestBytes(payload)
	return bundle
}

func marshalWithoutBundleDigest(bundle evidence) ([]byte, error) {
	bundle.Digests.BundleSHA256 = ""
	return json.Marshal(bundle)
}
