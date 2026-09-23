package lsp

import (
	"encoding/json"
)

func decodeDiagnosticProvenance(payload json.RawMessage) (diagnosticProvenanceObservation, error) {
	var value diagnosticProvenanceObservation
	if err := json.Unmarshal(payload, &value); err != nil {
		return diagnosticProvenanceObservation{}, err
	}
	if err := validateDiagnosticProvenance(value); err != nil {
		return diagnosticProvenanceObservation{}, err
	}
	return value, nil
}
