package provenance

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendReadCanonicalOrderingAndDigest(t *testing.T) {
	root := t.TempDir()
	sourceHash := strings.Repeat("a", 64)
	first := New(filepath.Join(root, "first", "evidence.jsonl"))
	second := New(filepath.Join(root, "second", "evidence.jsonl"))
	older := testEvidence("evidence/b", sourceHash, time.Date(2026, 8, 12, 1, 2, 3, 0, time.UTC), time.Time{})
	newer := testEvidence("evidence/a", sourceHash, time.Date(2026, 8, 12, 1, 2, 4, 0, time.UTC), time.Time{})
	if err := first.Append(older, newer); err != nil {
		t.Fatal(err)
	}
	if err := second.Append(newer); err != nil {
		t.Fatal(err)
	}
	if err := second.Append(older); err != nil {
		t.Fatal(err)
	}
	left, err := first.Read(ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	right, err := second.Read(ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if left.Digest != right.Digest {
		t.Fatalf("append order changed digest: %s != %s", left.Digest, right.Digest)
	}
	if len(left.Digest) != 64 || strings.ToLower(left.Digest) != left.Digest {
		t.Fatalf("snapshot digest is not canonical: %q", left.Digest)
	}
	if len(left.Records) != 2 || left.Records[0].ID != "evidence/a" || left.Records[1].ID != "evidence/b" {
		t.Fatalf("records were not returned in stable order: %#v", left.Records)
	}
	for _, record := range left.Records {
		if len(record.Hash) != 64 || strings.ToLower(record.Hash) != record.Hash {
			t.Fatalf("record hash is not canonical: %q", record.Hash)
		}
	}
}

func TestAppendRejectsDuplicateIDs(t *testing.T) {
	store := New(filepath.Join(t.TempDir(), "evidence.jsonl"))
	record := testEvidence("evidence/one", strings.Repeat("b", 64), time.Date(2026, 8, 12, 2, 0, 0, 0, time.UTC), time.Time{})
	if err := store.Append(record); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(record); err == nil {
		t.Fatal("duplicate evidence ID was accepted")
	}
	snapshot, err := store.Read(ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Records) != 1 {
		t.Fatalf("duplicate append changed store: %#v", snapshot.Records)
	}
}

func TestReadReportsCorruptionDiagnostics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "evidence.jsonl")
	store := New(path)
	if err := store.Append(testEvidence("evidence/one", strings.Repeat("c", 64), time.Date(2026, 8, 12, 3, 0, 0, 0, time.UTC), time.Time{})); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	prefix := "\"hash\":\""
	start := strings.Index(string(data), prefix)
	if start < 0 {
		t.Fatal("stored record did not contain a hash")
	}
	hashStart := start + len(prefix)
	corrupted := string(data[:hashStart]) + strings.Repeat("0", 64) + string(data[hashStart+64:])
	if err := os.WriteFile(path, []byte(corrupted), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = store.Read(ReadOptions{})
	var diagnostic *CorruptionError
	if !errors.As(err, &diagnostic) {
		t.Fatalf("expected corruption diagnostic, got %v", err)
	}
	if diagnostic.Line != 1 || diagnostic.Offset != 0 || diagnostic.Kind != "hash-mismatch" {
		t.Fatalf("unexpected corruption diagnostic: %#v", diagnostic)
	}
	if !strings.Contains(diagnostic.Error(), path) {
		t.Fatalf("diagnostic omitted path: %v", diagnostic)
	}
}

func TestReadReportsLineSyntaxDiagnostics(t *testing.T) {
	cases := []struct {
		name string
		data string
		kind string
	}{
		{name: "invalid-json", data: "not-json\n", kind: "invalid-json"},
		{name: "missing-newline", data: "not-json", kind: "missing-newline"},
		{name: "blank-line", data: "\n", kind: "blank-line"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "evidence.jsonl")
			if err := os.WriteFile(path, []byte(testCase.data), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := New(path).Read(ReadOptions{})
			var diagnostic *CorruptionError
			if !errors.As(err, &diagnostic) || diagnostic.Kind != testCase.kind {
				t.Fatalf("expected %s diagnostic, got %v", testCase.kind, err)
			}
		})
	}
}

func TestReadReportsDuplicateIDDiagnostic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "evidence.jsonl")
	store := New(path)
	if err := store.Append(testEvidence("evidence/duplicate", strings.Repeat("f", 64), time.Date(2026, 8, 12, 4, 0, 0, 0, time.UTC), time.Time{})); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, data...), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = store.Read(ReadOptions{})
	var diagnostic *CorruptionError
	if !errors.As(err, &diagnostic) || diagnostic.Kind != "duplicate-id" || diagnostic.Line != 2 || diagnostic.Offset != int64(len(data)) {
		t.Fatalf("unexpected duplicate diagnostic: %v", err)
	}
}

func TestFreshnessMetadataAndChecks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "evidence.jsonl")
	sourceHash := strings.Repeat("d", 64)
	produced := time.Date(2026, 8, 12, 0, 0, 0, 0, time.FixedZone("KST", 9*60*60))
	validUntil := produced.Add(2 * time.Hour)
	store := New(path)
	if err := store.Append(testEvidence("evidence/fresh", sourceHash, produced, validUntil)); err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.Read(ReadOptions{ExpectedSourceHash: sourceHash, RequireFresh: true, Now: produced.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Records[0].Freshness.ProducedAt != "2026-08-11T15:00:00Z" {
		t.Fatalf("timestamp was not normalized: %#v", snapshot.Records[0].Freshness)
	}
	_, err = store.Read(ReadOptions{ExpectedSourceHash: strings.Repeat("e", 64)})
	var freshness *FreshnessError
	if !errors.As(err, &freshness) || freshness.Kind != "source-mismatch" {
		t.Fatalf("source staleness was not reported: %v", err)
	}
	_, err = store.Read(ReadOptions{RequireFresh: true, Now: validUntil.Add(time.Nanosecond)})
	if !errors.As(err, &freshness) || freshness.Kind != "expired" {
		t.Fatalf("expiry staleness was not reported: %v", err)
	}
}

func testEvidence(id, sourceHash string, produced, validUntil time.Time) Evidence {
	return Evidence{
		ID:          id,
		Type:        "TestResult",
		Subject:     "artifact/billing",
		GeneratedBy: "activity/verify",
		Attributes:  map[string]string{"status": "passed"},
		Freshness:   NewFreshness(sourceHash, produced, validUntil),
	}
}
