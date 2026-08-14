package provenance

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type appendFaultCase struct {
	name  string
	point storageFaultPoint
}

func appendFaultCases() []appendFaultCase {
	return []appendFaultCase{
		{name: "ledger-write", point: faultLedgerAppendWrite},
		{name: "ledger-sync", point: faultLedgerAppendSync},
		{name: "ledger-close", point: faultLedgerAppendClose},
		{name: "prepared-write", point: faultPreparedWrite},
		{name: "prepared-sync", point: faultPreparedSync},
		{name: "prepared-close", point: faultPreparedClose},
		{name: "prepared-rename", point: faultPreparedRename},
		{name: "prepared-directory-sync", point: faultPreparedDirectorySync},
		{name: "committed-write", point: faultCommittedWrite},
		{name: "committed-sync", point: faultCommittedSync},
		{name: "committed-close", point: faultCommittedClose},
		{name: "committed-rename", point: faultCommittedRename},
		{name: "committed-directory-sync", point: faultCommittedDirectorySync},
	}
}

func recoveryFaultPoints() []storageFaultPoint {
	return []storageFaultPoint{
		faultRecoveryLedgerWrite,
		faultRecoveryLedgerSync,
		faultRecoveryLedgerClose,
		faultRecoveryLedgerRename,
		faultRecoveryDirectorySync,
		faultRecoveryManifestWrite,
		faultRecoveryManifestSync,
		faultRecoveryManifestClose,
		faultRecoveryManifestRename,
		faultRecoveryManifestDirectory,
	}
}

func TestFirstMultiRecordAppendFaultsNeverExposePartialAuthority(t *testing.T) {
	records := BillingFixture()
	post := expectedLedgerBytes(t, records)
	partial := firstLineEnd(post)
	for _, test := range appendFaultCases() {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "first.jsonl")
			store := New(path)
			restore := installStorageFaultForTest(test.point, partialFor(test.point, partial))
			err := store.Append(records...)
			restore()
			if err == nil {
				t.Fatal("faulted first append unexpectedly succeeded")
			}
			reopened, openErr := Open(path)
			if openErr != nil {
				t.Fatalf("restart did not recover a committed state: %v", openErr)
			}
			assertPossibleState(t, reopened, path, nil, post, 0, len(records))
			if err := reopened.Append(records...); err != nil {
				t.Fatalf("retry did not commit the complete batch: %v", err)
			}
			assertCommittedState(t, reopened, path, post, len(records))
		})
	}
}

func TestLaterAppendFaultsPreserveCommittedBaseOrCompletePostState(t *testing.T) {
	first := []Evidence{testRecord("event/base", "semantic/base", StatusVerified)}
	second := []Evidence{
		testRecord("event/second-a", "semantic/second-a", StatusVerified),
		testRecord("event/second-b", "semantic/second-b", StatusVerified),
	}
	base := expectedLedgerBytes(t, first)
	post := expectedLedgerBytes(t, append(append([]Evidence(nil), first...), second...))
	partial := firstLineEnd(post[len(base):])
	for _, test := range appendFaultCases() {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "later.jsonl")
			store := New(path)
			if err := store.Append(first...); err != nil {
				t.Fatal(err)
			}
			restore := installStorageFaultForTest(test.point, partialFor(test.point, partial))
			err := store.Append(second...)
			restore()
			if err == nil {
				t.Fatal("faulted later append unexpectedly succeeded")
			}
			reopened, openErr := Open(path)
			if openErr != nil {
				t.Fatalf("restart did not recover a committed state: %v", openErr)
			}
			assertPossibleState(t, reopened, path, base, post, len(first), len(first)+len(second))
			if err := reopened.Append(second...); err != nil {
				t.Fatalf("retry did not converge on the complete post-state: %v", err)
			}
			assertCommittedState(t, reopened, path, post, len(first)+len(second))
		})
	}
}

func TestPreparedRecoveryFaultsFailClosedThenRetry(t *testing.T) {
	records := BillingFixture()
	post := expectedLedgerBytes(t, records)
	partial := firstLineEnd(post)
	for _, point := range recoveryFaultPoints() {
		t.Run(string(point), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "recovery.jsonl")
			store := New(path)
			restoreAppend := installStorageFaultForTest(faultLedgerAppendWrite, partial)
			if err := store.Append(records...); err == nil {
				t.Fatal("setup fault did not fail")
			}
			restoreAppend()
			restoreRecovery := installStorageFaultForTest(point, 0)
			if _, err := Open(path); err == nil {
				t.Fatal("faulted recovery exposed a snapshot")
			}
			restoreRecovery()
			reopened, err := Open(path)
			if err != nil {
				t.Fatalf("recovery did not converge after retry: %v", err)
			}
			assertPossibleState(t, reopened, path, nil, post, 0, len(records))
			if err := reopened.Append(records...); err != nil {
				t.Fatalf("retry after recovery failure failed: %v", err)
			}
			assertCommittedState(t, reopened, path, post, len(records))
		})
	}
}

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
