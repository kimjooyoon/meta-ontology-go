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
	evidence, err := readRetentionEvidence(retentionEvidenceOptions{
		observationFilename: options.observationFilename, proposalFilename: options.proposalFilename,
		authorizationFilename: options.authorizationFilename, adoptionFilename: options.adoptionFilename,
	}, reader)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: %v\n", err)
		return exitFailure
	}
	if err := validateRetentionEvidence(inputs, evidence); err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: %v\n", err)
		return exitFailure
	}
	if err := validateRetentionAuthorization(evidence); err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: %v\n", err)
		return exitFailure
	}
	if !evidence.authorization.Authorized {
		fmt.Fprintln(stderr, "gooo: retention certification: explicit authorization is required")
		return exitFailure
	}
	compilerDigest, err := generation.SemanticRetentionCompilerDigest(reader.ReadFile)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: compiler binding: %v\n", err)
		return exitFailure
	}
	verifierDigest, err := generation.SemanticRetentionVerifierDigest(reader.ReadFile)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: verifier binding: %v\n", err)
		return exitFailure
	}

	started := time.Now()
	var beforeMem, afterMem runtime.MemStats
	runtime.ReadMemStats(&beforeMem)
	generated, err := generateWithDeadlineCore(inputs.file, nil, commandDeadline)
	runtime.ReadMemStats(&afterMem)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: compiler execution: %v\n", err)
		return exitFailure
	}
	normalizedDigest, err := cache.SemanticDigest(generated.ir)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: normalized IR digest: %v\n", err)
		return exitFailure
	}
	outputDigest := cache.HashBytes(generated.result.Source).String()
	if normalizedDigest.String() != evidence.proposal.Candidate.InputDigest ||
		outputDigest != evidence.adoption.Evidence.BeforeOutputDigest ||
		normalizedDigest.String() != evidence.adoption.Evidence.BeforeSemanticDigest {
		fmt.Fprintln(stderr, "gooo: retention certification: compiler result contradicts the authorized adoption evidence")
		return exitFailure
	}

	bindings := retentionBindings(inputs, evidence, compilerDigest, verifierDigest)
	certificate := generation.SemanticRetentionCertificate{
		Schema:                generation.SemanticRetentionCertificateSchema,
		AdoptionReportDigest:  bindings.AdoptionReportDigest,
		ObservationDigest:     bindings.ObservationDigest,
		ProposalDigest:        bindings.ProposalDigest,
		AuthorizationDigest:   bindings.AuthorizationDigest,
		CandidateStableID:     bindings.CandidateStableID,
		Target:                generation.SemanticRetentionTarget,
		Mode:                  generation.SemanticRetentionMode,
		ContractSourceDigest:  bindings.ContractSourceDigest,
		InputSourceDigest:     bindings.InputSourceDigest,
		NormalizedIRDigest:    normalizedDigest.String(),
		GeneratedOutputDigest: outputDigest,
		CompilerDigest:        bindings.CompilerDigest,
		ToolchainDigest:       bindings.ToolchainDigest,
		VerifierDigest:        bindings.VerifierDigest,
		PolicyDigest:          bindings.PolicyDigest,
		GeneratedSource:       append([]byte(nil), generated.result.Source...),
		RepositoryWrites:      0,
		LocalTestExecutions:   0,
	}
	certificate.CertificateID, err = generation.SemanticRetentionCertificateContentDigest(certificate)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: certificate identity: %v\n", err)
		return exitFailure
	}
	if err := generation.ValidateSemanticRetentionCertificate(certificate); err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: certificate validation: %v\n", err)
		return exitFailure
	}
	certificateData, err := json.MarshalIndent(certificate, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: encode certificate: %v\n", err)
		return exitFailure
	}
	certificateData = append(certificateData, '\n')
	certificateDigest := cache.HashBytes(certificateData).String()
	before := retentionResultBase(inputs, evidence, compilerDigest, verifierDigest)
	before.Lifecycle = "CERTIFY_BEFORE"
	before.Decision = "CLOSED"
	before.Reason = generation.SemanticRetentionCertifiedReason
	before.CertificateDigest = certificateDigest
	before.NormalizedIRDigest = normalizedDigest.String()
	before.GeneratedOutputDigest = outputDigest
	before.GeneratedSource = append([]byte(nil), generated.result.Source...)
	before.Metrics = retentionRuntimeMetrics(started, beforeMem, afterMem)
	before.Metrics.SemanticOperationCount = 1
	before.Metrics.CertificateMisses = 1
	before.Metrics.GeneratedBytesEqual = true
	before.Metrics.NormalizedSemanticEqual = true
	if err := generation.ValidateSemanticRetentionResult(before); err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: result validation: %v\n", err)
		return exitFailure
	}
	beforeData, err := json.MarshalIndent(before, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: encode before result: %v\n", err)
		return exitFailure
	}
	beforeData = append(beforeData, '\n')
	if err := writeRetentionArtifacts(options.outputDir, []retentionArtifact{
		{name: "retained-knowledge-certificate.json", data: certificateData},
		{name: "retention-before.json", data: beforeData},
	}); err != nil {
		fmt.Fprintf(stderr, "gooo: retention certification: output: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "retention certificate: %s\nretention before: %s\n", options.outputDir+"/retained-knowledge-certificate.json", options.outputDir+"/retention-before.json")
	return exitOK
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
