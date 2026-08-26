package userjourneyscorecard

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
)

type Report struct {
	Schema                             string         `json:"schema"`
	Status                             string         `json:"status"`
	Decision                           string         `json:"decision"`
	Resolution                         string         `json:"resolution"`
	Reason                             string         `json:"reason"`
	Source                             EvidenceSource `json:"source"`
	Summary                            Summary        `json:"summary"`
	Journeys                           []JourneyStats `json:"journeys"`
	Views                              []AudienceView `json:"views"`
	Indicators                         []Indicator    `json:"indicators"`
	Proofs                             []Proof        `json:"proofs"`
	Failures                           []string       `json:"failures"`
	NotClaimed                         []string       `json:"not_claimed"`
	ContractDigest                     string         `json:"contract_digest"`
	UpstreamDigest                     string         `json:"upstream_digest"`
	ProfileDigest                      string         `json:"profile_digest"`
	ResourceObservationMode            string         `json:"resource_observation_mode"`
	ResourceMeasurementReplayAuthority bool           `json:"resource_measurement_replay_authority"`
	ReducerReplayVerified              bool           `json:"reducer_replay_verified"`
	RepositoryWrites                   int            `json:"repository_writes"`
	MutationAuthority                  bool           `json:"mutation_authority"`
	Digest                             string         `json:"digest"`
}

type EvidenceSource struct {
	ExpectedHeadSHA string     `json:"expected_head_sha"`
	Runner          Runner     `json:"runner"`
	Executable      Executable `json:"executable"`
	SourcePath      string     `json:"source_path"`
	SourceDigest    string     `json:"source_digest"`
}

func seal(report *Report) {
	report.Digest = ""
	report.Digest = digestJSON(*report)
}

func digestJSON(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestFile(filename string) (string, int64, error) {
	raw, err := os.ReadFile(filename)
	return digestBytes(raw), int64(len(raw)), err
}

func mustJSON(report Report) []byte {
	raw, err := json.Marshal(report)
	if err != nil {
		panic(err)
	}
	return raw
}
