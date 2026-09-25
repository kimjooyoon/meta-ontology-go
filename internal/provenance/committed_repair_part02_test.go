package provenance

import (
	"bytes"
	"os"
	"testing"
)

func TestCommittedRepairBarriersFailClosedThenRecover(t *testing.T) {
	points := []committedRepairFault{
		{name: "temp-create", point: faultCommittedRepairCreate},
		{name: "temp-write-partial-line", point: faultCommittedRepairWrite, partial: true},
		{name: "temp-sync", point: faultCommittedRepairSync},
		{name: "temp-close", point: faultCommittedRepairClose},
		{name: "rename", point: faultCommittedRepairRename},
		{name: "directory-sync", point: faultCommittedRepairDirectory},
		{name: "revalidate", point: faultCommittedRepairRevalidate},
	}
	for _, test := range points {
		t.Run(test.name, func(t *testing.T) {
			path, expected := validLedgerBytes(t)
			divergent := append([]byte(nil), expected[:firstLineEnd(expected)]...)
			if err := os.WriteFile(path, divergent, 0o644); err != nil {
				t.Fatal(err)
			}
			partial := 0
			if test.partial {
				partial = firstLineEnd(expected)
			}
			restore := installStorageFaultForTest(test.point, partial)
			if _, err := Open(path); err == nil {
				t.Fatal("faulted committed repair returned success")
			}
			restore()
			actual := mustReadFile(t, path)
			if !bytes.Equal(actual, divergent) && !bytes.Equal(actual, expected) {
				t.Fatal("faulted repair left partial physical bytes")
			}
			reopened, err := Open(path)
			if err != nil {
				t.Fatalf("restart did not deterministically repair committed ledger: %v", err)
			}
			assertCommittedState(t, reopened, path, expected, 2)
		})
	}
}
func openedRecords(t *testing.T, store *Store) []Evidence {
	t.Helper()
	snapshot, err := store.Read(ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return snapshot.Records
}
