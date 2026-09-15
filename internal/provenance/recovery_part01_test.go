package provenance

import (
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
