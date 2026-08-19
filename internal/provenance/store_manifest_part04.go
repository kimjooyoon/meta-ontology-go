package provenance

import (
	"bytes"
	"fmt"
	"os"
)

// repairCommittedLedger treats validated committed metadata as an exact
// recovery image. The replacement is durable before the post-state is
// reread, parsed, and compared byte-for-byte; any barrier failure returns an
// error so callers never observe an unvalidated partial repair.
func repairCommittedLedger(path string, expected []byte) (ledgerState, error) {
	if err := materializeLedger(path, expected, committedRepairLedgerPoints()); err != nil {
		return ledgerState{}, fmt.Errorf("repair committed provenance ledger: %w", err)
	}
	if err := failStorageOperation(faultCommittedRepairRevalidate); err != nil {
		return ledgerState{}, fmt.Errorf("revalidate committed provenance ledger: %w", err)
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		return ledgerState{}, fmt.Errorf("re-read committed provenance ledger: %w", err)
	}
	if !bytes.Equal(actual, expected) {
		return ledgerState{}, corruption(path, 0, 0, "repair-mismatch", fmt.Errorf("repaired ledger differs from committed bytes"))
	}
	state, err := parseLedgerData(path, actual)
	if err != nil {
		return ledgerState{}, fmt.Errorf("revalidate committed provenance records: %w", err)
	}
	if state.digest != digestBytes(expected) {
		return ledgerState{}, corruption(path, 0, 0, "repair-mismatch", fmt.Errorf("repaired ledger digest differs from committed digest"))
	}
	return state, nil
}
func syncDirectory(directory string, point storageFaultPoint) error {
	if err := failStorageOperation(point); err != nil {
		return err
	}
	directoryFile, err := os.Open(directory)
	if err != nil {
		return err
	}
	if err := directoryFile.Sync(); err != nil {
		_ = directoryFile.Close()
		return err
	}
	return directoryFile.Close()
}
