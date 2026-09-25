package provenance

import (
	"encoding/base64"
	"fmt"
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
