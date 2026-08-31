package toolchainconformance

import (
	"encoding/json"
	"strings"
)

type artifactEnvelope struct {
	Schema     string `json:"schema"`
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	HeadSHA    string `json:"head_sha"`
	Source     struct {
		ExpectedHeadSHA string `json:"expected_head_sha"`
	} `json:"source"`
	Summary    map[string]json.RawMessage `json:"summary"`
	Indicators []struct {
		Satisfied bool `json:"satisfied"`
	} `json:"indicators"`
	Proofs []struct {
		Passed bool `json:"passed"`
	} `json:"proofs"`
	RepositoryWrites   int    `json:"repository_writes"`
	MutationAuthorized bool   `json:"mutation_authorized"`
	ReportDigest       string `json:"report_digest"`
}

func decodeEnvelope(raw []byte) (artifactEnvelope, error) {
	envelope := artifactEnvelope{}
	err := json.Unmarshal(raw, &envelope)
	return envelope, err
}

func summaryInteger(summary map[string]json.RawMessage, key string) (int, bool) {
	for candidate, raw := range summary {
		if !strings.EqualFold(candidate, key) {
			continue
		}
		value := 0
		if json.Unmarshal(raw, &value) != nil {
			return 0, false
		}
		return value, true
	}
	return 0, false
}
