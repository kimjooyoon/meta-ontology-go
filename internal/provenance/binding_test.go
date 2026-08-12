package provenance

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBindingCanonicalHashIsPermutationInvariant(t *testing.T) {
	sourceHash := strings.Repeat("a", 64)
	leftBinding := validBinding(sourceHash)
	reverseJobs(leftBinding.Jobs)
	leftBinding.Predecessors = []string{"predecessor/b", "predecessor/a"}
	leftBinding.EvidenceRefs = []string{"evidence/b", "evidence/a"}
	rightBinding := validBinding(sourceHash)
	rightBinding.Predecessors = []string{"predecessor/a", "predecessor/b"}
	rightBinding.EvidenceRefs = []string{"evidence/a", "evidence/b"}
	recordLeft := boundEvidence("evidence/authoritative", leftBinding)
	recordRight := boundEvidence("evidence/authoritative", rightBinding)
	left := New(filepath.Join(t.TempDir(), "left", "evidence.jsonl"))
	right := New(filepath.Join(t.TempDir(), "right", "evidence.jsonl"))
	if err := left.Append(recordLeft); err != nil {
		t.Fatal(err)
	}
	if err := right.Append(recordRight); err != nil {
		t.Fatal(err)
	}
	leftBytes, err := os.ReadFile(left.Path())
	if err != nil {
		t.Fatal(err)
	}
	rightBytes, err := os.ReadFile(right.Path())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(leftBytes, rightBytes) {
		t.Fatalf("binding permutation changed canonical JSONL:\nleft=%s\nright=%s", leftBytes, rightBytes)
	}
	snapshot, err := left.Read(ReadOptions{ExpectedSourceHash: sourceHash, ExpectedBinding: rightBinding})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Records) != 1 || snapshot.Records[0].Binding == nil || !strings.Contains(string(leftBytes), `"binding"`) {
		t.Fatalf("binding was not retained in canonical record: %#v", snapshot)
	}
}

func TestBindingTupleMismatchFailsClosed(t *testing.T) {
	sourceHash := strings.Repeat("b", 64)
	binding := validBinding(sourceHash)
	store := New(filepath.Join(t.TempDir(), "evidence.jsonl"))
	if err := store.Append(boundEvidence("evidence/verification", binding)); err != nil {
		t.Fatal(err)
	}
	expected := validBinding(sourceHash)
	expected.PolicyDigest = strings.Repeat("c", 64)
	_, err := store.Read(ReadOptions{ExpectedBinding: expected})
	var mismatch *BindingMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("tuple mismatch was accepted: %v", err)
	}
	before, err := os.ReadFile(store.Path())
	if err != nil {
		t.Fatal(err)
	}
	internalMismatch := validBinding(sourceHash)
	internalMismatch.BundleDigest = strings.Repeat("e", 64)
	record := boundEvidence("evidence/internal-mismatch", internalMismatch)
	record.Freshness.SourceHash = sourceHash
	var bindingErr *BindingError
	if err := store.Append(record); !errors.As(err, &bindingErr) || bindingErr.Kind != "source" {
		t.Fatalf("internal tuple mismatch was accepted: %v", err)
	}
	assertBytesUnchanged(t, store.Path(), before)
	_, err = store.Read(ReadOptions{ExpectedSourceHash: strings.Repeat("d", 64)})
	var freshness *FreshnessError
	if !errors.As(err, &freshness) || freshness.Kind != "source-mismatch" {
		t.Fatalf("stale source was accepted: %v", err)
	}
}

func TestTamperedJobHeadAndConclusionFailClosed(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*RunBinding)
		kind   string
	}{
		{name: "head", mutate: func(binding *RunBinding) { binding.Jobs[0].HeadSHA = strings.Repeat("e", 40) }, kind: "jobs"},
		{name: "conclusion", mutate: func(binding *RunBinding) { binding.Jobs[0].Conclusion = "failure" }, kind: "jobs"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			sourceHash := strings.Repeat("e", 64)
			store := New(filepath.Join(t.TempDir(), "evidence.jsonl"))
			baseline := testEvidence("evidence/baseline", sourceHash, time.Date(2026, 8, 12, 5, 0, 0, 0, time.UTC), time.Time{})
			if err := store.Append(baseline); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(store.Path())
			if err != nil {
				t.Fatal(err)
			}
			binding := validBinding(sourceHash)
			testCase.mutate(binding)
			err = store.Append(boundEvidence("evidence/tampered-"+testCase.name, binding))
			var bindingErr *BindingError
			if !errors.As(err, &bindingErr) || bindingErr.Kind != testCase.kind {
				t.Fatalf("tampered binding was accepted or misclassified: %v", err)
			}
			after, err := os.ReadFile(store.Path())
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatal("tampered append changed existing bytes")
			}
		})
	}
}

func TestStoredJobTamperIsCorruption(t *testing.T) {
	for _, testCase := range []struct {
		name string
		old  string
		new  string
		kind string
	}{
		{name: "head", old: `"head_sha":"` + strings.Repeat("1", 40) + `"`, new: `"head_sha":"` + strings.Repeat("9", 40) + `"`, kind: "binding-invalid"},
		{name: "conclusion", old: `"conclusion":"success"`, new: `"conclusion":"failure"`, kind: "binding-invalid"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			sourceHash := strings.Repeat("9", 64)
			path := filepath.Join(t.TempDir(), "evidence.jsonl")
			store := New(path)
			if err := store.Append(boundEvidence("evidence/tamper", validBinding(sourceHash))); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			mutated := bytes.Replace(data, []byte(testCase.old), []byte(testCase.new), 1)
			if bytes.Equal(data, mutated) {
				t.Fatal("test mutation did not change canonical bytes")
			}
			if err := os.WriteFile(path, mutated, 0o644); err != nil {
				t.Fatal(err)
			}
			_, err = store.Read(ReadOptions{})
			var diagnostic *CorruptionError
			if !errors.As(err, &diagnostic) || diagnostic.Kind != testCase.kind {
				t.Fatalf("stored tamper was accepted or misclassified: %v", err)
			}
		})
	}
}

func TestReplayMissingBindingAndWriteEffectPreserveBytes(t *testing.T) {
	sourceHash := strings.Repeat("f", 64)
	store := New(filepath.Join(t.TempDir(), "evidence.jsonl"))
	firstBinding := validBinding(sourceHash)
	firstBinding.Predecessors = []string{"predecessor/one"}
	if err := store.Append(boundEvidence("evidence/first", firstBinding)); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(store.Path())
	if err != nil {
		t.Fatal(err)
	}
	replayBinding := validBinding(sourceHash)
	replayBinding.Predecessors = []string{"predecessor/one"}
	err = store.Append(boundEvidence("evidence/replay", replayBinding))
	var replay *ReplayError
	if !errors.As(err, &replay) {
		t.Fatalf("predecessor replay was accepted: %v", err)
	}
	assertBytesUnchanged(t, store.Path(), before)

	missing := boundEvidence("evidence/missing", nil)
	err = store.Append(missing)
	var bindingErr *BindingError
	if !errors.As(err, &bindingErr) || bindingErr.Kind != "missing" {
		t.Fatalf("missing binding was accepted: %v", err)
	}
	assertBytesUnchanged(t, store.Path(), before)

	writeEffect := validBinding(sourceHash)
	writeEffect.WriteEffect = 1
	err = store.Append(boundEvidence("evidence/write-effect", writeEffect))
	if !errors.As(err, &bindingErr) || bindingErr.Kind != "write-effect" {
		t.Fatalf("nonzero write effect was accepted: %v", err)
	}
	assertBytesUnchanged(t, store.Path(), before)
}

func TestReplayDuplicateWithinBindingFails(t *testing.T) {
	sourceHash := strings.Repeat("1", 64)
	binding := validBinding(sourceHash)
	binding.Predecessors = []string{"predecessor/one", "predecessor/one"}
	store := New(filepath.Join(t.TempDir(), "evidence.jsonl"))
	err := store.Append(boundEvidence("evidence/duplicate-predecessor", binding))
	var bindingErr *BindingError
	if !errors.As(err, &bindingErr) || bindingErr.Kind != "predecessors" {
		t.Fatalf("duplicate predecessor was accepted: %v", err)
	}
	if _, err := os.Stat(store.Path()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed binding created a store: %v", err)
	}
}

func validBinding(sourceHash string) *RunBinding {
	head := strings.Repeat("1", 40)
	jobs := make(CanonicalJobs, 0, requiredCanonicalJobs)
	for index := 1; index <= requiredCanonicalJobs; index++ {
		jobs = append(jobs, JobReceipt{ID: "job-" + string(rune('0'+index)), Conclusion: "success", HeadSHA: head})
	}
	return &RunBinding{
		Repository:      "kimjooyoon/meta-ontology-go",
		Base:            "integration",
		Head:            head,
		EventRef:        "refs/pull/66/head",
		CheckoutRef:     head,
		RunID:           66,
		RunAttempt:      1,
		Workflow:        WorkflowIdentity{ID: "workflow-verify", Name: "verify", Path: ".github/workflows/verify.yml", Ref: "refs/heads/integration"},
		Jobs:            jobs,
		PolicyDigest:    strings.Repeat("2", 64),
		ToolchainDigest: strings.Repeat("3", 64),
		BundleDigest:    sourceHash,
		Predecessors:    []string{"predecessor/one"},
		EvidenceRefs:    []string{"evidence/input"},
	}
}

func boundEvidence(id string, binding *RunBinding) Evidence {
	return Evidence{
		ID:          id,
		Type:        VerificationEvidenceType,
		Subject:     "artifact/graph-proof-007",
		GeneratedBy: "activity/verify",
		Attributes:  map[string]string{"status": "passed"},
		Freshness:   NewFreshness(bindingSource(binding), time.Date(2026, 8, 12, 6, 0, 0, 0, time.UTC), time.Date(2026, 8, 12, 8, 0, 0, 0, time.UTC)),
		Binding:     binding,
	}
}

func bindingSource(binding *RunBinding) string {
	if binding == nil {
		return strings.Repeat("0", 64)
	}
	return binding.BundleDigest
}

func reverseJobs(jobs CanonicalJobs) {
	for left, right := 0, len(jobs)-1; left < right; left, right = left+1, right-1 {
		jobs[left], jobs[right] = jobs[right], jobs[left]
	}
}

func assertBytesUnchanged(t *testing.T, path string, before []byte) {
	t.Helper()
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("rejected append changed existing bytes")
	}
}
