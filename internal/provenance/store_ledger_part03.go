package provenance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

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
	} else if manifest.Base != nil {
		return ledgerManifest{}, corruption(path, 0, 0, "commit-metadata-incomplete", fmt.Errorf("committed manifest cannot contain a prepared base"))
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
