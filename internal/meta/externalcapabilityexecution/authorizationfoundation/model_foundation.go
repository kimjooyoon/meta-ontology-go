package authorizationfoundation

type Foundation struct {
	Schema                 string `json:"schema"`
	Repository             string `json:"repository"`
	ProducerRunID          int64  `json:"producer_run_id"`
	ProducerRunAttempt     int    `json:"producer_run_attempt"`
	SubjectSHA             string `json:"subject_sha"`
	ArtifactID             int64  `json:"artifact_id"`
	ArtifactName           string `json:"artifact_name"`
	ArchiveDigest          string `json:"archive_digest"`
	ReceiptFileDigest      string `json:"receipt_file_digest"`
	ReceiptDigest          string `json:"receipt_digest"`
	PolicySourceDigest     string `json:"policy_source_digest"`
	PolicyGeneratedDigest  string `json:"policy_generated_digest"`
	BootstrapDecision      string `json:"bootstrap_decision"`
	BootstrapResolution    string `json:"bootstrap_resolution"`
	BootstrapUnknownStage  string `json:"bootstrap_unknown_stage"`
	BootstrapUnknownReason string `json:"bootstrap_unknown_reason"`
	RepositoryWrites       int    `json:"repository_writes"`
	MutationAuthority      bool   `json:"mutation_authority"`
	PromotionAuthority     bool   `json:"promotion_authority"`
}

type ArtifactMetadata struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Digest    string `json:"digest"`
	Expired   bool   `json:"expired"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at"`
}

type Input struct {
	ExpectedSubject string
	FoundationRaw   []byte
	MetadataRaw     []byte
	PriorReceiptRaw []byte
	CurrentRaw      []byte
}
