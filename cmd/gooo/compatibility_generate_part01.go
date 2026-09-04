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
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compatibilitypolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/compilercompatibility"
)

func runCompatibilityGenerate(options generateOptions, input generateInput, reader SourceReader, jsonMode bool, stdout, stderr io.Writer, deadline time.Time) int {
	started := time.Now()
	policySource, err := reader.ReadFile(compilercompatibility.CanonicalPolicyPath)
	if err != nil {
		return writeCompatibilityFailure(options, input, "REFUTED", compilercompatibility.ReasonAxisMismatch, "", started, jsonMode, stdout, stderr)
	}
	policy, err := compatibilitypolicy.Load(compilercompatibility.CanonicalPolicyPath, policySource)
	if err != nil {
		return writeCompatibilityFailure(options, input, "REFUTED", compilercompatibility.ReasonAxisMismatch, "", started, jsonMode, stdout, stderr)
	}
	certificateData, err := reader.ReadFile(options.compatibilityCertificateFilename)
	if err != nil {
		return writeCompatibilityFailureWithPolicy(options, input, policy, "UNKNOWN", compilercompatibility.ReasonMissingCertificate, "", started, jsonMode, stdout, stderr)
	}
	certificateDigest := cache.HashBytes(certificateData).String()
	certificate, err := decodeCompatibilityCertificate(certificateData)
	if err != nil {
		return writeCompatibilityFailureWithPolicy(options, input, policy, "REFUTED", compilercompatibility.ReasonTamperedCertificate, certificateDigest, started, jsonMode, stdout, stderr)
	}
	generated, err := generateWithDeadlineCore(input.file, nil, remainingDeadline(deadline))
	if err != nil {
		return writeCompatibilityFailureWithPolicy(options, input, policy, "UNKNOWN", compilercompatibility.ReasonMissingSuccessorReplay, certificateDigest, started, jsonMode, stdout, stderr)
	}
	normalizedDigest, err := cache.SemanticDigest(generated.ir)
	if err != nil {
		return writeCompatibilityFailureWithPolicy(options, input, policy, "UNKNOWN", compilercompatibility.ReasonMissingSuccessorReplay, certificateDigest, started, jsonMode, stdout, stderr)
	}
	manifest, err := buildProjectionManifest(options.filename, generatedFileName, input.source, nil, generated.ir, generated.result)
	if err != nil {
		return writeCompatibilityFailureWithPolicy(options, input, policy, "UNKNOWN", compilercompatibility.ReasonMissingSuccessorReplay, certificateDigest, started, jsonMode, stdout, stderr)
	}
	manifestData, err := jsonManifestBytes(manifest)
	if err != nil {
		return writeCompatibilityFailureWithPolicy(options, input, policy, "UNKNOWN", compilercompatibility.ReasonMissingSuccessorReplay, certificateDigest, started, jsonMode, stdout, stderr)
	}
	compilerDigest, err := compilercompatibility.CompilerImplementationDigest(reader.ReadFile)
	if err != nil {
		return writeCompatibilityFailureWithPolicy(options, input, policy, "UNKNOWN", compilercompatibility.ReasonMissingSuccessorReplay, certificateDigest, started, jsonMode, stdout, stderr)
	}
	inputDigest := cache.HashBytes(input.source).String()
	current := compilercompatibility.ExecutionReceipt{
		Schema: compilercompatibility.ConsumptionSchema, Role: "successor-consumer-replay",
		CandidateStableID: certificate.CandidateStableID, SubjectDigest: inputDigest, SourceDigest: inputDigest,
		SemanticIRDigest: normalizedDigest.String(), GeneratedOutputDigest: cache.HashBytes(generated.result.Source).String(),
		GeneratedManifestDigest: cache.HashBytes(manifestData).String(), GeneratedSource: append([]byte(nil), generated.result.Source...), GeneratedManifest: append([]byte(nil), manifestData...),
		PolicyDigest: policy.SourceDigest, PolicyEvaluatorDigest: policy.EvaluatorDigest, PolicyResult: compilercompatibility.DecisionClosed,
		CompilerImplementationDigest: compilerDigest, GoToolchainDigest: compilercompatibility.CurrentToolchainDigest(),
		TestContractDigest: compilercompatibility.TestContractDigest(), TestContractResult: "NOT_EXECUTED",
		AuthorizationDigest: certificate.AuthorizationDigest,
	}
	evaluation := compilercompatibility.EvaluateOptIn(policy, compilercompatibility.Request{Mode: "OPT_IN", CandidateStableID: certificate.CandidateStableID,
		SubjectDigest: inputDigest, SourceDigest: inputDigest, Current: current, Certificate: &certificate})
	if evaluation.Decision != compilercompatibility.DecisionClosed {
		return writeCompatibilityFailureWithEvaluation(options, input, policy, evaluation, certificateDigest, started, jsonMode, stdout, stderr)
	}
	output, manifestPath, err := publicGeneratePaths(options)
	if err != nil {
		return writeCompatibilityFailureWithEvaluation(options, input, policy, compilercompatibility.Evaluation{Decision: compilercompatibility.DecisionRefuted, Reason: compilercompatibility.ReasonAxisMismatch, MismatchDetected: true}, certificateDigest, started, jsonMode, stdout, stderr)
	}
	report := compatibilityConsumptionReport(policy, certificate, certificateDigest, current, evaluation, output, manifestPath, started)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: compatibility generation: encode report: %v\n", err)
		return exitFailure
	}
	data = append(data, '\n')
	human := []byte(renderCompatibilityReport(report))
	if err := ensurePublicOutputDirectory(options.outputDir); err != nil {
		fmt.Fprintf(stderr, "gooo: compatibility generation: output: %v\n", err)
		return exitFailure
	}
	writes := []atomicWrite{{path: output, data: generated.result.Source}, {path: manifestPath, data: manifestData},
		{path: filepath.Join(options.outputDir, "generation-report.json"), data: data},
		{path: filepath.Join(options.outputDir, "generation-report.md"), data: human}}
	if err := writeAtomicFiles(writes); err != nil {
		fmt.Fprintf(stderr, "gooo: compatibility generation: output: %v\n", err)
		return exitFailure
	}
	if jsonMode {
		_, _ = stdout.Write(data)
	} else {
		fmt.Fprintf(stdout, "generated: %s\ncompatibility: %s\nreport: %s\n", output, report.Reason, filepath.Join(options.outputDir, "generation-report.md"))
	}
	return exitOK
}

func decodeCompatibilityCertificate(data []byte) (compilercompatibility.Certificate, error) {
	var certificate compilercompatibility.Certificate
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&certificate); err != nil {
		return certificate, err
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return certificate, errors.New("compatibility certificate contains multiple JSON values")
	} else if !errors.Is(err, io.EOF) {
		return certificate, err
	}
	if err := compilercompatibility.ValidateCertificate(certificate); err != nil {
		return certificate, err
	}
	return certificate, nil
}

func writeCompatibilityFailure(options generateOptions, input generateInput, decision, reason, certificateDigest string, started time.Time, jsonMode bool, stdout, stderr io.Writer) int {
	return writeCompatibilityFailureWithPolicy(options, input, compatibilitypolicy.Policy{}, decision, reason, certificateDigest, started, jsonMode, stdout, stderr)
}

func writeCompatibilityFailureWithPolicy(options generateOptions, input generateInput, policy compatibilitypolicy.Policy, decision, reason, certificateDigest string, started time.Time, jsonMode bool, stdout, stderr io.Writer) int {
	evaluation := compilercompatibility.Evaluation{Decision: decision, Reason: reason}
	if decision == compilercompatibility.DecisionUnknown {
		evaluation.Unknown = &compilercompatibility.UnknownState{Stage: "COMPATIBILITY", Step: "LOAD_COMPATIBILITY_CERTIFICATE", Reason: reason, UnknownClass: "INCOMPLETE_EVIDENCE", NextOperation: "PROVIDE_BOUNDED_SUCCESSOR_CERTIFICATE", BlockedBy: []string{"compatibility_certificate"}}
	}
	return writeCompatibilityFailureWithEvaluation(options, input, policy, evaluation, certificateDigest, started, jsonMode, stdout, stderr)
}

func writeCompatibilityFailureWithEvaluation(options generateOptions, input generateInput, policy compatibilitypolicy.Policy, evaluation compilercompatibility.Evaluation, certificateDigest string, started time.Time, jsonMode bool, stdout, stderr io.Writer) int {
	report := compilercompatibility.ConsumptionReport{Schema: compilercompatibility.ReportSchema, Lifecycle: "COMPATIBILITY_CONSUMPTION", Decision: evaluation.Decision,
		Reason: evaluation.Reason, Unknown: evaluation.Unknown, CertificateDigest: certificateDigest, SubjectDigest: cache.HashBytes(input.source).String(),
		SourceDigest: cache.HashBytes(input.source).String(), IdentityAxisCount: compatibilitypolicy.AxisCount, AxisComparisons: evaluation.Axes,
		CompatibilityHits: 0, CompatibilityMisses: 1, EvidenceArtifactCount: policy.EvidenceArtifacts, ContinuityEdgeCount: policy.ContinuityEdges,
		Claim: "BOUNDED_IMPLEMENTATION_REUSE_ELIGIBILITY", PerformanceClaim: false, GeneralCompatibilityClaim: false,
		UnsupportedFrontierDecision: compilercompatibility.UnsupportedFrontierDecision, UnsupportedFrontierClaims: append([]string(nil), compilercompatibility.UnsupportedFrontierClaims...),
		ArtifactCount: 2, RepositoryWrites: 0, LocalTestExecutions: 0, WallMS: compatibilityWallMS(started), PeakRSSKib: readPeakRSSKib()}
	if report.EvidenceArtifactCount == 0 {
		report.EvidenceArtifactCount = compatibilitypolicy.EvidenceArtifactCount
		report.ContinuityEdgeCount = compatibilitypolicy.ContinuityEdgeCount
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: compatibility generation: encode report: %v\n", err)
		return exitFailure
	}
	data = append(data, '\n')
	human := []byte(renderCompatibilityReport(report))
	if err := ensurePublicOutputDirectory(options.outputDir); err != nil {
		fmt.Fprintf(stderr, "gooo: compatibility generation: output: %v\n", err)
		return exitFailure
	}
	writes := []atomicWrite{{path: filepath.Join(options.outputDir, "generation-report.json"), data: data},
		{path: filepath.Join(options.outputDir, "generation-report.md"), data: human}}
	if err := writeAtomicFiles(writes); err != nil {
		fmt.Fprintf(stderr, "gooo: compatibility generation: output: %v\n", err)
		return exitFailure
	}
	if jsonMode {
		_, _ = stdout.Write(data)
	} else {
		fmt.Fprintf(stdout, "compatibility: %s (%s)\nreport: %s\n", filepath.Join(options.outputDir, "generation-report.json"), evaluation.Decision, filepath.Join(options.outputDir, "generation-report.md"))
	}
	return exitOK
}
