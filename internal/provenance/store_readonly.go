package provenance

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ErrReadOnlyRecoveryRequired means observation cannot proceed without a
// separate recovery-capable operation. ReadOnly never performs that operation.
var ErrReadOnlyRecoveryRequired = errors.New("provenance read-only observation requires explicit recovery")

const (
	ReadOnlyPreparedTransaction     = "PREPARED_TRANSACTION_REQUIRES_RECOVERY"
	ReadOnlyMaterializationMismatch = "COMMITTED_MATERIALIZATION_REQUIRES_REPAIR"
)

// ReadOnlyRecoveryError preserves the recovery cause without granting writes.
type ReadOnlyRecoveryError struct {
	Path   string
	Reason string
}

func (err *ReadOnlyRecoveryError) Error() string {
	return fmt.Sprintf("%s: %s (%s)", ErrReadOnlyRecoveryRequired, err.Reason, err.Path)
}

func (err *ReadOnlyRecoveryError) Unwrap() error { return ErrReadOnlyRecoveryRequired }

// ReadOnly validates an exact committed physical ledger without repair,
// rollback, creation, or append. Unlike Read, it never materializes metadata.
// Candidate and deferred records remain observations, not verified claims.
func (s *Store) ReadOnly(options ReadOptions) (Snapshot, error) {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return Snapshot{}, fmt.Errorf("provenance store path is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	lock := lockForPath(s.path)
	lock.Lock()
	defer lock.Unlock()
	state, err := readLedgerWithoutRecovery(s.path)
	if err != nil {
		return Snapshot{}, err
	}
	if err := checkReadOptions(state.records, options); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Records: state.records, Digest: state.digest}, nil
}

func readLedgerWithoutRecovery(path string) (ledgerState, error) {
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
			return ledgerState{}, &ReadOnlyRecoveryError{Path: path, Reason: ReadOnlyMaterializationMismatch}
		}
		return state, nil
	case manifestPrepared:
		return rejectPreparedReadOnly(path, manifest)
	default:
		return ledgerState{}, corruption(path, 0, 0, "commit-metadata-incomplete", fmt.Errorf("unsupported commit phase %q", manifest.Phase))
	}
}

func rejectPreparedReadOnly(path string, manifest ledgerManifest) (ledgerState, error) {
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
	return ledgerState{}, &ReadOnlyRecoveryError{Path: path, Reason: ReadOnlyPreparedTransaction}
}
