package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"

const foundationPromotionEvidenceSchema = "gooo/foundation-promotion-evidence/v1"

var foundationPromotionJobNames = []string{"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy", "CI evidence"}

type foundationPromotionReceipt struct {
	Schema             string                                  `json:"schema"`
	Report             feedbackpredecessor.Report              `json:"report"`
	ProofChoice        string                                  `json:"proof_choice,omitempty"`
	Foundation         *feedbackpredecessor.FoundationEvidence `json:"foundation,omitempty"`
	ReplayReportDigest string                                  `json:"replay_report_digest"`
	ExpectedDigest     string                                  `json:"expected_digest,omitempty"`
	ReplayVerified     bool                                    `json:"replay_verified"`
	RepositoryWrites   int                                     `json:"repository_writes"`
	ReceiptDigest      string                                  `json:"receipt_digest"`
}

type foundationPromotionJob struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"head_sha"`
	RunID      int64  `json:"run_id"`
	RunAttempt int64  `json:"run_attempt"`
}

type foundationPromotionEvidence struct {
	Schema          string                     `json:"schema"`
	Decision        string                     `json:"decision"`
	HumanDecision   string                     `json:"human_decision"`
	Repository      string                     `json:"repository"`
	PRNumber        int64                      `json:"pr_number"`
	BaseRef         string                     `json:"base_ref"`
	BaseSHA         string                     `json:"base_sha"`
	HeadRef         string                     `json:"head_ref"`
	HeadSHA         string                     `json:"head_sha"`
	ArtifactID      int64                      `json:"artifact_id"`
	ArtifactName    string                     `json:"artifact_name"`
	ArtifactSize    int64                      `json:"artifact_size_bytes"`
	ArtifactDigest  string                     `json:"artifact_digest"`
	ArtifactRunID   int64                      `json:"artifact_run_id"`
	ArtifactAttempt int64                      `json:"artifact_run_attempt"`
	Input           feedbackpredecessor.Input  `json:"input"`
	Receipt         foundationPromotionReceipt `json:"receipt"`
	NonRouteJobs    []foundationPromotionJob   `json:"non_route_jobs"`
}
