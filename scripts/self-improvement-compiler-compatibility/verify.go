package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compatibilitypolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compilercompatibility"
)

type verificationReport struct {
	Schema                string `json:"schema"`
	Decision              string `json:"decision"`
	Reason                string `json:"reason"`
	PolicySourceDigest    string `json:"policy_source_digest"`
	PolicyEvaluatorDigest string `json:"policy_evaluator_digest"`
	CertificateDigest     string `json:"certificate_digest"`
	EvidenceArtifactCount int    `json:"evidence_artifact_count"`
	ContinuityEdgeCount   int    `json:"continuity_edge_count"`
	CaseDenominator       int    `json:"case_denominator"`
	ClosedCases           int    `json:"closed_cases"`
	UnknownCases          int    `json:"unknown_cases"`
	RefutedCases          int    `json:"refuted_cases"`
	RepositoryWrites      int    `json:"repository_writes"`
	LocalTestExecutions   int    `json:"local_test_executions"`
}

func verifyConformance(contractPath, certificatePath, consumptionPath, publicationRoot, outputPath string) error {
	if contractPath == "" || certificatePath == "" || consumptionPath == "" || publicationRoot == "" || outputPath == "" {
		return errors.New("verify requires contract, certificate, consumption-report, publication-root, and output")
	}
	policy, err := loadPolicy(contractPath)
	if err != nil {
		return err
	}
	contractData, err := os.ReadFile(contractPath)
	if err != nil {
		return err
	}
	certificateData, certificate, err := readCertificate(certificatePath)
	if err != nil {
		return err
	}
	consumption, err := readConsumption(consumptionPath)
	if err != nil {
		return err
	}
	var report conformanceReport
	reportData, err := readStrict(outputPath, &report)
	if err != nil {
		return err
	}
	if err := verifyConformanceReport(policy, report); err != nil {
		return err
	}
	if consumption.Schema != compilercompatibility.ReportSchema ||
		consumption.Decision != compilercompatibility.DecisionClosed ||
		consumption.Reason != compilercompatibility.ReasonBoundedSuccessorReplay ||
		!consumption.ImplementationDigestDifferent ||
		!consumption.SemanticEqual ||
		!consumption.GeneratedBytesEqual ||
		!consumption.GeneratedManifestEqual ||
		!consumption.FullTestContractEqual ||
		consumption.UnsupportedFrontierDecision != compilercompatibility.UnsupportedFrontierDecision ||
		!sameStringSlice(consumption.UnsupportedFrontierClaims, compilercompatibility.UnsupportedFrontierClaims) ||
		consumption.CertificateCount != 1 ||
		consumption.CompatibilityHits != 1 ||
		consumption.CompatibilityMisses != 0 ||
		consumption.IdentityAxisCount != compatibilitypolicy.AxisCount ||
		consumption.CertificateBytes != len(certificateData) ||
		consumption.CertificateDigest != cache.HashBytes(certificateData).String() ||
		consumption.EvidenceArtifactCount != compatibilitypolicy.EvidenceArtifactCount ||
		consumption.ContinuityEdgeCount != compatibilitypolicy.ContinuityEdgeCount ||
		consumption.RepositoryWrites != 0 ||
		consumption.LocalTestExecutions != 0 {
		return errors.New("actual compatibility consumption is not the closed bounded replay")
	}
	entries, err := os.ReadDir(publicationRoot)
	if err != nil {
		return err
	}
	if len(entries) != len(publicationNames) {
		return fmt.Errorf("publication artifact count = %d, want %d", len(entries), len(publicationNames))
	}
	for _, name := range publicationNames {
		info, err := os.Stat(filepath.Join(publicationRoot, name))
		if err != nil {
			return fmt.Errorf("publication artifact %s: %w", name, err)
		}
		if info.IsDir() {
			return fmt.Errorf("publication artifact %s is a directory", name)
		}
	}
	publishedContract, err := os.ReadFile(filepath.Join(publicationRoot, "canonical-policy.gooo"))
	if err != nil {
		return err
	}
	if !bytes.Equal(publishedContract, contractData) {
		return errors.New("published canonical policy differs from the verified source")
	}
	publishedCertificate, err := os.ReadFile(filepath.Join(publicationRoot, "compatibility-certificate.json"))
	if err != nil {
		return err
	}
	if !bytes.Equal(publishedCertificate, certificateData) {
		return errors.New("published certificate bytes differ from the verified certificate")
	}
	publishedDigest, err := os.ReadFile(filepath.Join(publicationRoot, "compatibility-certificate-digest.txt"))
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(publishedDigest)) != cache.HashBytes(certificateData).String() {
		return errors.New("published certificate digest does not bind the certificate bytes")
	}
	publishedPolicy, err := os.ReadFile(filepath.Join(publicationRoot, "compatibility-policy.json"))
	if err != nil {
		return err
	}
	if !bytes.Contains(publishedPolicy, []byte(policy.SourceDigest)) || !bytes.Contains(publishedPolicy, []byte(policy.EvaluatorDigest)) {
		return errors.New("published policy does not bind source and evaluator identities")
	}
	var publishedAuthorization compilercompatibility.Authorization
	if _, err := readStrict(filepath.Join(publicationRoot, "compatibility-authorization.json"), &publishedAuthorization); err != nil {
		return err
	}
	if err := compilercompatibility.ValidateAuthorization(publishedAuthorization); err != nil {
		return err
	}
	if publishedAuthorization.AuthorizationID != certificate.Authorization.AuthorizationID {
		return errors.New("published authorization differs from the certificate authorization")
	}
	if len(reportData) == 0 {
		return errors.New("conformance report is empty")
	}
	return writeJSON(filepath.Join(filepath.Dir(outputPath), "verification-report.json"), verificationReport{
		Schema:                "gooo/compiler-successor-compatibility-verification/v1",
		Decision:              compilercompatibility.DecisionClosed,
		Reason:                "VERIFIED_BOUNDED_IMPLEMENTATION_SUCCESSOR_CONFORMANCE",
		PolicySourceDigest:    policy.SourceDigest,
		PolicyEvaluatorDigest: policy.EvaluatorDigest,
		CertificateDigest:     cache.HashBytes(certificateData).String(),
		EvidenceArtifactCount: len(publicationNames),
		ContinuityEdgeCount:   policy.ContinuityEdges,
		CaseDenominator:       report.CaseDenominator,
		ClosedCases:           report.ClosedCases,
		UnknownCases:          report.UnknownCases,
		RefutedCases:          report.RefutedCases,
		RepositoryWrites:      report.RepositoryWrites,
		LocalTestExecutions:   report.LocalTestExecutions,
	})
}

func verifyConformanceReport(policy compatibilitypolicy.Policy, report conformanceReport) error {
	if report.Schema != "gooo/compiler-successor-compatibility-conformance/v1" ||
		report.Decision != compilercompatibility.DecisionClosed ||
		report.Reason != compilercompatibility.ReasonBoundedSuccessorReplay ||
		report.PolicySourceDigest != policy.SourceDigest ||
		report.PolicyEvaluatorDigest != policy.EvaluatorDigest ||
		report.CaseDenominator != 6 ||
		report.IdentityAxisCount != compatibilitypolicy.AxisCount ||
		report.ClosedCases != 2 || report.UnknownCases != 2 || report.RefutedCases != 2 ||
		report.ImplementationDigestEqual ||
		!report.ImplementationDigestDifferent ||
		!report.SemanticEqual ||
		!report.GeneratedBytesEqual ||
		!report.GeneratedManifestEqual ||
		!report.PolicyResultEqual ||
		!report.FullTestContractEqual ||
		report.IndependentReplayExecutions != 2 ||
		report.TestContractReplays != 2 ||
		report.CompatibilityScopeSubjects != 1 ||
		report.CertificateCount != 1 ||
		report.CertificateBytes <= 0 ||
		report.CompatibilityHits != 1 ||
		report.CompatibilityMisses != 0 ||
		report.MismatchDetections != 6 ||
		report.CertificateTamperDetections != 1 ||
		report.ScopeWideningDetections != 1 ||
		report.FallbackAttempts != 1 ||
		report.FallbackAccepted != 0 ||
		report.FallbackRejected != 1 ||
		report.EvidenceArtifactCount != compatibilitypolicy.EvidenceArtifactCount ||
		report.ContinuityEdgeCount != compatibilitypolicy.ContinuityEdgeCount ||
		report.Claim != "BOUNDED_IMPLEMENTATION_REUSE_ELIGIBILITY" ||
		report.UnsupportedFrontierDecision != compilercompatibility.UnsupportedFrontierDecision ||
		!sameStringSlice(report.UnsupportedFrontierClaims, compilercompatibility.UnsupportedFrontierClaims) ||
		report.PerformanceClaim ||
		report.GeneralCompatibilityClaim ||
		report.RepositoryWrites != 0 ||
		report.LocalTestExecutions != 0 ||
		report.WallMS <= 0 ||
		report.PeakRSSKib <= 0 ||
		!sameStringSlice(report.CaseIDs, compatibilitypolicy.CaseIDs()) ||
		!sameStringSlice(report.EvidenceArtifactNames, publicationNames) ||
		report.StrictPredecessorConsumption.Decision != compilercompatibility.DecisionRefuted ||
		report.StrictPredecessorConsumption.State != "STRICT_IMPLEMENTATION_MISMATCH" {
		return errors.New("conformance report does not satisfy the fixed v18 contract")
	}
	if len(report.Cases) != 6 {
		return errors.New("conformance report case count is not six")
	}
	for index, item := range report.Cases {
		expected, ok := policy.Decision(item.ID)
		if !ok || item.ID != report.CaseIDs[index] || item.ExpectedDecision != expected || item.ObservedDecision != expected {
			return fmt.Errorf("conformance case %q is not source-bound", item.ID)
		}
		if item.ObservedDecision == compilercompatibility.DecisionUnknown && !completeUnknown(item.Unknown) {
			return fmt.Errorf("conformance case %q has incomplete UNKNOWN causality", item.ID)
		}
		if item.ObservedDecision == compilercompatibility.DecisionRefuted && item.Unknown != nil {
			return fmt.Errorf("conformance case %q has UNKNOWN state on REFUTED", item.ID)
		}
	}
	return nil
}

func sameStringSlice(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
