package languagereadiness

import (
	"encoding/json"
	"fmt"
)

const conceptArtifactSchema = "gooo/language-concept-artifact/v1"

type conceptArtifact struct {
	Schema             string        `json:"schema"`
	Decision           string        `json:"decision"`
	CatalogDigest      string        `json:"catalog_digest"`
	Report             conceptReport `json:"report"`
	ReplayReportDigest string        `json:"replay_report_digest"`
	ReplayEqual        bool          `json:"replay_equal"`
	Bindings           bindingState  `json:"bindings"`
	RepositoryWrites   int           `json:"repository_writes"`
	ArtifactDigest     string        `json:"artifact_digest"`
}

type conceptReport struct {
	Concepts     []conceptEvidence `json:"concepts"`
	ReportDigest string            `json:"report_digest"`
}

type conceptEvidence struct {
	ID             string            `json:"id"`
	Stage          string            `json:"stage"`
	CodeBindings   []string          `json:"code_bindings"`
	MetricBindings []string          `json:"metric_bindings"`
	UseCases       []useCaseEvidence `json:"use_cases"`
}

type useCaseEvidence struct {
	ID              string `json:"id"`
	Trigger         string `json:"trigger"`
	ExpectedOutcome string `json:"expected_outcome"`
}

type bindingState struct {
	Missing     int `json:"missing"`
	Unsupported int `json:"unsupported"`
}

func decodeArtifact(raw []byte) (conceptArtifact, error) {
	var artifact conceptArtifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return conceptArtifact{}, fmt.Errorf("decode language concept artifact: %w", err)
	}
	return artifact, nil
}
