package provenance

import (
	"bytes"
	"fmt"
	"os"
)

// readLedger exposes only a committed state. The JSONL file is deliberately
// not authoritative by itself: a missing commit record rejects non-empty
// bytes, a prepared record deterministically rolls back to Base, and a fully
// validated committed record repairs any divergent materialization from its
// exact post-state image.
func readLedger(path string) (ledgerState, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		data = nil
	} else if err != nil {
		return ledgerState{}, fmt.Errorf("open provenance store: %w", err)
	}

	manifest, err := readManifest(path)
	if os.IsNotExist(err) {
		if len(data) != 0 {
			return ledgerState{}, corruption(path, 0, 0, "commit-metadata-missing", fmt.Errorf("non-empty ledger has no committed metadata"))
		}
		return parseLedgerData(path, data)
	}
	if err != nil {
		return ledgerState{}, err
	}

	switch manifest.Phase {
	case manifestCommitted:
		state, err := stateFromManifest(path, manifest)
		if err != nil {
			return ledgerState{}, err
		}
		if !bytes.Equal(data, state.bytes) {
			return repairCommittedLedger(path, state.bytes)
		}
		return state, nil
	case manifestPrepared:
		return recoverPrepared(path, manifest)
	default:
		return ledgerState{}, corruption(path, 0, 0, "commit-metadata-incomplete", fmt.Errorf("unsupported commit phase %q", manifest.Phase))
	}
}
func recoverPrepared(path string, manifest ledgerManifest) (ledgerState, error) {
	if manifest.Base == nil {
		return ledgerState{}, corruption(path, 0, 0, "commit-metadata-incomplete", fmt.Errorf("prepared metadata has no base state"))
	}
	base, err := stateFromSummary(path, *manifest.Base)
	if err != nil {
		return ledgerState{}, err
	}
	next, err := stateFromManifest(path, manifest)
	if err != nil {
		return ledgerState{}, err
	}
	if !bytes.HasPrefix(next.bytes, base.bytes) {
		return ledgerState{}, corruption(path, 0, 0, "commit-metadata-mutation", fmt.Errorf("prepared state does not append to its base"))
	}
	if err := materializeLedger(path, base.bytes, preparedRecoveryLedgerPoints()); err != nil {
		return ledgerState{}, fmt.Errorf("recover provenance ledger: %w", err)
	}
	committed := ledgerManifestFor(base.bytes, base.records)
	if err := writeManifest(path, committed, true); err != nil {
		return ledgerState{}, fmt.Errorf("publish recovered provenance state: %w", err)
	}
	return base, nil
}
