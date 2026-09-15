package provenance

import (
	"path/filepath"
	"testing"
)

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
