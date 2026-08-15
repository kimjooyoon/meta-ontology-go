package provenance

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func decodeManifestData(encoded string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode manifest data: %w", err)
	}
	return data, nil
}

func manifestPath(path string) string { return path + ".manifest" }

func ledgerManifestStateFor(data []byte, records []Evidence) ledgerManifestState {
	result := ledgerManifestState{Bytes: int64(len(data)), Lines: len(records), Digest: digestBytes(data), Data: base64.StdEncoding.EncodeToString(data)}
	if len(records) > 0 {
		result.LastID = records[len(records)-1].ID
		result.LastHash = records[len(records)-1].Hash
	}
	return result
}

func ledgerManifestFor(data []byte, records []Evidence) ledgerManifest {
	state := ledgerManifestStateFor(data, records)
	return ledgerManifest{Schema: SchemaVersion, Phase: manifestCommitted, Bytes: state.Bytes, Lines: state.Lines, Digest: state.Digest, LastID: state.LastID, LastHash: state.LastHash, Data: state.Data}
}

func preparedManifestFor(baseData []byte, baseRecords []Evidence, nextData []byte, nextRecords []Evidence) ledgerManifest {
	next := ledgerManifestFor(nextData, nextRecords)
	next.Phase = manifestPrepared
	base := ledgerManifestStateFor(baseData, baseRecords)
	next.Base = &base
	return next
}

func manifestStateMatches(manifest ledgerManifest, expected ledgerManifestState) bool {
	return manifest.Bytes == expected.Bytes && manifest.Lines == expected.Lines && manifest.Digest == expected.Digest && manifest.LastID == expected.LastID && manifest.LastHash == expected.LastHash && manifest.Data == expected.Data
}

type ledgerMaterializeFaultPoints struct {
	create, write, sync, close, rename, directorySync storageFaultPoint
}

func preparedRecoveryLedgerPoints() ledgerMaterializeFaultPoints {
	return ledgerMaterializeFaultPoints{
		write: faultRecoveryLedgerWrite, sync: faultRecoveryLedgerSync,
		close: faultRecoveryLedgerClose, rename: faultRecoveryLedgerRename,
		directorySync: faultRecoveryDirectorySync,
	}
}

func committedRepairLedgerPoints() ledgerMaterializeFaultPoints {
	return ledgerMaterializeFaultPoints{
		create: faultCommittedRepairCreate, write: faultCommittedRepairWrite,
		sync: faultCommittedRepairSync, close: faultCommittedRepairClose,
		rename: faultCommittedRepairRename, directorySync: faultCommittedRepairDirectory,
	}
}

type manifestFaultPoints struct {
	write, sync, close, rename, directorySync storageFaultPoint
}

func manifestPoints(manifest ledgerManifest, recovery bool) manifestFaultPoints {
	if recovery {
		return manifestFaultPoints{faultRecoveryManifestWrite, faultRecoveryManifestSync, faultRecoveryManifestClose, faultRecoveryManifestRename, faultRecoveryManifestDirectory}
	}
	if manifest.Phase == manifestPrepared {
		return manifestFaultPoints{faultPreparedWrite, faultPreparedSync, faultPreparedClose, faultPreparedRename, faultPreparedDirectorySync}
	}
	return manifestFaultPoints{faultCommittedWrite, faultCommittedSync, faultCommittedClose, faultCommittedRename, faultCommittedDirectorySync}
}

func writeManifest(path string, manifest ledgerManifest, recovery bool) error {
	directory := filepath.Dir(manifestPath(path))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create provenance manifest directory: %w", err)
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("marshal provenance manifest: %w", err)
	}
	payload = append(payload, '\n')
	points := manifestPoints(manifest, recovery)
	temporary, err := os.CreateTemp(directory, ".provenance-manifest-*")
	if err != nil {
		return fmt.Errorf("create provenance manifest: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := writeFullAt(temporary, payload, points.write); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write provenance manifest: %w", err)
	}
	if err := syncFile(temporary, points.sync); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync provenance manifest: %w", err)
	}
	if err := closeFile(temporary, points.close); err != nil {
		return fmt.Errorf("close provenance manifest: %w", err)
	}
	if err := failStorageOperation(points.rename); err != nil {
		return fmt.Errorf("commit provenance manifest: %w", err)
	}
	if err := os.Rename(temporaryName, manifestPath(path)); err != nil {
		return fmt.Errorf("commit provenance manifest: %w", err)
	}
	if err := syncDirectory(directory, points.directorySync); err != nil {
		return fmt.Errorf("sync provenance manifest directory: %w", err)
	}
	return nil
}

func materializeLedger(path string, data []byte, points ledgerMaterializeFaultPoints) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create provenance directory: %w", err)
	}
	if points.create != "" {
		if err := failStorageOperation(points.create); err != nil {
			return fmt.Errorf("create recovery ledger: %w", err)
		}
	}
	temporary, err := os.CreateTemp(directory, ".provenance-ledger-*")
	if err != nil {
		return fmt.Errorf("create recovery ledger: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := writeFullAt(temporary, data, points.write); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write recovery ledger: %w", err)
	}
	if err := syncFile(temporary, points.sync); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync recovery ledger: %w", err)
	}
	if err := closeFile(temporary, points.close); err != nil {
		return fmt.Errorf("close recovery ledger: %w", err)
	}
	if err := failStorageOperation(points.rename); err != nil {
		return fmt.Errorf("rename recovery ledger: %w", err)
	}
	if err := os.Rename(temporaryName, path); err != nil {
		return fmt.Errorf("rename recovery ledger: %w", err)
	}
	if err := syncDirectory(directory, points.directorySync); err != nil {
		return fmt.Errorf("sync recovery directory: %w", err)
	}
	return nil
}

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
