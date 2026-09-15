package provenance

import (
	"bytes"
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
