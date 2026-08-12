package provenance

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCurrentContractIsExecutable(t *testing.T) {
	spec := CurrentContract()
	if spec.Version != ContractVersion || spec.Format == "" {
		t.Fatalf("contract identity is incomplete: %#v", spec)
	}
	for _, field := range append(append(spec.Input.RequiredFields, spec.Input.FreshnessFields...), spec.Input.BindingFields...) {
		if strings.TrimSpace(field) == "" {
			t.Fatal("contract contains an empty input field")
		}
	}
	if len(spec.Adapters) < 6 || len(spec.HostingStages) != 2 || len(spec.Hypotheses) < 6 || len(spec.NegativeCases) < 8 || len(spec.Deferred) < 5 {
		t.Fatalf("contract evidence plan is incomplete: %#v", spec)
	}
	for _, adapter := range spec.Adapters {
		if adapter.Name == "" || adapter.Input == "" || adapter.Output == "" || adapter.MustPreserve == "" {
			t.Fatalf("adapter boundary is incomplete: %#v", adapter)
		}
	}
	if spec.HostingStages[0].Status == "deferred" || spec.HostingStages[1].Status != "deferred" {
		t.Fatalf("hosting status claims are not explicit: %#v", spec.HostingStages)
	}
	for _, hypothesis := range spec.Hypotheses {
		if hypothesis.ID == "" || hypothesis.Claim == "" || hypothesis.Fixture == "" || hypothesis.PassCriterion == "" || hypothesis.FailCriterion == "" {
			t.Fatalf("hypothesis lacks falsifiable criteria: %#v", hypothesis)
		}
	}
}

func TestMinimalFixturePermutationInvariant(t *testing.T) {
	fixture := MinimalFixture()
	root := t.TempDir()
	leftStore := New(filepath.Join(root, "left", "evidence.jsonl"))
	rightStore := New(filepath.Join(root, "right", "evidence.jsonl"))
	if err := leftStore.Append(fixture.Records...); err != nil {
		t.Fatal(err)
	}
	if err := rightStore.Append(fixture.Records[1], fixture.Records[0]); err != nil {
		t.Fatal(err)
	}
	left, err := leftStore.Read(ReadOptions{ExpectedSourceHash: fixture.SourceHash, RequireFresh: true, Now: time.Date(2026, 8, 12, 1, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	right, err := rightStore.Read(ReadOptions{ExpectedSourceHash: fixture.SourceHash, RequireFresh: true, Now: time.Date(2026, 8, 12, 1, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	if left.Digest != right.Digest || !sameIDs(left.Records, fixture.ExpectedOrder) || !sameIDs(right.Records, fixture.ExpectedOrder) {
		t.Fatalf("H1 failed: left=%#v right=%#v expected=%#v", left, right, fixture.ExpectedOrder)
	}
	data, err := os.ReadFile(leftStore.Path())
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Count(data, []byte{'\n'})
	t.Logf("measurement fixture=%s records=%d lines=%d bytes=%d digest=%s", fixture.Name, len(left.Records), lines, len(data), left.Digest)
	if lines != len(fixture.Records) || len(data) == 0 || len(left.Digest) != 64 {
		t.Fatalf("fixture measurement is invalid: lines=%d bytes=%d digest=%q", lines, len(data), left.Digest)
	}
}

func TestNegativeTamperRejectsWithoutSnapshot(t *testing.T) {
	fixture := MinimalFixture()
	path := filepath.Join(t.TempDir(), "evidence.jsonl")
	store := New(path)
	if err := store.Append(fixture.Records...); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	marker := []byte(`"hash":"`)
	start := bytes.Index(before, marker)
	if start < 0 {
		t.Fatal("fixture line has no hash marker")
	}
	hashStart := start + len(marker)
	corrupted := append([]byte(nil), before...)
	copy(corrupted[hashStart:hashStart+64], strings.Repeat("0", 64))
	if err := os.WriteFile(path, corrupted, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = store.Read(ReadOptions{})
	var diagnostic *CorruptionError
	if !errors.As(err, &diagnostic) || diagnostic.Kind != "hash-mismatch" {
		t.Fatalf("H2 failed: %v", err)
	}
}

func TestNegativeDuplicatePreservesBytes(t *testing.T) {
	fixture := MinimalFixture()
	path := filepath.Join(t.TempDir(), "evidence.jsonl")
	store := New(path)
	if err := store.Append(fixture.Records...); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Append(fixture.Records[0]); err == nil {
		t.Fatal("H3 failed: duplicate append succeeded")
	} else if !errors.Is(err, ErrDuplicateID) {
		t.Fatalf("H3 returned the wrong error: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("H3 failed: duplicate append changed file bytes")
	}
}

func TestNegativeFreshnessRejectsStaleEvidence(t *testing.T) {
	fixture := MinimalFixture()
	store := New(filepath.Join(t.TempDir(), "evidence.jsonl"))
	if err := store.Append(fixture.Records...); err != nil {
		t.Fatal(err)
	}
	_, err := store.Read(ReadOptions{ExpectedSourceHash: strings.Repeat("2", 64)})
	var freshness *FreshnessError
	if !errors.As(err, &freshness) || freshness.Kind != "source-mismatch" {
		t.Fatalf("H4 source mismatch failed: %v", err)
	}
	_, err = store.Read(ReadOptions{RequireFresh: true, Now: time.Date(2026, 8, 12, 3, 0, 0, 0, time.UTC)})
	if !errors.As(err, &freshness) || freshness.Kind != "expired" {
		t.Fatalf("H4 expiry failed: %v", err)
	}
}

func BenchmarkMinimalFixtureRoundTrip(b *testing.B) {
	fixture := MinimalFixture()
	root := b.TempDir()
	b.ReportMetric(float64(len(fixture.Records)), "records/op")
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		store := New(filepath.Join(root, fmt.Sprintf("evidence-%d.jsonl", index)))
		if err := store.Append(fixture.Records...); err != nil {
			b.Fatal(err)
		}
		if _, err := store.Read(ReadOptions{}); err != nil {
			b.Fatal(err)
		}
	}
}

func sameIDs(records []Evidence, expected []string) bool {
	if len(records) != len(expected) {
		return false
	}
	for index, record := range records {
		if record.ID != expected[index] {
			return false
		}
	}
	return true
}
