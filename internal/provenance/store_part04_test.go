package provenance

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
