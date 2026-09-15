package provenance

import (
	"bytes"
	"os"
	"testing"
)

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
