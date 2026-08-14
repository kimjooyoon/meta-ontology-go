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

func TestUnknownMalformedCorruptReorderedAndTruncatedLedgerReject(t *testing.T) {
	t.Run("unknown-field", func(t *testing.T) {
		path, data := validLedgerBytes(t)
		mutated := bytes.Replace(data, []byte(`{"schema"`), []byte(`{"authority":"inferred","schema"`), 1)
		if err := os.WriteFile(path, mutated, 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := New(path).Read(ReadOptions{})
		var diagnostic *CorruptionError
		if !errors.As(err, &diagnostic) || diagnostic.Kind != "invalid-json" {
			t.Fatalf("unknown field was accepted: %v", err)
		}
	})

	t.Run("corrupt", func(t *testing.T) {
		path, data := validLedgerBytes(t)
		marker := []byte(`"source_digest":"`)
		index := bytes.Index(data, marker)
		if index < 0 {
			t.Fatal("golden source digest not found")
		}
		position := index + len(marker)
		data[position] = 'f'
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := New(path).Read(ReadOptions{})
		var diagnostic *CorruptionError
		if !errors.As(err, &diagnostic) || diagnostic.Kind != "hash-mismatch" {
			t.Fatalf("corrupt record was accepted: %v", err)
		}
	})

	t.Run("reordered", func(t *testing.T) {
		path, data := validLedgerBytes(t)
		lines := bytes.Split(data, []byte{'\n'})
		lines[0], lines[1] = lines[1], lines[0]
		if err := os.WriteFile(path, bytes.Join(lines, []byte{'\n'}), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := New(path).Read(ReadOptions{})
		var diagnostic *CorruptionError
		if !errors.As(err, &diagnostic) || diagnostic.Kind != "chain-gap" {
			t.Fatalf("reordered ledger was accepted: %v", err)
		}
	})

	t.Run("truncated", func(t *testing.T) {
		path, data := validLedgerBytes(t)
		lastLine := bytes.LastIndex(data[:len(data)-1], []byte{'\n'}) + 1
		if err := os.WriteFile(path, data[:lastLine], 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := New(path).Read(ReadOptions{})
		var diagnostic *CorruptionError
		if !errors.As(err, &diagnostic) || diagnostic.Kind != "ledger-mutation" {
			t.Fatalf("truncated ledger was accepted: %v", err)
		}
	})
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
