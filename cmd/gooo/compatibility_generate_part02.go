package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compatibilitypolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compilercompatibility"
)

func compatibilityConsumptionReport(policy compatibilitypolicy.Policy, certificate compilercompatibility.Certificate, certificateDigest string, current compilercompatibility.ExecutionReceipt, evaluation compilercompatibility.Evaluation, output, manifest string, started time.Time) compilercompatibility.ConsumptionReport {
	_ = current
	comparisons := compilercompatibility.CompareAxes(certificate.PredecessorAxes, certificate.SuccessorAxes)
	return compilercompatibility.ConsumptionReport{
		Schema: compilercompatibility.ReportSchema, Lifecycle: "COMPATIBILITY_CONSUMPTION",
		Decision: evaluation.Decision, Reason: evaluation.Reason, CertificateDigest: certificateDigest,
		CandidateStableID: certificate.CandidateStableID, SubjectDigest: certificate.SubjectDigest, SourceDigest: certificate.SourceDigest,
		StrictPredecessorConsumption: certificate.StrictPredecessorConsumption, IdentityAxisCount: compatibilitypolicy.AxisCount,
		AxisComparisons: comparisons, ImplementationDigestEqual: comparisons[1].Equal, ImplementationDigestDifferent: !comparisons[1].Equal,
		SemanticEqual: certificate.NormalizedSemanticEqual, GeneratedBytesEqual: certificate.GeneratedBytesEqual,
		GeneratedManifestEqual: certificate.GeneratedManifestEqual, PolicyResultEqual: certificate.PolicyResultEqual,
		FullTestContractEqual:       certificate.FullTestContractEqual,
		IndependentReplayExecutions: certificate.IndependentReplayExecutions, TestContractReplays: 2,
		CompatibilityScopeSubjects: len(certificate.Scope), CertificateCount: 1, CertificateBytes: certificateBytes(certificate),
		CompatibilityHits: 1, CompatibilityMisses: 0, EvidenceArtifactCount: policy.EvidenceArtifacts,
		ContinuityEdgeCount: policy.ContinuityEdges, Claim: "BOUNDED_IMPLEMENTATION_REUSE_ELIGIBILITY",
		PerformanceClaim: false, GeneralCompatibilityClaim: false, OutputFile: output, ManifestFile: manifest,
		UnsupportedFrontierDecision: compilercompatibility.UnsupportedFrontierDecision, UnsupportedFrontierClaims: append([]string(nil), compilercompatibility.UnsupportedFrontierClaims...),
		ArtifactCount: 4, OutputBytes: int64(len(certificate.GeneratedSource) + len(certificate.GeneratedManifest)),
		RepositoryWrites: 0, LocalTestExecutions: 0, WallMS: compatibilityWallMS(started), PeakRSSKib: readPeakRSSKib(),
	}
}

func certificateBytes(certificate compilercompatibility.Certificate) int {
	data, err := json.MarshalIndent(certificate, "", "  ")
	if err != nil {
		return 0
	}
	return len(append(data, '\n'))
}

func compatibilityWallMS(started time.Time) int64 {
	return int64((time.Since(started).Nanoseconds() + int64(time.Millisecond) - 1) / int64(time.Millisecond))
}

func renderCompatibilityReport(report compilercompatibility.ConsumptionReport) string {
	return fmt.Sprintf("# Gooo compiler compatibility consumption\n\nDecision: %s\nReason: %s\nIdentity axes: %d\nImplementation digest equal/different: %t/%t\nSemantic/output/manifest/test equality: %t/%t/%t/%t\nIndependent replay executions: %d\nCompatibility scope subjects: %d\nCompatibility hit/miss: %d/%d\nEvidence artifacts: %d\nRepository writes: %d\nLocal test executions: %d\nWall time (ms): %d\nPeak RSS (KiB): %d\n\nThe only claim is bounded implementation-successor reuse eligibility; no performance or general compiler-compatibility claim is made.\n", report.Decision, report.Reason, report.IdentityAxisCount, report.ImplementationDigestEqual, report.ImplementationDigestDifferent, report.SemanticEqual, report.GeneratedBytesEqual, report.GeneratedManifestEqual, report.FullTestContractEqual, report.IndependentReplayExecutions, report.CompatibilityScopeSubjects, report.CompatibilityHits, report.CompatibilityMisses, report.EvidenceArtifactCount, report.RepositoryWrites, report.LocalTestExecutions, report.WallMS, report.PeakRSSKib)
}
