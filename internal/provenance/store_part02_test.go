package provenance

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConflictingDuplicateAndTimestampIdentityAreRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "conflict.jsonl")
	store := New(path)
	record := testRecord("event/one", "semantic/one", StatusVerified)
	if err := store.Append(record); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := store.Read(ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	changed := replayed.Records[0]
	changed.GraphDigest = strings.Repeat("e", 64)
	if err := store.Append(changed); !errors.Is(err, ErrConflict) {
		t.Fatalf("conflicting digest was accepted: %v", err)
	}
	changed = replayed.Records[0]
	changed.Freshness.ProducedAt = "2026-08-14T01:00:00Z"
	if err := store.Append(changed); !errors.Is(err, ErrConflict) {
		t.Fatalf("timestamp-only identity mutation was accepted: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("rejected duplicate changed ledger bytes")
	}
}
func TestSourceMutationFreshnessFailure(t *testing.T) {
	store := New(filepath.Join(t.TempDir(), "freshness.jsonl"))
	if err := store.Append(testRecord("event/fresh", "semantic/fresh", StatusVerified)); err != nil {
		t.Fatal(err)
	}
	_, err := store.Read(ReadOptions{ExpectedSourceDigest: strings.Repeat("f", 64)})
	var freshness *FreshnessError
	if !errors.As(err, &freshness) || freshness.Kind != "source-mismatch" {
		t.Fatalf("source mutation was accepted: %v", err)
	}
}
func TestMalformedCommittedMetadataRejectsRepair(t *testing.T) {
	t.Run("unknown-field", testUnknownCommittedMetadata)
	t.Run("corrupt", testCorruptCommittedMetadata)
	t.Run("stale-digest", testStaleCommittedMetadata)
	t.Run("truncated", testTruncatedCommittedMetadata)
	t.Run("no-repair", testMalformedMetadataDoesNotRepair)
}
func testUnknownCommittedMetadata(t *testing.T) {
	path, data := validLedgerBytes(t)
	metadata := bytes.Replace(mustReadFile(t, manifestPath(path)), []byte(`{"schema"`), []byte(`{"authority":"inferred","schema"`), 1)
	if err := os.WriteFile(manifestPath(path), metadata, 0o644); err != nil {
		t.Fatal(err)
	}
	assertManifestError(t, path, "manifest-malformed", "unknown committed metadata field")
	assertLedgerUnchanged(t, path, data)
}
