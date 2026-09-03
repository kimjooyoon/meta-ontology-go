package main

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

const retentionConsumeUsage = "usage: gooo consume <observation.gooo> --input <file.gooo> --observation FILE --proposal FILE --adoption FILE [--authorization FILE] [--certificate FILE] [--baseline FILE] --out <directory>"

type retentionConsumeOptions struct {
	contractFilename      string
	inputFilename         string
	observationFilename   string
	proposalFilename      string
	authorizationFilename string
	adoptionFilename      string
	certificateFilename   string
	baselineFilename      string
	outputDir             string
}

func runRetentionConsume(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	options, err := parseRetentionConsumeArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	inputs, diagnostics, err := loadObservationInputs(observeOptions{contractFilename: options.contractFilename, inputFilename: options.inputFilename}, reader, parser)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	if diagnostics.HasErrors() || !reportDiagnostics(diagnostics, stderr) {
		return exitFailure
	}
	evidence, err := readRetentionEvidence(retentionEvidenceOptions{
		observationFilename: options.observationFilename, proposalFilename: options.proposalFilename,
		authorizationFilename: options.authorizationFilename, adoptionFilename: options.adoptionFilename,
	}, reader)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention consume: %v\n", err)
		return exitFailure
	}
	compilerDigest, err := generation.SemanticRetentionCompilerDigest(reader.ReadFile)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention consume: compiler binding: %v\n", err)
		return exitFailure
	}
	verifierDigest, err := generation.SemanticRetentionVerifierDigest(reader.ReadFile)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention consume: verifier binding: %v\n", err)
		return exitFailure
	}
	base := retentionResultBase(inputs, evidence, compilerDigest, verifierDigest)
	if err := validateRetentionEvidence(inputs, evidence); err != nil {
		if _, isBindingMismatch := err.(retentionInputBindingError); isBindingMismatch {
			return writeRetentionResult(options.outputDir, retentionRefutedResult(base, ""), stdout, stderr)
		}
		fmt.Fprintf(stderr, "gooo: retention consume: %v\n", err)
		return exitFailure
	}
	if options.authorizationFilename == "" {
		return writeRetentionResult(options.outputDir, retentionUnknownResult(base, generation.SemanticRetentionUnknownAuthorizationReason, "AUTHORIZATION_REQUIRED"), stdout, stderr)
	}
	if !evidence.authorization.Authorized {
		return writeRetentionResult(options.outputDir, retentionUnknownResult(base, generation.SemanticRetentionUnknownAuthorizationReason, "AUTHORIZATION_REQUIRED"), stdout, stderr)
	}
	if err := validateRetentionAuthorization(evidence); err != nil {
		return writeRetentionResult(options.outputDir, retentionRefutedResult(base, ""), stdout, stderr)
	}
	if options.certificateFilename == "" {
		return writeRetentionResult(options.outputDir, retentionUnknownResult(base, generation.SemanticRetentionUnknownCertificateReason, "CERTIFICATE_REQUIRED"), stdout, stderr)
	}
	certificateData, err := reader.ReadFile(options.certificateFilename)
	if err != nil {
		return writeRetentionResult(options.outputDir, retentionUnknownResult(base, generation.SemanticRetentionUnknownCertificateReason, "CERTIFICATE_REQUIRED"), stdout, stderr)
	}
	certificateDigest := cache.HashBytes(certificateData).String()
	var certificate generation.SemanticRetentionCertificate
	if err := json.Unmarshal(certificateData, &certificate); err != nil {
		return writeRetentionResult(options.outputDir, retentionRefutedResult(base, certificateDigest), stdout, stderr)
	}
	expected := retentionBindings(inputs, evidence, compilerDigest, verifierDigest)
	if err := generation.VerifySemanticRetentionCertificate(certificate, expected); err != nil {
		return writeRetentionResult(options.outputDir, retentionRefutedResult(base, certificateDigest), stdout, stderr)
	}

	started := time.Now()
	var beforeMem, afterMem runtime.MemStats
	runtime.ReadMemStats(&beforeMem)
	baseline, err := readRetentionBaseline(options.baselineFilename, reader)
	if err != nil {
		return writeRetentionResult(options.outputDir, retentionRefutedResult(base, certificateDigest), stdout, stderr)
	}
	bytesEqual := false
	semanticEqual := false
	if options.baselineFilename != "" {
		bytesEqual = retentionResultEqualBytes(baseline, generation.SemanticRetentionResult{GeneratedSource: certificate.GeneratedSource})
		semanticEqual = baseline.NormalizedIRDigest == certificate.NormalizedIRDigest
		if baseline.Decision != "CLOSED" || baseline.Reason != generation.SemanticRetentionCertifiedReason ||
			baseline.CertificateDigest != certificateDigest || baseline.Metrics.SemanticOperationCount != 1 || baseline.Metrics.CertificateMisses != 1 ||
			!bytesEqual || !semanticEqual {
			return writeRetentionResult(options.outputDir, retentionRefutedResult(base, certificateDigest), stdout, stderr)
		}
	}
	runtime.ReadMemStats(&afterMem)
	after := base
	after.Lifecycle = "SEPARATE_INVOCATION_CONSUME"
	after.Decision = "CLOSED"
	after.Reason = generation.SemanticRetentionHitReason
	after.CertificateDigest = certificateDigest
	after.NormalizedIRDigest = certificate.NormalizedIRDigest
	after.GeneratedOutputDigest = certificate.GeneratedOutputDigest
	after.CompilerDigest = certificate.CompilerDigest
	after.ToolchainDigest = certificate.ToolchainDigest
	after.VerifierDigest = certificate.VerifierDigest
	after.PolicyDigest = certificate.PolicyDigest
	after.GeneratedSource = append([]byte(nil), certificate.GeneratedSource...)
	after.Metrics = retentionRuntimeMetrics(started, beforeMem, afterMem)
	after.Metrics.CertificateHits = 1
	after.Metrics.GeneratedBytesEqual = bytesEqual
	after.Metrics.NormalizedSemanticEqual = semanticEqual
	if err := generation.ValidateSemanticRetentionResult(after); err != nil {
		fmt.Fprintf(stderr, "gooo: retention consume: result validation: %v\n", err)
		return exitFailure
	}
	return writeRetentionResult(options.outputDir, after, stdout, stderr)
}

func readRetentionBaseline(filename string, reader SourceReader) (generation.SemanticRetentionResult, error) {
	if filename == "" {
		return generation.SemanticRetentionResult{}, nil
	}
	data, err := reader.ReadFile(filename)
	if err != nil {
		return generation.SemanticRetentionResult{}, fmt.Errorf("read baseline: %w", err)
	}
	var result generation.SemanticRetentionResult
	if err := json.Unmarshal(data, &result); err != nil {
		return generation.SemanticRetentionResult{}, fmt.Errorf("decode baseline: %w", err)
	}
	if err := generation.ValidateSemanticRetentionResult(result); err != nil {
		return generation.SemanticRetentionResult{}, fmt.Errorf("validate baseline: %w", err)
	}
	return result, nil
}

func parseRetentionConsumeArguments(args []string) (retentionConsumeOptions, error) {
	if len(args) == 0 {
		return retentionConsumeOptions{}, fmt.Errorf("%s", retentionConsumeUsage)
	}
	options := retentionConsumeOptions{contractFilename: args[0]}
	for index := 1; index < len(args); index++ {
		if index+1 >= len(args) || args[index+1] == "" {
			return retentionConsumeOptions{}, fmt.Errorf("%s", retentionConsumeUsage)
		}
		value := args[index+1]
		switch args[index] {
		case "--input":
			if options.inputFilename != "" {
				return retentionConsumeOptions{}, fmt.Errorf("%s", retentionConsumeUsage)
			}
			options.inputFilename = value
		case "--observation":
			if options.observationFilename != "" {
				return retentionConsumeOptions{}, fmt.Errorf("%s", retentionConsumeUsage)
			}
			options.observationFilename = value
		case "--proposal":
			if options.proposalFilename != "" {
				return retentionConsumeOptions{}, fmt.Errorf("%s", retentionConsumeUsage)
			}
			options.proposalFilename = value
		case "--authorization":
			if options.authorizationFilename != "" {
				return retentionConsumeOptions{}, fmt.Errorf("%s", retentionConsumeUsage)
			}
			options.authorizationFilename = value
		case "--adoption":
			if options.adoptionFilename != "" {
				return retentionConsumeOptions{}, fmt.Errorf("%s", retentionConsumeUsage)
			}
			options.adoptionFilename = value
		case "--certificate":
			if options.certificateFilename != "" {
				return retentionConsumeOptions{}, fmt.Errorf("%s", retentionConsumeUsage)
			}
			options.certificateFilename = value
		case "--baseline":
			if options.baselineFilename != "" {
				return retentionConsumeOptions{}, fmt.Errorf("%s", retentionConsumeUsage)
			}
			options.baselineFilename = value
		case "--out":
			if options.outputDir != "" {
				return retentionConsumeOptions{}, fmt.Errorf("%s", retentionConsumeUsage)
			}
			options.outputDir = value
		default:
			return retentionConsumeOptions{}, fmt.Errorf("%s", retentionConsumeUsage)
		}
		index++
	}
	if options.contractFilename == "" || options.inputFilename == "" || options.observationFilename == "" ||
		options.proposalFilename == "" || options.adoptionFilename == "" || options.outputDir == "" {
		return retentionConsumeOptions{}, fmt.Errorf("%s", retentionConsumeUsage)
	}
	return options, nil
}
