package provenance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	manifestCommitted = "committed"
	manifestPrepared  = "prepared"
)

type ledgerState struct {
	records []Evidence
	bytes   []byte
	digest  string
	lines   int
}

// ledgerManifest is the commit record for the JSONL materialization. A
// prepared record is a transaction: Base is the last committed state and the
// top-level fields describe the proposed append. A committed record is the
// only metadata state that authorizes the top-level bytes.
type ledgerManifest struct {
	Schema   int                  `json:"schema"`
	Phase    string               `json:"phase"`
	Bytes    int64                `json:"bytes"`
	Lines    int                  `json:"lines"`
	Digest   string               `json:"digest"`
	LastID   string               `json:"last_id,omitempty"`
	LastHash string               `json:"last_hash,omitempty"`
	Data     string               `json:"data"`
	Base     *ledgerManifestState `json:"base,omitempty"`
}

type ledgerManifestState struct {
	Bytes    int64  `json:"bytes"`
	Lines    int    `json:"lines"`
	Digest   string `json:"digest"`
	LastID   string `json:"last_id,omitempty"`
	LastHash string `json:"last_hash,omitempty"`
	Data     string `json:"data"`
}

// readLedger exposes only a committed state. The JSONL file is deliberately
// not authoritative by itself: a missing commit record rejects non-empty
// bytes, and a prepared record deterministically rolls back to Base.
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
			if len(data) != 0 {
				if _, parseErr := parseLedgerData(path, data); parseErr != nil {
					return ledgerState{}, parseErr
				}
			}
			return ledgerState{}, corruption(path, 0, 0, "ledger-mutation", fmt.Errorf("ledger bytes differ from committed metadata"))
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
	if err := materializeLedger(path, base.bytes); err != nil {
		return ledgerState{}, fmt.Errorf("recover provenance ledger: %w", err)
	}
	committed := ledgerManifestFor(base.bytes, base.records)
	if err := writeManifest(path, committed, true); err != nil {
		return ledgerState{}, fmt.Errorf("publish recovered provenance state: %w", err)
	}
	return base, nil
}

func readManifest(path string) (ledgerManifest, error) {
	manifestData, err := os.ReadFile(manifestPath(path))
	if err != nil {
		return ledgerManifest{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(manifestData))
	decoder.DisallowUnknownFields()
	var manifest ledgerManifest
	if err := decoder.Decode(&manifest); err != nil {
		return ledgerManifest{}, corruption(path, 0, 0, "manifest-malformed", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return ledgerManifest{}, corruption(path, 0, 0, "manifest-malformed", fmt.Errorf("manifest must contain one JSON value"))
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(manifestData, &fields); err != nil {
		return ledgerManifest{}, corruption(path, 0, 0, "manifest-malformed", err)
	}
	if _, ok := fields["phase"]; !ok {
		return ledgerManifest{}, corruption(path, 0, 0, "commit-metadata-incomplete", fmt.Errorf("manifest phase is required"))
	}
	dataField, ok := fields["data"]
	if !ok || bytes.Equal(bytes.TrimSpace(dataField), []byte("null")) {
		return ledgerManifest{}, corruption(path, 0, 0, "commit-metadata-incomplete", fmt.Errorf("manifest data is required"))
	}
	if manifest.Schema != SchemaVersion {
		return ledgerManifest{}, corruption(path, 0, 0, "commit-metadata-incomplete", fmt.Errorf("unsupported manifest schema %d", manifest.Schema))
	}
	if manifest.Phase != manifestCommitted && manifest.Phase != manifestPrepared {
		return ledgerManifest{}, corruption(path, 0, 0, "commit-metadata-incomplete", fmt.Errorf("unsupported commit phase %q", manifest.Phase))
	}
	if manifest.Phase == manifestPrepared {
		if _, ok := fields["base"]; !ok || manifest.Base == nil {
			return ledgerManifest{}, corruption(path, 0, 0, "commit-metadata-incomplete", fmt.Errorf("prepared manifest base is required"))
		}
	}
	return manifest, nil
}

func stateFromManifest(path string, manifest ledgerManifest) (ledgerState, error) {
	data, err := decodeManifestData(manifest.Data)
	if err != nil {
		return ledgerState{}, corruption(path, 0, 0, "commit-metadata-malformed", err)
	}
	state, err := parseLedgerData(path, data)
	if err != nil {
		return ledgerState{}, corruption(path, 0, 0, "commit-metadata-invalid", err)
	}
	expected := ledgerManifestStateFor(data, state.records)
	if !manifestStateMatches(manifest, expected) {
		return ledgerState{}, corruption(path, 0, 0, "commit-metadata-stale", fmt.Errorf("manifest summary does not match canonical data"))
	}
	return state, nil
}

func stateFromSummary(path string, summary ledgerManifestState) (ledgerState, error) {
	data, err := decodeManifestData(summary.Data)
	if err != nil {
		return ledgerState{}, corruption(path, 0, 0, "commit-metadata-malformed", err)
	}
	state, err := parseLedgerData(path, data)
	if err != nil {
		return ledgerState{}, corruption(path, 0, 0, "commit-metadata-invalid", err)
	}
	expected := ledgerManifestStateFor(data, state.records)
	if summary != expected {
		return ledgerState{}, corruption(path, 0, 0, "commit-metadata-stale", fmt.Errorf("base summary does not match canonical data"))
	}
	return state, nil
}
