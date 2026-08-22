package artifactcoverage

const (
	ObservationSchema = "gooo/meta-operation-artifact-observations/v1"
	ReportSchema      = "gooo/meta-operation-artifact-coverage-report/v1"
)

type ArtifactObservation struct {
	Name          string   `json:"name"`
	HeadSHA       string   `json:"head_sha"`
	Digest        string   `json:"digest"`
	ReplayDigest  string   `json:"replay_digest"`
	EvidenceKeys []string `json:"evidence_keys"`
}

type ObservationDocument struct {
	Schema           string                `json:"schema"`
	CommitSHA        string                `json:"commit_sha"`
	Repository       string                `json:"repository"`
	RunID            int64                 `json:"run_id"`
	RunAttempt       int                   `json:"run_attempt"`
	RepositoryWrites int                   `json:"repository_writes"`
	Artifacts        []ArtifactObservation `json:"artifacts"`
}

type Summary struct {
	RequiredOperations             int `json:"required_operations"`
	CanonicalOperations            int `json:"canonical_operations"`
	UncoveredOperations            int `json:"uncovered_operations"`
	AmbiguousOperations            int `json:"ambiguous_operations"`
	ExactHeadOperations            int `json:"exact_head_operations"`
	DigestBoundOperations          int `json:"digest_bound_operations"`
	ReplayBoundOperations          int `json:"replay_bound_operations"`
	CanonicalCoverageBasisPoints   int `json:"canonical_coverage_basis_points"`
	ExactHeadCoverageBasisPoints   int `json:"exact_head_coverage_basis_points"`
	DigestBoundCoverageBasisPoints int `json:"digest_bound_coverage_basis_points"`
	ReplayBoundCoverageBasisPoints int `json:"replay_bound_coverage_basis_points"`
	RepositoryWrites               int `json:"repository_writes"`
}
