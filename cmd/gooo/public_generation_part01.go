package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type publicRetentionContext struct {
	inputs         observationInputs
	evidence       retentionEvidence
	compilerDigest string
	verifierDigest string
}

func (options generateOptions) publicRetentionRequested() bool {
	return options.retainedCertificateFilename != "" || options.retentionContractFilename != "" ||
		options.retentionObservationFilename != "" || options.retentionProposalFilename != "" ||
		options.retentionAuthorizationFilename != "" || options.retentionAdoptionFilename != ""
}

func runPublicGenerate(options generateOptions, input generateInput, reader SourceReader, parser SourceParser, jsonMode bool, stdout, stderr io.Writer, deadline time.Time) int {
	started := time.Now()
	context, err := loadPublicRetentionContext(options, input, reader, parser, deadline)
	if err != nil {
		var bindingError retentionInputBindingError
		if errors.As(err, &bindingError) && context.inputs.inputSource != nil {
			return writePublicRefuted(options, context, "", started, jsonMode, stdout, stderr)
		}
		fmt.Fprintf(stderr, "gooo: generate retained knowledge: %v\n", err)
		return exitFailure
	}
	if len(context.evidence.authorizationData) == 0 || !context.evidence.authorization.Authorized {
		return writePublicUnknown(options, context, generation.SemanticRetentionUnknownAuthorizationReason, started, jsonMode, stdout, stderr)
	}
	if options.retainedCertificateFilename == "" {
		return writePublicUnknown(options, context, generation.SemanticRetentionUnknownCertificateReason, started, jsonMode, stdout, stderr)
	}
	certificateData, err := reader.ReadFile(options.retainedCertificateFilename)
	if err != nil {
		return writePublicUnknown(options, context, generation.SemanticRetentionUnknownCertificateReason, started, jsonMode, stdout, stderr)
	}
	certificateDigest := cache.HashBytes(certificateData).String()
	var certificate generation.SemanticRetentionCertificate
	if err := json.Unmarshal(certificateData, &certificate); err != nil {
		return writePublicRefuted(options, context, certificateDigest, started, jsonMode, stdout, stderr)
	}
	if options.previousGo != "" || !publicCertificateMatches(certificate, context) {
		return writePublicRefuted(options, context, certificateDigest, started, jsonMode, stdout, stderr)
	}
	return writePublicCertificate(options, context, certificate, certificateDigest, started, jsonMode, stdout, stderr)
}

func loadPublicRetentionContext(options generateOptions, input generateInput, reader SourceReader, parser SourceParser, deadline time.Time) (publicRetentionContext, error) {
	if options.retentionContractFilename == "" || options.retentionObservationFilename == "" || options.retentionProposalFilename == "" || options.retentionAdoptionFilename == "" {
		return publicRetentionContext{}, errors.New("retention contract, observation, proposal, and adoption are required")
	}
	contractSource, err := reader.ReadFile(options.retentionContractFilename)
	if err != nil {
		return publicRetentionContext{}, fmt.Errorf("read retention contract: %w", err)
	}
	contract, err := generation.ParseSemanticObservationContract(contractSource)
	if err != nil {
		return publicRetentionContext{}, fmt.Errorf("retention contract: %w", err)
	}
	contractFile, diagnostics, err := parseWithDeadline(parser, options.retentionContractFilename, string(contractSource), remainingDeadline(deadline))
	if err != nil {
		return publicRetentionContext{}, fmt.Errorf("parse retention contract: %w", err)
	}
	if diagnostics.HasErrors() {
		return publicRetentionContext{}, errors.New("parse retention contract: diagnostics contain errors")
	}
	if err := validatePublicRetentionPolicy(contractFile); err != nil {
		return publicRetentionContext{}, err
	}
	inputs := observationInputs{contractSource: contractSource, inputSource: input.source, contract: contract, file: input.file}
	evidence, err := readRetentionEvidence(retentionEvidenceOptions{
		observationFilename: options.retentionObservationFilename, proposalFilename: options.retentionProposalFilename,
		authorizationFilename: options.retentionAuthorizationFilename, adoptionFilename: options.retentionAdoptionFilename,
	}, reader)
	if err != nil {
		return publicRetentionContext{}, err
	}
	compilerDigest, err := generation.SemanticRetentionCompilerDigest(reader.ReadFile)
	if err != nil {
		return publicRetentionContext{}, fmt.Errorf("compiler binding: %w", err)
	}
	verifierDigest, err := generation.SemanticRetentionVerifierDigest(reader.ReadFile)
	if err != nil {
		return publicRetentionContext{}, fmt.Errorf("verifier binding: %w", err)
	}
	context := publicRetentionContext{inputs: inputs, evidence: evidence, compilerDigest: compilerDigest, verifierDigest: verifierDigest}
	if err := validateRetentionEvidence(inputs, evidence); err != nil {
		return context, err
	}
	return context, nil
}

func validatePublicRetentionPolicy(file *syntax.File) error {
	if file == nil {
		return errors.New("retention policy is missing")
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, syntax.EntityFieldsV1Support())
	if err != nil {
		return fmt.Errorf("retention policy lowering failed: %w", err)
	}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.Name != "PublishOperationReceipt" {
			continue
		}
		program := node.ValueProgram
		for _, marker := range []string{"public-generate=baseline-or-retained", "certificate=explicit-only", "invalid=fail-closed", "fallback=none"} {
			if !strings.Contains(program, marker) {
				return fmt.Errorf("retention policy omits %s", marker)
			}
		}
		return nil
	}
	return errors.New("retention policy activity is missing")
}

func publicCertificateMatches(certificate generation.SemanticRetentionCertificate, context publicRetentionContext) bool {
	expected := retentionBindings(context.inputs, context.evidence, context.compilerDigest, context.verifierDigest)
	return generation.VerifySemanticRetentionCertificate(certificate, expected) == nil
}

func writePublicCertificate(options generateOptions, context publicRetentionContext, certificate generation.SemanticRetentionCertificate, certificateDigest string, started time.Time, jsonMode bool, stdout, stderr io.Writer) int {
	output, manifest, err := publicGeneratePaths(options)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: generate retained knowledge: %v\n", err)
		return exitFailure
	}
	report := publicRetentionReport(context, certificate, certificateDigest, output, manifest)
	writes := []atomicWrite{{path: output, data: certificate.GeneratedSource}, {path: manifest, data: certificate.GeneratedManifest}}
	return writePublicReportAndArtifacts(options.outputDir, writes, report, started, jsonMode, stdout, stderr)
}

func writePublicUnknown(options generateOptions, context publicRetentionContext, reason string, started time.Time, jsonMode bool, stdout, stderr io.Writer) int {
	report := publicReportBase(context)
	report.Lifecycle = generation.SemanticPublicGenerationFailClosed
	report.Decision = "UNKNOWN"
	report.Reason = reason
	report.Unknown = generation.SemanticRetentionUnknownState(reason)
	return writePublicReportAndArtifacts(options.outputDir, nil, report, started, jsonMode, stdout, stderr)
}

func writePublicRefuted(options generateOptions, context publicRetentionContext, certificateDigest string, started time.Time, jsonMode bool, stdout, stderr io.Writer) int {
	report := publicReportBase(context)
	report.Lifecycle = generation.SemanticPublicGenerationFailClosed
	report.Decision = generation.SemanticRetentionRefuted
	report.Reason = generation.SemanticRetentionRefutedReason
	report.CertificateDigest = certificateDigest
	report.Metrics.CertificateMisses = 1
	return writePublicReportAndArtifacts(options.outputDir, nil, report, started, jsonMode, stdout, stderr)
}

func publicReportBase(context publicRetentionContext) generation.SemanticPublicGenerationReport {
	bindings := retentionBindings(context.inputs, context.evidence, context.compilerDigest, context.verifierDigest)
	return generation.SemanticPublicGenerationReport{
		Schema: generation.SemanticPublicGenerationReportSchema, AdoptionReportDigest: bindings.AdoptionReportDigest,
		ObservationDigest: bindings.ObservationDigest, ProposalDigest: bindings.ProposalDigest,
		AuthorizationDigest: bindings.AuthorizationDigest, CandidateStableID: bindings.CandidateStableID,
		ContractSourceDigest: bindings.ContractSourceDigest, InputSourceDigest: bindings.InputSourceDigest,
		CompilerDigest: bindings.CompilerDigest, ToolchainDigest: bindings.ToolchainDigest,
		VerifierDigest: bindings.VerifierDigest, PolicyDigest: bindings.PolicyDigest,
		RepositoryWrites: 0, LocalTestExecutions: 0,
	}
}

func publicRetentionReport(context publicRetentionContext, certificate generation.SemanticRetentionCertificate, certificateDigest, output, manifest string) generation.SemanticPublicGenerationReport {
	report := publicReportBase(context)
	report.Lifecycle = generation.SemanticPublicGenerationRetainedLifecycle
	report.Decision = "CLOSED"
	report.Reason = generation.SemanticPublicGenerationHitReason
	report.CertificateDigest = certificateDigest
	report.NormalizedIRDigest = certificate.NormalizedIRDigest
	report.GeneratedOutputDigest = certificate.GeneratedOutputDigest
	report.GeneratedManifestDigest = certificate.GeneratedManifestDigest
	report.OutputFile = output
	report.ManifestFile = manifest
	report.ArtifactCount = generation.SemanticPublicGenerationArtifactsClosed
	report.Metrics.SemanticOperationCount = 0
	report.Metrics.CertificateHits = 1
	report.Metrics.GeneratedBytesEqual = true
	report.Metrics.NormalizedSemanticEqual = true
	return report
}

func publicGeneratePaths(options generateOptions) (string, string, error) {
	root, err := canonicalOutputRoot(options.outputDir)
	if err != nil {
		return "", "", err
	}
	output, err := resolveOutputPath(root, generatedFileName)
	if err != nil {
		return "", "", err
	}
	manifest := options.manifestPath
	if manifest == "" {
		manifest = filepath.Join(options.outputDir, generatedManifestFileName)
	}
	manifest, err = resolveManifestPath(root, manifest)
	return output, manifest, err
}

func writePublicReportAndArtifacts(outputDir string, writes []atomicWrite, report generation.SemanticPublicGenerationReport, started time.Time, jsonMode bool, stdout, stderr io.Writer) int {
	if err := ensurePublicOutputDirectory(outputDir); err != nil {
		fmt.Fprintf(stderr, "gooo: generate retained knowledge: %v\n", err)
		return exitFailure
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	report.Metrics.WallMS = retentionWallMS(started)
	runtime.ReadMemStats(&after)
	report.Metrics.AllocationCount = int64(after.Mallocs - before.Mallocs)
	report.Metrics.AllocationBytes = int64(after.TotalAlloc - before.TotalAlloc)
	report.Metrics.PeakRSSKib = readPeakRSSKib()
	if report.Decision == "UNKNOWN" || report.Decision == generation.SemanticRetentionRefuted {
		report.Metrics = generation.SemanticRetentionRuntimeMetrics{}
		if report.Decision == generation.SemanticRetentionRefuted {
			report.Metrics.CertificateMisses = 1
		}
	}
	jsonPath := filepath.Join(outputDir, "generation-report.json")
	markdownPath := filepath.Join(outputDir, "generation-report.md")
	allWrites := append([]atomicWrite(nil), writes...)
	allWrites = append(allWrites, atomicWrite{path: jsonPath}, atomicWrite{path: markdownPath})
	report.OutputFileCount = len(allWrites)
	report.ArtifactCount = len(allWrites)
	for range 3 {
		jsonData, err := publicReportJSON(report)
		if err != nil {
			fmt.Fprintf(stderr, "gooo: generate retained knowledge: %v\n", err)
			return exitFailure
		}
		markdownData := []byte(publicReportMarkdown(report))
		allWrites[len(writes)].data = jsonData
		allWrites[len(writes)+1].data = markdownData
		var total int64
		for _, write := range allWrites {
			total += int64(len(write.data))
		}
		if report.OutputBytes == total {
			break
		}
		report.OutputBytes = total
	}
	jsonData, err := publicReportJSON(report)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: generate retained knowledge: %v\n", err)
		return exitFailure
	}
	allWrites[len(writes)].data = jsonData
	allWrites[len(writes)+1].data = []byte(publicReportMarkdown(report))
	if err := generation.ValidateSemanticPublicGenerationReport(report); err != nil {
		fmt.Fprintf(stderr, "gooo: generate retained knowledge: %v\n", err)
		return exitFailure
	}
	if err := writeAtomicFiles(allWrites); err != nil {
		fmt.Fprintf(stderr, "gooo: generate retained knowledge output: %v\n", err)
		return exitFailure
	}
	if jsonMode {
		_, _ = stdout.Write(jsonData)
	} else if report.Decision == "CLOSED" {
		fmt.Fprintf(stdout, "generated: %s\nretained knowledge: %s\nreport: %s\n", report.OutputFile, report.Reason, markdownPath)
	} else {
		fmt.Fprintf(stdout, "retention: %s (%s)\nreport: %s\n", jsonPath, report.Decision, markdownPath)
	}
	return exitOK
}
