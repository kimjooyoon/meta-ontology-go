package provenance

import (
	"bytes"
	"os"
	"testing"
)

func assertCommittedState(t *testing.T, store *Store, path string, expected []byte, lines int) {
	t.Helper()
	if data := mustReadFile(t, path); !bytes.Equal(data, expected) {
		t.Fatalf("committed bytes changed on retry: got %q want %q", data, expected)
	}
	snapshot, err := store.Read(ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Records) != lines || snapshot.Digest != digestBytes(expected) {
		t.Fatalf("committed snapshot is not canonical: lines=%d digest=%s", len(snapshot.Records), snapshot.Digest)
	}
	assertChain(t, snapshot.Records)
}
func assertChain(t *testing.T, records []Evidence) {
	t.Helper()
	for index, record := range records {
		if record.Sequence != uint64(index+1) {
			t.Fatalf("record %q has sequence %d", record.ID, record.Sequence)
		}
		if index == 0 {
			if record.Predecessor != nil {
				t.Fatalf("first record %q has a predecessor", record.ID)
			}
			continue
		}
		previous := records[index-1]
		if record.Predecessor == nil || record.Predecessor.ID != previous.ID || record.Predecessor.Digest != previous.Hash {
			t.Fatalf("record %q does not chain to %q", record.ID, previous.ID)
		}
	}
}
func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return data
}
