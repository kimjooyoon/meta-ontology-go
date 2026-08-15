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

func TestBillingAppendCloseReopenAndReplayByteForByte(t *testing.T) {
	path := filepath.Join(t.TempDir(), "billing.jsonl")
	store := New(path)
	fixture := BillingFixture()
	if err := store.Append(fixture...); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := snapshot.Read(ReadOptions{ExpectedSourceDigest: strings.Repeat("a", 64), RequireFresh: true, Now: time.Date(2026, 8, 14, 1, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Records) != 2 || got.Records[1].Predecessor == nil || got.Records[1].Predecessor.ID != got.Records[0].ID {
		t.Fatalf("billing chain was not retained: %#v", got.Records)
	}
	if err := snapshot.Append(got.Records...); err != nil {
		t.Fatalf("exact duplicate replay was not idempotent: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("exact duplicate replay changed ledger bytes")
	}
}

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

func testCorruptCommittedMetadata(t *testing.T) {
	path, data := validLedgerBytes(t)
	metadata := mustReadFile(t, manifestPath(path))
	marker := []byte(`"data":"`)
	index := bytes.Index(metadata, marker)
	if index < 0 {
		t.Fatal("committed metadata data not found")
	}
	metadata[index+len(marker)] = '!'
	if err := os.WriteFile(manifestPath(path), metadata, 0o644); err != nil {
		t.Fatal(err)
	}
	assertManifestError(t, path, "commit-metadata-malformed", "corrupt committed metadata")
	assertLedgerUnchanged(t, path, data)
}

func testStaleCommittedMetadata(t *testing.T) {
	path, data := validLedgerBytes(t)
	metadata := mustReadFile(t, manifestPath(path))
	marker := []byte(`"digest":"`)
	index := bytes.Index(metadata, marker)
	if index < 0 {
		t.Fatal("committed metadata digest not found")
	}
	position := index + len(marker)
	if metadata[position] == '0' {
		metadata[position] = '1'
	} else {
		metadata[position] = '0'
	}
	if err := os.WriteFile(manifestPath(path), metadata, 0o644); err != nil {
		t.Fatal(err)
	}
	assertManifestError(t, path, "commit-metadata-stale", "stale committed metadata")
	assertLedgerUnchanged(t, path, data)
}

func testTruncatedCommittedMetadata(t *testing.T) {
	path, data := validLedgerBytes(t)
	metadata := mustReadFile(t, manifestPath(path))
	if err := os.WriteFile(manifestPath(path), metadata[:len(metadata)/2], 0o644); err != nil {
		t.Fatal(err)
	}
	assertManifestError(t, path, "manifest-malformed", "truncated committed metadata")
	assertLedgerUnchanged(t, path, data)
}

func testMalformedMetadataDoesNotRepair(t *testing.T) {
	path, _ := validLedgerBytes(t)
	metadata := bytes.Replace(mustReadFile(t, manifestPath(path)), []byte(`{"schema"`), []byte(`{"authority":"inferred","schema"`), 1)
	if err := os.WriteFile(manifestPath(path), metadata, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	assertManifestError(t, path, "manifest-malformed", "malformed metadata authorized repair")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("malformed metadata repair changed the ledger: %v", err)
	}
}

func assertManifestError(t *testing.T, path, kind, message string) {
	t.Helper()
	_, err := New(path).Read(ReadOptions{})
	var diagnostic *CorruptionError
	if !errors.As(err, &diagnostic) || diagnostic.Kind != kind {
		t.Fatalf("%s: %v", message, err)
	}
}

func assertLedgerUnchanged(t *testing.T, path string, expected []byte) {
	t.Helper()
	if !bytes.Equal(expected, mustReadFile(t, path)) {
		t.Fatal("metadata rejection changed the ledger")
	}
}

func TestChainGapAndCandidateDeferredCannotSatisfyVerifiedClaim(t *testing.T) {
	path := filepath.Join(t.TempDir(), "claims.jsonl")
	store := New(path)
	first := testRecord("event/candidate", "semantic/claim", StatusCandidate)
	if err := store.Append(first); err != nil {
		t.Fatal(err)
	}
	gap := testRecord("event/gap", "semantic/gap", StatusDeferred)
	gap.Predecessor = &DigestLink{ID: "event/not-tail", Digest: strings.Repeat("1", 64)}
	if err := store.Append(gap); !errors.Is(err, ErrChainGap) {
		t.Fatalf("chain gap was accepted: %v", err)
	}
	claim := VerifiedClaim{SemanticID: "semantic/claim", SemanticDigest: first.SemanticDigest, GraphDigest: first.GraphDigest}
	_, err := store.Verify(claim)
	var claimErr *ClaimError
	if !errors.As(err, &claimErr) || !errors.Is(err, ErrClaimNotVerified) {
		t.Fatalf("candidate evidence satisfied verified claim: %v", err)
	}
	verified := testRecord("event/verified", "semantic/claim", StatusVerified)
	verified.Predecessor = nil
	if err := store.Append(verified); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Verify(claim); err != nil {
		t.Fatalf("explicit verified evidence did not satisfy claim: %v", err)
	}
}

func validLedgerBytes(t *testing.T) (string, []byte) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ledger.jsonl")
	store := New(path)
	if err := store.Append(testRecord("event/a", "semantic/a", StatusVerified), testRecord("event/b", "semantic/b", StatusVerified)); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return path, data
}

func testRecord(id, semanticID string, status EvidenceStatus) Evidence {
	source := strings.Repeat("d", 64)
	return Evidence{
		ID: id, SemanticID: semanticID, Producer: "test://producer/store", Kind: KindVerification, Status: status,
		SourceSpan:   SourceSpan{URI: "examples/billing/main.gooo", Start: Position{Offset: 0, Line: 1, Column: 1}, End: Position{Offset: 10, Line: 1, Column: 11}},
		SourceDigest: source, SemanticDigest: strings.Repeat("e", 64), GraphDigest: strings.Repeat("f", 64),
		Freshness:  NewFreshness(source, time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 14, 2, 0, 0, 0, time.UTC)),
		Attributes: map[string]string{"fixture": "billing", "status": string(status)},
	}
}
