package repositorytopology

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type Report struct {
	Schema                    string            `json:"schema"`
	ExecutionPolicy           string            `json:"execution_policy"`
	Repository                string            `json:"repository"`
	CommitSHA                 string            `json:"commit_sha"`
	Status                    string            `json:"status"`
	Decision                  string            `json:"decision"`
	Resolution                string            `json:"resolution"`
	Reason                    string            `json:"reason"`
	Summary                   Summary           `json:"summary"`
	Views                     []AudienceView    `json:"views"`
	Files                     []FileMetric      `json:"files"`
	Directories               []DirectoryMetric `json:"directories"`
	Indicators                []Indicator       `json:"indicators"`
	Proofs                    []Proof           `json:"proofs"`
	Failures                  []string          `json:"failures"`
	SourceMetricsDigest       string            `json:"source_metrics_digest"`
	RootOntologyDigest        string            `json:"root_ontology_digest"`
	BindingOntologyDigest     string            `json:"binding_ontology_digest"`
	RepositoryWorkspaceWrites int               `json:"repository_workspace_writes"`
	MutationAuthority         bool              `json:"mutation_authority"`
	ReplayVerified            bool              `json:"replay_verified"`
	Digest                    string            `json:"digest"`
}

func seal(report *Report) {
	report.Digest = ""
	report.Digest = digestBytes(mustJSON(*report))
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func mustJSON(value Report) []byte {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return encoded
}
