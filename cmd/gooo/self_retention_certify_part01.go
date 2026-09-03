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

const retentionCertifyUsage = "usage: gooo certify <observation.gooo> --input <file.gooo> --observation FILE --proposal FILE --authorization FILE --adoption FILE --out <directory>"

type retentionCertifyOptions struct {
	contractFilename      string
	inputFilename         string
	observationFilename   string
	proposalFilename      string
	authorizationFilename string
	adoptionFilename      string
	outputDir             string
}

func runRetentionCertify(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	options, err := parseRetentionCertifyArguments(args)
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
	context, err := prepareRetentionCertify(inputs, options, reader)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: %v\n", err)
		return exitFailure
	}
	certification, err := executeRetentionCertification(context)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: %v\n", err)
		return exitFailure
	}
	beforeData, err := retentionBeforeData(context, certification)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: %v\n", err)
		return exitFailure
	}
	if err := writeRetentionArtifacts(options.outputDir, []retentionArtifact{
		{name: "retained-knowledge-certificate.json", data: certification.certificateData},
		{name: "retention-before.json", data: beforeData},
	}); err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: output: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "retention certificate: %s\nretention before: %s\n", options.outputDir+"/retained-knowledge-certificate.json", options.outputDir+"/retention-before.json")
	return exitOK
}

type retentionCertifyContext struct {
	inputs         observationInputs
	evidence       retentionEvidence
	compilerDigest string
	verifierDigest string
}

type retentionCertification struct {
	certificateData   []byte
	normalizedDigest  string
	outputDigest      string
	generatedSource   []byte
	certificateDigest string
	metrics           generation.SemanticRetentionRuntimeMetrics
}

func prepareRetentionCertify(inputs observationInputs, options retentionCertifyOptions, reader SourceReader) (retentionCertifyContext, error) {
	evidence, err := readRetentionEvidence(retentionEvidenceOptions{
		observationFilename: options.observationFilename, proposalFilename: options.proposalFilename,
		authorizationFilename: options.authorizationFilename, adoptionFilename: options.adoptionFilename,
	}, reader)
	if err != nil {
		return retentionCertifyContext{}, err
	}
	if err := validateRetentionEvidence(inputs, evidence); err != nil {
		return retentionCertifyContext{}, err
	}
	if err := validateRetentionAuthorization(evidence); err != nil {
		return retentionCertifyContext{}, err
	}
	if !evidence.authorization.Authorized {
		return retentionCertifyContext{}, fmt.Errorf("explicit authorization is required")
	}
	compilerDigest, err := generation.SemanticRetentionCompilerDigest(reader.ReadFile)
	if err != nil {
		return retentionCertifyContext{}, fmt.Errorf("compiler binding: %w", err)
	}
	verifierDigest, err := generation.SemanticRetentionVerifierDigest(reader.ReadFile)
	if err != nil {
		return retentionCertifyContext{}, fmt.Errorf("verifier binding: %w", err)
	}
	return retentionCertifyContext{inputs: inputs, evidence: evidence, compilerDigest: compilerDigest, verifierDigest: verifierDigest}, nil
}

func executeRetentionCertification(context retentionCertifyContext) (retentionCertification, error) {
	started := time.Now()
	var beforeMem, afterMem runtime.MemStats
	runtime.ReadMemStats(&beforeMem)
	generated, err := generateWithDeadlineCore(context.inputs.file, nil, commandDeadline)
	runtime.ReadMemStats(&afterMem)
	if err != nil {
		return retentionCertification{}, fmt.Errorf("compiler execution: %w", err)
	}
	normalizedDigest, err := cache.SemanticDigest(generated.ir)
	if err != nil {
		return retentionCertification{}, fmt.Errorf("normalized IR digest: %w", err)
	}
	outputDigest := cache.HashBytes(generated.result.Source).String()
	if normalizedDigest.String() != context.evidence.proposal.Candidate.InputDigest ||
		outputDigest != context.evidence.adoption.Evidence.BeforeOutputDigest ||
		normalizedDigest.String() != context.evidence.adoption.Evidence.BeforeSemanticDigest {
		return retentionCertification{}, fmt.Errorf("compiler result contradicts the authorized adoption evidence")
	}
	certificate, err := buildRetentionCertificate(context, normalizedDigest.String(), outputDigest, generated.result.Source)
	if err != nil {
		return retentionCertification{}, err
	}
	certificateData, err := json.MarshalIndent(certificate, "", "  ")
	if err != nil {
		return retentionCertification{}, fmt.Errorf("encode certificate: %w", err)
	}
	certificateData = append(certificateData, '\n')
	return retentionCertification{
		certificateData: certificateData, normalizedDigest: normalizedDigest.String(), outputDigest: outputDigest,
		generatedSource: append([]byte(nil), generated.result.Source...), certificateDigest: cache.HashBytes(certificateData).String(),
		metrics: retentionRuntimeMetrics(started, beforeMem, afterMem),
	}, nil
}

func buildRetentionCertificate(context retentionCertifyContext, normalizedDigest, outputDigest string, source []byte) (generation.SemanticRetentionCertificate, error) {
	bindings := retentionBindings(context.inputs, context.evidence, context.compilerDigest, context.verifierDigest)
	certificate := generation.SemanticRetentionCertificate{
		Schema: generation.SemanticRetentionCertificateSchema, AdoptionReportDigest: bindings.AdoptionReportDigest,
		ObservationDigest: bindings.ObservationDigest, ProposalDigest: bindings.ProposalDigest,
		AuthorizationDigest: bindings.AuthorizationDigest, CandidateStableID: bindings.CandidateStableID,
		Target: generation.SemanticRetentionTarget, Mode: generation.SemanticRetentionMode,
		ContractSourceDigest: bindings.ContractSourceDigest, InputSourceDigest: bindings.InputSourceDigest,
		NormalizedIRDigest: normalizedDigest, GeneratedOutputDigest: outputDigest,
		CompilerDigest: bindings.CompilerDigest, ToolchainDigest: bindings.ToolchainDigest,
		VerifierDigest: bindings.VerifierDigest, PolicyDigest: bindings.PolicyDigest,
		GeneratedSource: append([]byte(nil), source...), RepositoryWrites: 0, LocalTestExecutions: 0,
	}
	var err error
	certificate.CertificateID, err = generation.SemanticRetentionCertificateContentDigest(certificate)
	if err != nil {
		return generation.SemanticRetentionCertificate{}, fmt.Errorf("certificate identity: %w", err)
	}
	if err := generation.ValidateSemanticRetentionCertificate(certificate); err != nil {
		return generation.SemanticRetentionCertificate{}, fmt.Errorf("certificate validation: %w", err)
	}
	return certificate, nil
}

func retentionBeforeData(context retentionCertifyContext, certification retentionCertification) ([]byte, error) {
	before := retentionResultBase(context.inputs, context.evidence, context.compilerDigest, context.verifierDigest)
	before.Lifecycle = "CERTIFY_BEFORE"
	before.Decision = "CLOSED"
	before.Reason = generation.SemanticRetentionCertifiedReason
	before.CertificateDigest = certification.certificateDigest
	before.NormalizedIRDigest = certification.normalizedDigest
	before.GeneratedOutputDigest = certification.outputDigest
	before.GeneratedSource = append([]byte(nil), certification.generatedSource...)
	before.Metrics = certification.metrics
	before.Metrics.SemanticOperationCount = 1
	before.Metrics.CertificateMisses = 1
	before.Metrics.GeneratedBytesEqual = true
	before.Metrics.NormalizedSemanticEqual = true
	if err := generation.ValidateSemanticRetentionResult(before); err != nil {
		return nil, fmt.Errorf("result validation: %w", err)
	}
	data, err := json.MarshalIndent(before, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode before result: %w", err)
	}
	return append(data, '\n'), nil
}

func parseRetentionCertifyArguments(args []string) (retentionCertifyOptions, error) {
	if len(args) == 0 {
		return retentionCertifyOptions{}, fmt.Errorf("%s", retentionCertifyUsage)
	}
	options := retentionCertifyOptions{contractFilename: args[0]}
	for index := 1; index < len(args); index++ {
		if index+1 >= len(args) || args[index+1] == "" {
			return retentionCertifyOptions{}, fmt.Errorf("%s", retentionCertifyUsage)
		}
		value := args[index+1]
		switch args[index] {
		case "--input":
			if options.inputFilename != "" {
				return retentionCertifyOptions{}, fmt.Errorf("%s", retentionCertifyUsage)
			}
			options.inputFilename = value
		case "--observation":
			if options.observationFilename != "" {
				return retentionCertifyOptions{}, fmt.Errorf("%s", retentionCertifyUsage)
			}
			options.observationFilename = value
		case "--proposal":
			if options.proposalFilename != "" {
				return retentionCertifyOptions{}, fmt.Errorf("%s", retentionCertifyUsage)
			}
			options.proposalFilename = value
		case "--authorization":
			if options.authorizationFilename != "" {
				return retentionCertifyOptions{}, fmt.Errorf("%s", retentionCertifyUsage)
			}
			options.authorizationFilename = value
		case "--adoption":
			if options.adoptionFilename != "" {
				return retentionCertifyOptions{}, fmt.Errorf("%s", retentionCertifyUsage)
			}
			options.adoptionFilename = value
		case "--out":
			if options.outputDir != "" {
				return retentionCertifyOptions{}, fmt.Errorf("%s", retentionCertifyUsage)
			}
			options.outputDir = value
		default:
			return retentionCertifyOptions{}, fmt.Errorf("%s", retentionCertifyUsage)
		}
		index++
	}
	if options.contractFilename == "" || options.inputFilename == "" || options.observationFilename == "" ||
		options.proposalFilename == "" || options.authorizationFilename == "" || options.adoptionFilename == "" || options.outputDir == "" {
		return retentionCertifyOptions{}, fmt.Errorf("%s", retentionCertifyUsage)
	}
	return options, nil
}
