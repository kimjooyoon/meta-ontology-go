package generation

import (
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const (
	SemanticPublicGenerationBaselineLifecycle = "PUBLIC_GENERATE_BASELINE"
	SemanticPublicGenerationRetainedLifecycle = "PUBLIC_GENERATE_RETAINED"
	SemanticPublicGenerationFailClosed        = "PUBLIC_GENERATE_FAIL_CLOSED"
)

// SemanticPublicGenerationReport is the caller-owned explanation emitted by
// the ordinary compiler path when reporting baseline or retained generation.
type SemanticPublicGenerationReport struct {
	Schema                  string                          `json:"schema"`
	Lifecycle               string                          `json:"lifecycle"`
	Decision                string                          `json:"decision"`
	Reason                  string                          `json:"reason"`
	Unknown                 *EnvelopeUnknownState           `json:"unknown"`
	CertificateDigest       string                          `json:"certificate_digest"`
	AdoptionReportDigest    string                          `json:"adoption_report_digest"`
	ObservationDigest       string                          `json:"observation_digest"`
	ProposalDigest          string                          `json:"proposal_digest"`
	AuthorizationDigest     string                          `json:"authorization_digest"`
	CandidateStableID       string                          `json:"candidate_stable_id"`
	ContractSourceDigest    string                          `json:"contract_source_digest"`
	InputSourceDigest       string                          `json:"input_source_digest"`
	NormalizedIRDigest      string                          `json:"normalized_ir_digest"`
	GeneratedOutputDigest   string                          `json:"generated_output_digest"`
	GeneratedManifestDigest string                          `json:"generated_manifest_digest"`
	CompilerDigest          string                          `json:"compiler_digest"`
	ToolchainDigest         string                          `json:"toolchain_digest"`
	VerifierDigest          string                          `json:"verifier_digest"`
	PolicyDigest            string                          `json:"policy_digest"`
	EvaluatorDigest         string                          `json:"evaluator_digest"`
	OutputFile              string                          `json:"output_file"`
	ManifestFile            string                          `json:"manifest_file"`
	ArtifactCount           int                             `json:"artifact_count"`
	OutputFileCount         int                             `json:"output_file_count"`
	OutputBytes             int64                           `json:"output_bytes"`
	Metrics                 SemanticRetentionRuntimeMetrics `json:"metrics"`
	RepositoryWrites        int                             `json:"repository_writes"`
	LocalTestExecutions     int                             `json:"local_test_executions"`
}

// ValidateSemanticPublicGenerationReport keeps fail-closed reports distinct
// from an applied retained result and fixes the caller artifact denominator.
func ValidateSemanticPublicGenerationReport(report SemanticPublicGenerationReport) error {
	if report.Schema != SemanticPublicGenerationReportSchema || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || report.ArtifactCount < 2 || report.OutputFileCount != report.ArtifactCount || report.OutputBytes <= 0 || !cache.Digest(report.EvaluatorDigest).Known() {
		return errors.New("semantic public generation report is invalid")
	}
	if report.Decision == "CLOSED" {
		return validatePublicGenerationClosed(report)
	}
	if report.Decision == "UNKNOWN" {
		return validatePublicGenerationUnknown(report)
	}
	if report.Decision == SemanticRetentionRefuted {
		if report.Lifecycle != SemanticPublicGenerationFailClosed || report.Reason != SemanticRetentionRefutedReason || report.Unknown != nil || report.ArtifactCount != SemanticPublicGenerationArtifactsFailClosed || report.Metrics.CertificateMisses != 1 {
			return errors.New("semantic public generation REFUTED report is invalid")
		}
		return nil
	}
	return errors.New("semantic public generation decision is invalid")
}

func validatePublicGenerationClosed(report SemanticPublicGenerationReport) error {
	if report.ArtifactCount != SemanticPublicGenerationArtifactsClosed || report.OutputFile == "" || report.ManifestFile == "" || !cache.Digest(report.InputSourceDigest).Known() || !cache.Digest(report.NormalizedIRDigest).Known() || !cache.Digest(report.GeneratedOutputDigest).Known() || !cache.Digest(report.GeneratedManifestDigest).Known() || report.Unknown != nil || !report.Metrics.GeneratedBytesEqual || !report.Metrics.NormalizedSemanticEqual {
		return errors.New("semantic public generation CLOSED report is invalid")
	}
	if report.Lifecycle == SemanticPublicGenerationBaselineLifecycle {
		if report.Reason != SemanticPublicGenerationBaselineReason || report.CertificateDigest != "" || report.Metrics.SemanticOperationCount != 1 || report.Metrics.CertificateHits != 0 || report.Metrics.CertificateMisses != 0 {
			return errors.New("semantic public generation baseline report is invalid")
		}
		return nil
	}
	if report.Lifecycle != SemanticPublicGenerationRetainedLifecycle || report.Reason != SemanticPublicGenerationHitReason || !cache.Digest(report.CertificateDigest).Known() || report.Metrics.SemanticOperationCount != 0 || report.Metrics.CertificateHits != 1 || report.Metrics.CertificateMisses != 0 {
		return errors.New("semantic public generation retained report is invalid")
	}
	return validatePublicGenerationBindings(report)
}

func validatePublicGenerationUnknown(report SemanticPublicGenerationReport) error {
	if report.Lifecycle != SemanticPublicGenerationFailClosed || report.ArtifactCount != SemanticPublicGenerationArtifactsFailClosed || report.CertificateDigest != "" || report.Unknown == nil || report.Metrics != (SemanticRetentionRuntimeMetrics{}) {
		return errors.New("semantic public generation UNKNOWN report is invalid")
	}
	if report.Reason != SemanticRetentionUnknownCertificateReason && report.Reason != SemanticRetentionUnknownAuthorizationReason {
		return errors.New("semantic public generation UNKNOWN reason is invalid")
	}
	return nil
}

func validatePublicGenerationBindings(report SemanticPublicGenerationReport) error {
	for _, digest := range []string{report.AdoptionReportDigest, report.ObservationDigest, report.ProposalDigest, report.AuthorizationDigest, report.ContractSourceDigest, report.InputSourceDigest, report.NormalizedIRDigest, report.GeneratedOutputDigest, report.GeneratedManifestDigest, report.CompilerDigest, report.ToolchainDigest, report.VerifierDigest, report.PolicyDigest, report.EvaluatorDigest} {
		if !cache.Digest(digest).Known() {
			return errors.New("semantic public generation report contains unknown binding")
		}
	}
	if report.CandidateStableID == "" {
		return errors.New("semantic public generation report has no candidate binding")
	}
	return nil
}
