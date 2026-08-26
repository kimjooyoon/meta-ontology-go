package provenance

import (
	"fmt"
)

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
