package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEvidenceRejectsMissingCanonicalJob(t *testing.T) {
	jobs := validJobs()
	jobs = jobs[:len(jobs)-1]
	if _, err := normalizeJobs(jobs, strings.Repeat("a", 40), 1, 1, true); err == nil {
		t.Fatal("missing canonical job was accepted")
	}
}

func TestEvidenceRejectsMismatchedJobHead(t *testing.T) {
	jobs := validJobs()
	jobs[0].HeadSHA = strings.Repeat("b", 40)
	if _, err := normalizeJobs(jobs, strings.Repeat("a", 40), 1, 1, true); err == nil {
		t.Fatal("mismatched canonical job head was accepted")
	}
}

func TestEvidenceRejectsEmptyPolicyConclusion(t *testing.T) {
	jobs := validJobs()
	jobs[len(jobs)-1].Conclusion = ""
	if _, err := normalizeJobs(jobs, strings.Repeat("a", 40), 1, 1, false); err == nil {
		t.Fatal("empty policy conclusion was accepted")
	}
}

func TestEvidenceBindsSelfPolicySuccessAsCompleted(t *testing.T) {
	jobs := validJobs()
	policy := len(jobs) - 1
	jobs[policy].Status = "in_progress"
	jobs[policy].Conclusion = ""
	result, err := normalizeJobs(jobs, strings.Repeat("a", 40), 1, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	if result[policy].Status != "completed" || result[policy].Conclusion != "success" {
		t.Fatalf("self policy was not normalized to a terminal success: %+v", result[policy])
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
		jobs[index] = apiJob{ID: int64(index + 1), Name: name, Status: "completed", Conclusion: "success", HeadSHA: head, RunID: 1}
	}
	return jobs
}

func validEvidence() evidence {
	head := strings.Repeat("a", 40)
	jobs := make([]jobEvidence, len(canonicalJobs))
	for index, name := range canonicalJobs {
		jobs[index] = jobEvidence{ID: int64(index + 1), Name: name, Status: "completed", Conclusion: "success", HeadSHA: head, RunID: 1, RunAttempt: 1}
	}
	bundle := evidence{Schema: evidenceSchema, Repository: "owner/repo", Event: "pull_request", EventRef: "refs/pull/1/merge", CheckoutRef: head, BaseRef: "integration", BaseSHA: strings.Repeat("b", 40), HeadSHA: head, RunID: 1, RunAttempt: 1, WorkflowSHA: strings.Repeat("c", 40), Toolchain: "go1.26.5", SlotPreservation: true, NoWriteOutsideGenerated: true, Jobs: jobs, Digests: digests{SourceSHA256: strings.Repeat("1", 64), IRSHA256: strings.Repeat("2", 64), GeneratorFixtureSHA256: strings.Repeat("3", 64), GeneratedOutputSHA256: strings.Repeat("4", 64), SourceMapSHA256: strings.Repeat("5", 64), PolicySHA256: strings.Repeat("6", 64), ToolchainSHA256: strings.Repeat("7", 64)}}
	payload, _ := marshalWithoutBundleDigest(bundle)
	bundle.Digests.BundleSHA256 = digestBytes(payload)
	return bundle
}

func marshalWithoutBundleDigest(bundle evidence) ([]byte, error) {
	bundle.Digests.BundleSHA256 = ""
	return json.Marshal(bundle)
}
