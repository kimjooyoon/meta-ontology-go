package publiccontinuity

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicdiscovery"
)

const (
	DecisionReceiptSchema = "gooo/public-self-observation-decision/v1"
	CertificateSchema     = "gooo/public-self-observation-continuity-certificate/v1"
	ReportSchema          = "gooo/public-self-observation-continuity-report/v1"

	ConversionSchema = "gooo/public-discovery-candidate-to-continuity-certificate/v1"
	CertificateMode  = "caller-owned-immutable-digest-bound"

	DecisionAccept = "ACCEPT"
	DecisionReject = "REJECT"
	ReasonAccepted = "EXPLICIT_HUMAN_ACCEPT"
	ReasonRejected = "EXPLICIT_HUMAN_REJECT"

	Operation = publicdiscovery.Operation
)

var compilerManifestPaths = append(generation.SemanticRetentionCompilerManifestPaths(),
	"cmd/gooo/public_continuity_authorize_part01.go", "cmd/gooo/public_continuity_certify_part01.go",
	"cmd/gooo/public_continuity_generate_part01.go", "cmd/gooo/emit_dispatch.go", "internal/meta/publiccontinuity/continuity.go",
	"internal/meta/publicdiscovery/discovery.go", "internal/meta/discoverypolicy/policy.go",
	"internal/meta/discoverypolicy/generated/evaluator.go")

var verifierManifestPaths = append(generation.SemanticRetentionVerifierManifestPaths(),
	"scripts/self-improvement-public-continuity/main.go", "internal/meta/publiccontinuity/continuity.go")

func CompilerDigest(readFile func(string) ([]byte, error)) (string, error) {
	return generation.SemanticRetentionManifestDigest(readFile, compilerManifestPaths)
}

func VerifierDigest(readFile func(string) ([]byte, error)) (string, error) {
	return generation.SemanticRetentionManifestDigest(readFile, verifierManifestPaths)
}

// Binding is copied into each handoff artifact so an independent consumer can
// compare the full public-discovery identity without relying on producer-side
// counts or a path convention.
type Binding struct {
	CandidateDigest         string `json:"candidate_digest"`
	CandidateID             string `json:"candidate_id"`
	GroupKeyDigest          string `json:"group_key_digest"`
	SourceDigest            string `json:"source_digest"`
	InputSemanticDigest     string `json:"input_semantic_digest"`
	PreviousGoDigest        string `json:"previous_go_digest"`
	ToolchainDigest         string `json:"toolchain_digest"`
	ContractDigest          string `json:"contract_digest"`
	EvaluatorDigest         string `json:"evaluator_digest"`
	GeneratedSemanticDigest string `json:"generated_semantic_digest"`
	GeneratedOutputDigest   string `json:"generated_output_digest"`
	GeneratedManifestDigest string `json:"generated_manifest_digest"`
}

type DecisionReceipt struct {
	Schema                string                    `json:"schema"`
	ReceiptID             string                    `json:"receipt_id"`
	Operation             string                    `json:"operation"`
	Decision              string                    `json:"decision"`
	Reason                string                    `json:"reason"`
	ExplicitHumanDecision bool                      `json:"explicit_human_decision"`
	ExecutionAllowed      bool                      `json:"execution_allowed"`
	ManualTransformations int                       `json:"manual_transformations"`
	Binding               Binding                   `json:"binding"`
	Candidate             publicdiscovery.Candidate `json:"candidate"`
	RepositoryWrites      int                       `json:"repository_writes"`
	LocalBuildExecutions  int                       `json:"local_build_executions"`
	LocalTestExecutions   int                       `json:"local_test_executions"`
}

type Certificate struct {
	Schema                  string  `json:"schema"`
	CertificateID           string  `json:"certificate_id"`
	Mode                    string  `json:"mode"`
	ConversionSchema        string  `json:"conversion_schema"`
	SourceOperation         string  `json:"source_operation"`
	TargetOperation         string  `json:"target_operation"`
	DecisionReceiptDigest   string  `json:"decision_receipt_digest"`
	Binding                 Binding `json:"binding"`
	ContractSourceDigest    string  `json:"contract_source_digest"`
	InputSourceDigest       string  `json:"input_source_digest"`
	CompilerDigest          string  `json:"compiler_digest"`
	VerifierDigest          string  `json:"verifier_digest"`
	PolicyDigest            string  `json:"policy_digest"`
	EvaluatorDigest         string  `json:"evaluator_digest"`
	GeneratedSource         []byte  `json:"generated_source"`
	GeneratedManifest       []byte  `json:"generated_manifest"`
	GeneratedManifestDigest string  `json:"generated_manifest_digest"`
	ManualTransformations   int     `json:"manual_transformations"`
	RepositoryWrites        int     `json:"repository_writes"`
	LocalBuildExecutions    int     `json:"local_build_executions"`
	LocalTestExecutions     int     `json:"local_test_executions"`
}

type Report struct {
	Schema                                   string                           `json:"schema"`
	Lifecycle                                string                           `json:"lifecycle"`
	Decision                                 string                           `json:"decision"`
	Reason                                   string                           `json:"reason"`
	CaseID                                   string                           `json:"case_id"`
	Unknown                                  *generation.EnvelopeUnknownState `json:"unknown"`
	Binding                                  Binding                          `json:"binding"`
	DecisionReceiptDigest                    string                           `json:"decision_receipt_digest"`
	CertificateDigest                        string                           `json:"certificate_digest"`
	PublicInvocations                        int                              `json:"public_invocations"`
	LedgerEntries                            int                              `json:"ledger_entries"`
	Candidates                               int                              `json:"candidates"`
	DecisionReceipts                         int                              `json:"decision_receipts"`
	AcceptedDecisions                        int                              `json:"accepted_decisions"`
	RejectedDecisions                        int                              `json:"rejected_decisions"`
	MissingDecisions                         int                              `json:"missing_decisions"`
	Certificates                             int                              `json:"certificates"`
	DigestContinuityEdgesExpected            int                              `json:"digest_continuity_edges_expected"`
	DigestContinuityEdgesObserved            int                              `json:"digest_continuity_edges_observed"`
	ManualTransformations                    int                              `json:"manual_transformations"`
	SemanticOperationsBefore                 int                              `json:"semantic_operations_before"`
	SemanticOperationsAfter                  int                              `json:"semantic_operations_after"`
	CandidateCertificateByteReplayMismatches int                              `json:"candidate_certificate_byte_replay_mismatches"`
	GeneratedBytesEqual                      bool                             `json:"generated_bytes_equal"`
	NormalizedSemanticEqual                  bool                             `json:"normalized_semantic_equal"`
	ArtifactDenominator                      int                              `json:"artifact_denominator"`
	ArtifactCount                            int                              `json:"artifact_count"`
	RepositoryWrites                         int                              `json:"repository_writes"`
	LocalBuildExecutions                     int                              `json:"local_build_executions"`
	LocalTestExecutions                      int                              `json:"local_test_executions"`
	WallMS                                   int64                            `json:"wall_ms"`
	PeakRSSKib                               int64                            `json:"peak_rss_kib"`
}

func BindingFromCandidate(candidate publicdiscovery.Candidate, candidateDigest string) Binding {
	return Binding{CandidateDigest: candidateDigest, CandidateID: candidate.CandidateID,
		GroupKeyDigest: candidate.GroupKeyDigest,
		SourceDigest:   candidate.SourceDigest, InputSemanticDigest: candidate.InputSemanticDigest,
		PreviousGoDigest: candidate.PreviousGoDigest, ToolchainDigest: candidate.ToolchainDigest,
		ContractDigest: candidate.ContractDigest, EvaluatorDigest: candidate.EvaluatorDigest,
		GeneratedSemanticDigest: candidate.GeneratedSemanticDigest, GeneratedOutputDigest: candidate.GeneratedOutputDigest,
		GeneratedManifestDigest: candidate.GeneratedManifestDigest}
}

func DecodeCandidate(raw []byte) (publicdiscovery.Candidate, error) {
	var candidate publicdiscovery.Candidate
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&candidate); err != nil {
		return candidate, fmt.Errorf("decode discovery candidate: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return candidate, errors.New("discovery candidate contains multiple JSON values")
	} else if err != io.EOF {
		return candidate, fmt.Errorf("decode discovery candidate trailer: %w", err)
	}
	if err := ValidateCandidate(candidate); err != nil {
		return candidate, err
	}
	return candidate, nil
}

func ValidateCandidate(candidate publicdiscovery.Candidate) error {
	if candidate.Schema != publicdiscovery.CandidateSchema || candidate.Operation != Operation ||
		candidate.Decision != "CLOSED" || candidate.Reason == "" || candidate.CandidateID == "" || candidate.Quorum != publicdiscovery.Quorum ||
		candidate.ExecutionAllowed || !candidate.AuthorizationRequired || !candidate.ProposalRequired || !candidate.CertificateRequired ||
		candidate.RepositoryWrites != 0 || candidate.LocalBuildExecutions != 0 || candidate.LocalTestExecutions != 0 {
		return errors.New("discovery candidate is not an exact non-executing handoff candidate")
	}
	for _, value := range []string{candidate.GroupKeyDigest, candidate.SourceDigest, candidate.InputSemanticDigest,
		candidate.PreviousGoDigest, candidate.ToolchainDigest, candidate.ContractDigest, candidate.EvaluatorDigest,
		candidate.GeneratedSemanticDigest, candidate.GeneratedOutputDigest, candidate.GeneratedManifestDigest} {
		if !cache.Digest(value).Known() {
			return errors.New("discovery candidate contains an unknown digest")
		}
	}
	return nil
}

func ValidateBinding(binding Binding, candidate publicdiscovery.Candidate, candidateDigest string) error {
	want := BindingFromCandidate(candidate, candidateDigest)
	if binding != want {
		return errors.New("handoff binding does not match the exact discovery candidate")
	}
	return nil
}

func DecisionReceiptContentDigest(receipt DecisionReceipt) (string, error) {
	receipt.ReceiptID = ""
	digest, err := cache.DigestOf(receipt)
	if err != nil {
		return "", fmt.Errorf("decision receipt content digest: %w", err)
	}
	return digest.String(), nil
}

func ValidateDecisionReceipt(receipt DecisionReceipt) error {
	if receipt.Schema != DecisionReceiptSchema || receipt.Operation != Operation || receipt.ReceiptID == "" ||
		!cache.Digest(receipt.ReceiptID).Known() || !receipt.ExplicitHumanDecision || receipt.ExecutionAllowed ||
		receipt.ManualTransformations != 0 || receipt.RepositoryWrites != 0 || receipt.LocalBuildExecutions != 0 || receipt.LocalTestExecutions != 0 {
		return errors.New("decision receipt is invalid or not explicitly human-controlled")
	}
	if receipt.Decision != DecisionAccept && receipt.Decision != DecisionReject {
		return errors.New("decision receipt has an unknown decision")
	}
	wantReason := ReasonAccepted
	if receipt.Decision == DecisionReject {
		wantReason = ReasonRejected
	}
	if receipt.Reason != wantReason {
		return errors.New("decision receipt reason does not describe the explicit decision")
	}
	if err := ValidateCandidate(receipt.Candidate); err != nil {
		return err
	}
	if !cache.Digest(receipt.Binding.CandidateDigest).Known() || !cache.Digest(receipt.Binding.GroupKeyDigest).Known() || receipt.Binding.CandidateID != receipt.Candidate.CandidateID {
		return errors.New("decision receipt candidate binding is invalid")
	}
	if err := ValidateBinding(receipt.Binding, receipt.Candidate, receipt.Binding.CandidateDigest); err != nil {
		return err
	}
	digest, err := DecisionReceiptContentDigest(receipt)
	if err != nil || digest != receipt.ReceiptID {
		return errors.New("decision receipt is not content-addressed")
	}
	return nil
}

func CertificateContentDigest(certificate Certificate) (string, error) {
	certificate.CertificateID = ""
	digest, err := cache.DigestOf(certificate)
	if err != nil {
		return "", fmt.Errorf("continuity certificate content digest: %w", err)
	}
	return digest.String(), nil
}

func ValidateCertificate(certificate Certificate) error {
	if certificate.Schema != CertificateSchema || certificate.CertificateID == "" || !cache.Digest(certificate.CertificateID).Known() ||
		certificate.Mode != CertificateMode || certificate.ConversionSchema != ConversionSchema || certificate.SourceOperation != Operation ||
		certificate.TargetOperation != "gooo.generate.public-self-observation-consumption" || !cache.Digest(certificate.DecisionReceiptDigest).Known() ||
		certificate.ManualTransformations != 0 || certificate.RepositoryWrites != 0 || certificate.LocalBuildExecutions != 0 || certificate.LocalTestExecutions != 0 ||
		certificate.Binding.CandidateID == "" ||
		len(certificate.GeneratedSource) == 0 || len(certificate.GeneratedManifest) == 0 {
		return errors.New("continuity certificate is invalid")
	}
	for _, value := range []string{certificate.ContractSourceDigest, certificate.InputSourceDigest, certificate.CompilerDigest,
		certificate.VerifierDigest, certificate.PolicyDigest, certificate.EvaluatorDigest, certificate.Binding.CandidateDigest, certificate.Binding.GroupKeyDigest,
		certificate.Binding.SourceDigest, certificate.Binding.InputSemanticDigest, certificate.Binding.PreviousGoDigest,
		certificate.Binding.ToolchainDigest, certificate.Binding.ContractDigest, certificate.Binding.EvaluatorDigest,
		certificate.Binding.GeneratedSemanticDigest, certificate.Binding.GeneratedOutputDigest, certificate.Binding.GeneratedManifestDigest,
		certificate.GeneratedManifestDigest} {
		if !cache.Digest(value).Known() {
			return errors.New("continuity certificate contains an unknown digest")
		}
	}
	if certificate.ContractSourceDigest != certificate.Binding.ContractDigest || certificate.PolicyDigest != certificate.ContractSourceDigest ||
		certificate.InputSourceDigest != certificate.Binding.SourceDigest || certificate.EvaluatorDigest != certificate.Binding.EvaluatorDigest ||
		cache.HashBytes(certificate.GeneratedSource).String() != certificate.Binding.GeneratedOutputDigest ||
		cache.HashBytes(certificate.GeneratedManifest).String() != certificate.GeneratedManifestDigest ||
		certificate.Binding.GeneratedManifestDigest != certificate.GeneratedManifestDigest {
		return errors.New("continuity certificate bytes and bindings disagree")
	}
	digest, err := CertificateContentDigest(certificate)
	if err != nil || digest != certificate.CertificateID {
		return errors.New("continuity certificate is not content-addressed")
	}
	return nil
}
