package provenance

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNonEmptyLedgerRequiresCompleteCommitMetadata(t *testing.T) {
	path, data := validLedgerBytes(t)
	if err := os.Remove(manifestPath(path)); err != nil {
		t.Fatal(err)
	}
	_, err := New(path).Read(ReadOptions{})
	var diagnostic *CorruptionError
	if !errors.As(err, &diagnostic) || diagnostic.Kind != "commit-metadata-missing" {
		t.Fatalf("non-empty ledger without commit metadata was accepted: %v", err)
	}
	if err := os.WriteFile(manifestPath(path), []byte("{\"schema\":1,\"phase\":\"committed\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = New(path).Read(ReadOptions{})
	if !errors.As(err, &diagnostic) || diagnostic.Kind != "commit-metadata-incomplete" {
		t.Fatalf("incomplete commit metadata was accepted: %v", err)
	}
	if !bytes.Equal(data, mustReadFile(t, path)) {
		t.Fatal("metadata rejection changed the append-only ledger")
	}
}
func partialFor(point storageFaultPoint, partial int) int {
	if point == faultLedgerAppendWrite {
		return partial
	}
	return 0
}
func expectedLedgerBytes(t *testing.T, records []Evidence) []byte {
	t.Helper()
	path := filepath.Join(t.TempDir(), "expected.jsonl")
	if err := New(path).Append(records...); err != nil {
		t.Fatal(err)
	}
	return mustReadFile(t, path)
}
func firstLineEnd(data []byte) int {
	return bytes.IndexByte(data, '\n') + 1
}
func assertPossibleState(t *testing.T, store *Store, path string, base, post []byte, baseLines, postLines int) {
	t.Helper()
	snapshot, err := store.Read(ReadOptions{})
	if err != nil {
		t.Fatalf("recovered state was not readable: %v", err)
	}
	data := mustReadFile(t, path)
	if !bytes.Equal(data, base) && !bytes.Equal(data, post) {
		t.Fatalf("recovery exposed bytes from neither transaction boundary: %q", data)
	}
	if len(snapshot.Records) != baseLines && len(snapshot.Records) != postLines {
		t.Fatalf("recovery exposed an invalid record count: %d", len(snapshot.Records))
	}
	assertChain(t, snapshot.Records)
}
