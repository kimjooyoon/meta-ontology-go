package languageartifactoracle

import (
	"encoding/json"
	"fmt"
)

func decodeArtifact(raw []byte) (sourceArtifact, error) {
	var artifact sourceArtifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return sourceArtifact{}, fmt.Errorf("decode source artifact: %w", err)
	}
	return artifact, nil
}

func DecodeIndependence(raw []byte) (IndependenceEvidence, error) {
	var evidence IndependenceEvidence
	if err := json.Unmarshal(raw, &evidence); err != nil {
		return IndependenceEvidence{}, err
	}
	if evidence.Schema != IndependenceSchema || evidence.ProducerDependencies < 0 {
		return IndependenceEvidence{}, fmt.Errorf("ARTIFACT_ORACLE_INDEPENDENCE_UNKNOWN")
	}
	return evidence, nil
}

func legacyAccepted(raw []byte) bool {
	var value struct {
		Decision string `json:"decision"`
		Summary  struct {
			CasesSatisfied int `json:"cases_satisfied"`
			CasesTotal     int `json:"cases_total"`
		} `json:"summary"`
	}
	return json.Unmarshal(raw, &value) == nil && value.Decision == "PASS" &&
		value.Summary.CasesSatisfied == CaseTotal && value.Summary.CasesTotal == CaseTotal
}
