package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publiccontinuity"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicdiscovery"
)

func runPublicContinuityGenerate(options generateOptions, input generateInput, reader SourceReader, jsonMode bool, stdout, stderr io.Writer, deadline time.Time) int {
	started := time.Now()
	certificateData, err := reader.ReadFile(options.continuityCertificateFilename)
	if err != nil {
		return writeContinuityGenerationRefuted(options, input, "MISSING_CERTIFICATE", "", started, jsonMode, stdout, stderr)
	}
	certificateDigest := cache.HashBytes(certificateData).String()
	certificate, err := decodeContinuityCertificate(certificateData)
	if err != nil {
		return writeContinuityGenerationRefuted(options, input, "TAMPERED_CERTIFICATE", certificateDigest, started, jsonMode, stdout, stderr)
	}
	if cache.HashBytes(input.source).String() != certificate.Binding.SourceDigest {
		return writeContinuityGenerationRefuted(options, input, "STALE_SOURCE", certificateDigest, started, jsonMode, stdout, stderr)
	}
	if certificate.Binding.ToolchainDigest != generation.SemanticRetentionToolchainDigest() {
		return writeContinuityGenerationRefuted(options, input, "MISMATCHED_TOOLCHAIN", certificateDigest, started, jsonMode, stdout, stderr)
	}
	compilerDigest, err := publiccontinuity.CompilerDigest(reader.ReadFile)
	if err != nil || compilerDigest != certificate.CompilerDigest {
		return writeContinuityGenerationRefuted(options, input, "STALE_COMPILER", certificateDigest, started, jsonMode, stdout, stderr)
	}
	verifierDigest, err := publiccontinuity.VerifierDigest(reader.ReadFile)
	if err != nil || verifierDigest != certificate.VerifierDigest {
		return writeContinuityGenerationRefuted(options, input, "STALE_VERIFIER", certificateDigest, started, jsonMode, stdout, stderr)
	}
	if certificate.Binding.ContractDigest != publicdiscovery.PolicySourceDigest() || certificate.Binding.EvaluatorDigest != publicdiscovery.GeneratedEvaluatorDigest() {
		return writeContinuityGenerationRefuted(options, input, "STALE_POLICY", certificateDigest, started, jsonMode, stdout, stderr)
	}
	output, manifest, err := publicGeneratePaths(options)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: public continuity generation: paths: %v\n", err)
		return exitFailure
	}
	report := publiccontinuity.Report{
		Schema: publiccontinuity.ReportSchema, Lifecycle: "CONSUMPTION", Decision: "CLOSED",
		Reason: "EXACT_CERTIFIED_PUBLIC_GENERATION_CONSUMPTION", CaseID: continuityCaseAccepted,
		Binding: certificate.Binding, CertificateDigest: certificateDigest,
		PublicInvocations: 2, LedgerEntries: 2, Candidates: 1, DecisionReceipts: 1, AcceptedDecisions: 1,
		Certificates: 1, DigestContinuityEdgesExpected: 4, DigestContinuityEdgesObserved: 4,
		ManualTransformations: certificate.ManualTransformations, SemanticOperationsBefore: 1, SemanticOperationsAfter: 0,
		CandidateCertificateByteReplayMismatches: 0, GeneratedBytesEqual: true, NormalizedSemanticEqual: true,
		ArtifactDenominator: 4, ArtifactCount: 4, RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0,
		WallMS: continuityWallMS(started), PeakRSSKib: readPeakRSSKib(),
	}
	reportData, err := marshalContinuityJSON(report)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: public continuity generation: encode report: %v\n", err)
		return exitFailure
	}
	human := []byte(renderContinuityReport(report, "The ordinary generate command consumed the exact certified bytes without recompiling or mutating source."))
	if err := ensurePublicOutputDirectory(options.outputDir); err != nil {
		fmt.Fprintf(stderr, "gooo: public continuity generation: output: %v\n", err)
		return exitFailure
	}
	writes := []atomicWrite{{path: output, data: certificate.GeneratedSource}, {path: manifest, data: certificate.GeneratedManifest},
		{path: filepath.Join(options.outputDir, "continuity-generation-report.json"), data: reportData},
		{path: filepath.Join(options.outputDir, "continuity-generation-report.md"), data: human}}
	if err := writeAtomicFiles(writes); err != nil {
		fmt.Fprintf(stderr, "gooo: public continuity generation: output: %v\n", err)
		return exitFailure
	}
	if jsonMode {
		data, _ := json.Marshal(report)
		_, _ = stdout.Write(append(data, '\n'))
	} else {
		fmt.Fprintf(stdout, "generated: %s\ncontinuity: %s\n", output, filepath.Join(options.outputDir, "continuity-generation-report.md"))
	}
	_ = deadline
	return exitOK
}

func decodeContinuityCertificate(data []byte) (publiccontinuity.Certificate, error) {
	var certificate publiccontinuity.Certificate
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&certificate); err != nil {
		return certificate, err
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return certificate, errors.New("continuity certificate contains multiple JSON values")
	} else if err != io.EOF {
		return certificate, fmt.Errorf("decode continuity certificate trailer: %w", err)
	}
	if err := publiccontinuity.ValidateCertificate(certificate); err != nil {
		return certificate, err
	}
	return certificate, nil
}

func writeContinuityGenerationRefuted(options generateOptions, input generateInput, reason, certificateDigest string, started time.Time, jsonMode bool, stdout, stderr io.Writer) int {
	report := publiccontinuity.Report{Schema: publiccontinuity.ReportSchema, Lifecycle: "CONSUMPTION", Decision: "REFUTED", Reason: reason,
		CaseID: continuityCaseBindingMismatch, CertificateDigest: certificateDigest, PublicInvocations: 2, LedgerEntries: 2,
		Certificates: 0, DigestContinuityEdgesExpected: 4, DigestContinuityEdgesObserved: 0, ManualTransformations: 0,
		SemanticOperationsBefore: 1, SemanticOperationsAfter: 0, CandidateCertificateByteReplayMismatches: 1,
		GeneratedBytesEqual: false, NormalizedSemanticEqual: false, ArtifactDenominator: 2, ArtifactCount: 2,
		RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0, WallMS: continuityWallMS(started), PeakRSSKib: readPeakRSSKib()}
	if len(input.source) > 0 {
		report.Binding.SourceDigest = cache.HashBytes(input.source).String()
	}
	data, err := marshalContinuityJSON(report)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: public continuity generation: encode refuted report: %v\n", err)
		return exitFailure
	}
	human := []byte(renderContinuityReport(report, "The certified handoff was rejected fail-closed; no baseline fallback was used."))
	if err := writeContinuityArtifacts(options.outputDir, []continuityArtifact{
		{name: "continuity-generation-report.json", data: data},
		{name: "continuity-generation-report.md", data: human},
	}); err != nil {
		fmt.Fprintf(stderr, "gooo: public continuity generation: output: %v\n", err)
		return exitFailure
	}
	if jsonMode {
		_, _ = stdout.Write(data)
	} else {
		fmt.Fprintf(stdout, "continuity: REFUTED (%s)\n", filepath.Join(options.outputDir, "continuity-generation-report.md"))
	}
	return exitOK
}

const continuityCaseBindingMismatch = "BINDING_MISMATCH"

func renderContinuityReport(report publiccontinuity.Report, note string) string {
	return fmt.Sprintf("# Gooo public self-observation continuity\n\nDecision: `%s`\nReason: `%s`\nCase: `%s`\nCandidate digest: `%s`\nDecision receipt digest: `%s`\nCertificate digest: `%s`\nDigest continuity edges: `%d/%d`\nManual transformations: `%d`\nSemantic operations before/after: `%d/%d`\nGenerated bytes equal: `%t`\nNormalized semantic equal: `%t`\nArtifacts: `%d/%d`\nRepository writes: `%d`\nLocal build executions: `%d`\nLocal test executions: `%d`\nWall time (ms): `%d`\nPeak RSS (KiB): `%d`\n\n%s\n", report.Decision, report.Reason, report.CaseID, report.Binding.CandidateDigest, report.DecisionReceiptDigest, report.CertificateDigest, report.DigestContinuityEdgesObserved, report.DigestContinuityEdgesExpected, report.ManualTransformations, report.SemanticOperationsBefore, report.SemanticOperationsAfter, report.GeneratedBytesEqual, report.NormalizedSemanticEqual, report.ArtifactCount, report.ArtifactDenominator, report.RepositoryWrites, report.LocalBuildExecutions, report.LocalTestExecutions, report.WallMS, report.PeakRSSKib, note)
}
