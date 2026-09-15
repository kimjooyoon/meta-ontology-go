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
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/discoverypolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publiccontinuity"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicdiscovery"
)

const continuityCertifyUsage = "usage: gooo certify-discovery CONTRACT.gooo --input FILE.gooo --candidate FILE --decision FILE --out DIRECTORY"

type continuityCertifyOptions struct {
	contractFilename string
	inputFilename    string
	candidate        string
	decision         string
	outputDir        string
}

type continuityCertificationContext struct {
	contractSource  []byte
	policy          discoverypolicy.Policy
	candidate       publicdiscovery.Candidate
	candidateDigest string
	decision        publiccontinuity.DecisionReceipt
	decisionDigest  string
	input           generateInput
	compilerDigest  string
	verifierDigest  string
}

func runContinuityCertify(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	options, err := parseContinuityCertifyArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	started := time.Now()
	context, code := loadContinuityCertificationContext(options, reader, parser, stdout, stderr)
	if code != exitOK {
		return code
	}
	certificate, err := certifyContinuityCandidate(options, context)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	return writeContinuityCertificateResult(options, context, certificate, started, stdout, stderr)
}

func loadContinuityCertificationContext(options continuityCertifyOptions, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) (continuityCertificationContext, int) {
	contractSource, policy, err := loadContinuityPolicy(options.contractFilename, reader)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: policy: %v\n", err)
		return continuityCertificationContext{}, exitFailure
	}
	candidate, candidateDigest, err := loadContinuityCandidate(options.candidate, reader)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: candidate: %v\n", err)
		return continuityCertificationContext{}, exitFailure
	}
	decision, decisionData, err := loadContinuityDecision(options.decision, reader)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: decision: %v\n", err)
		return continuityCertificationContext{}, exitFailure
	}
	if err := validateContinuityDecision(policy, candidate, candidateDigest, decision); err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: %v\n", err)
		return continuityCertificationContext{}, exitFailure
	}
	input, code := readGenerateInput(generateOptions{filename: options.inputFilename}, reader, parser, false, stdout, stderr, time.Now().Add(commandDeadline))
	if code != exitOK {
		return continuityCertificationContext{}, code
	}
	if cache.HashBytes(input.source).String() != candidate.SourceDigest {
		fmt.Fprintln(stderr, "gooo: discovery certification: candidate source binding is stale")
		return continuityCertificationContext{}, exitFailure
	}
	compilerDigest, err := publiccontinuity.CompilerDigest(reader.ReadFile)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: compiler binding: %v\n", err)
		return continuityCertificationContext{}, exitFailure
	}
	verifierDigest, err := publiccontinuity.VerifierDigest(reader.ReadFile)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: verifier binding: %v\n", err)
		return continuityCertificationContext{}, exitFailure
	}
	return continuityCertificationContext{contractSource: contractSource, policy: policy, candidate: candidate, candidateDigest: candidateDigest,
		decision: decision, decisionDigest: cache.HashBytes(decisionData).String(), input: input, compilerDigest: compilerDigest, verifierDigest: verifierDigest}, exitOK
}

func loadContinuityPolicy(filename string, reader SourceReader) ([]byte, discoverypolicy.Policy, error) {
	source, err := reader.ReadFile(filename)
	if err != nil {
		return nil, discoverypolicy.Policy{}, fmt.Errorf("read contract: %w", err)
	}
	policy, err := discoverypolicy.Load(filename, source)
	if err != nil {
		return nil, discoverypolicy.Policy{}, err
	}
	if policy.SourceDigest != discoverypolicy.PolicySourceDigest() || policy.EvaluatorDigest != discoverypolicy.GeneratedEvaluatorDigest() {
		return nil, discoverypolicy.Policy{}, fmt.Errorf("policy is not bound to the generated evaluator")
	}
	return source, policy, nil
}

func loadContinuityCandidate(filename string, reader SourceReader) (publicdiscovery.Candidate, string, error) {
	data, err := reader.ReadFile(filename)
	if err != nil {
		return publicdiscovery.Candidate{}, "", fmt.Errorf("read candidate: %w", err)
	}
	candidate, err := publiccontinuity.DecodeCandidate(data)
	if err != nil {
		return publicdiscovery.Candidate{}, "", err
	}
	return candidate, cache.HashBytes(data).String(), nil
}

func loadContinuityDecision(filename string, reader SourceReader) (publiccontinuity.DecisionReceipt, []byte, error) {
	data, err := reader.ReadFile(filename)
	if err != nil {
		return publiccontinuity.DecisionReceipt{}, nil, fmt.Errorf("read decision: %w", err)
	}
	decision, err := decodeContinuityDecision(data)
	return decision, data, err
}

func validateContinuityDecision(policy discoverypolicy.Policy, candidate publicdiscovery.Candidate, candidateDigest string, decision publiccontinuity.DecisionReceipt) error {
	if decision.Decision != publiccontinuity.DecisionAccept {
		return fmt.Errorf("explicit human rejection is terminal and cannot certify")
	}
	if err := publiccontinuity.ValidateBinding(decision.Binding, candidate, candidateDigest); err != nil {
		return fmt.Errorf("decision binding: %w", err)
	}
	if decision.Binding.ContractDigest != policy.SourceDigest || decision.Binding.EvaluatorDigest != policy.EvaluatorDigest {
		return fmt.Errorf("decision policy binding mismatch")
	}
	return nil
}

func certifyContinuityCandidate(options continuityCertifyOptions, context continuityCertificationContext) (publiccontinuity.Certificate, error) {
	source, manifest, normalizedDigest, outputDigest, manifestDigest, err := replayContinuityCandidate(options, context)
	if err != nil {
		return publiccontinuity.Certificate{}, err
	}
	candidate := context.candidate
	if normalizedDigest != candidate.InputSemanticDigest || normalizedDigest != candidate.GeneratedSemanticDigest || outputDigest != candidate.GeneratedOutputDigest ||
		manifestDigest != candidate.GeneratedManifestDigest || candidate.ToolchainDigest != generation.SemanticRetentionToolchainDigest() ||
		candidate.ContractDigest != context.policy.SourceDigest || candidate.EvaluatorDigest != context.policy.EvaluatorDigest {
		return publiccontinuity.Certificate{}, fmt.Errorf("gooo: discovery certification: compiler replay does not match the exact candidate")
	}
	certificate := publiccontinuity.Certificate{Schema: publiccontinuity.CertificateSchema, Mode: publiccontinuity.CertificateMode,
		ConversionSchema: publiccontinuity.ConversionSchema, SourceOperation: publiccontinuity.Operation,
		TargetOperation: "gooo.generate.public-self-observation-consumption", DecisionReceiptDigest: context.decisionDigest,
		Binding: publiccontinuity.BindingFromCandidate(candidate, context.candidateDigest), ContractSourceDigest: cache.HashBytes(context.contractSource).String(),
		InputSourceDigest: cache.HashBytes(context.input.source).String(), CompilerDigest: context.compilerDigest, VerifierDigest: context.verifierDigest,
		PolicyDigest: context.policy.SourceDigest, EvaluatorDigest: context.policy.EvaluatorDigest, GeneratedSource: source,
		GeneratedManifest: manifest, GeneratedManifestDigest: manifestDigest, ManualTransformations: 0, RepositoryWrites: 0,
		LocalBuildExecutions: 0, LocalTestExecutions: 0}
	certificate.CertificateID, err = publiccontinuity.CertificateContentDigest(certificate)
	if err != nil {
		return publiccontinuity.Certificate{}, fmt.Errorf("gooo: discovery certification: certificate identity: %w", err)
	}
	if err := publiccontinuity.ValidateCertificate(certificate); err != nil {
		return publiccontinuity.Certificate{}, fmt.Errorf("gooo: discovery certification: certificate validation: %w", err)
	}
	return certificate, nil
}

func replayContinuityCandidate(options continuityCertifyOptions, context continuityCertificationContext) ([]byte, []byte, string, string, string, error) {
	generated, err := generateWithDeadlineCore(context.input.file, nil, commandDeadline)
	if err != nil {
		return nil, nil, "", "", "", fmt.Errorf("gooo: discovery certification: compiler execution: %w", err)
	}
	normalizedDigest, err := cache.SemanticDigest(generated.ir)
	if err != nil {
		return nil, nil, "", "", "", fmt.Errorf("gooo: discovery certification: semantic digest: %w", err)
	}
	manifest, err := buildProjectionManifest(options.inputFilename, generatedFileName, context.input.source, nil, generated.ir, generated.result)
	if err != nil {
		return nil, nil, "", "", "", fmt.Errorf("gooo: discovery certification: manifest: %w", err)
	}
	manifestData, err := jsonManifestBytes(manifest)
	if err != nil {
		return nil, nil, "", "", "", fmt.Errorf("gooo: discovery certification: encode manifest: %w", err)
	}
	return append([]byte(nil), generated.result.Source...), append([]byte(nil), manifestData...), normalizedDigest.String(),
		cache.HashBytes(generated.result.Source).String(), cache.HashBytes(manifestData).String(), nil
}

func writeContinuityCertificateResult(options continuityCertifyOptions, context continuityCertificationContext, certificate publiccontinuity.Certificate, started time.Time, stdout, stderr io.Writer) int {
	certificateData, err := marshalContinuityJSON(certificate)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: encode certificate: %v\n", err)
		return exitFailure
	}
	certificateDigest := cache.HashBytes(certificateData).String()
	report := publiccontinuity.Report{Schema: publiccontinuity.ReportSchema, Lifecycle: "CERTIFICATION", Decision: "CLOSED",
		Reason: "EXACT_ACCEPTED_CANDIDATE_CERTIFIED", CaseID: continuityCaseAccepted, Binding: certificate.Binding,
		DecisionReceiptDigest: context.decisionDigest, CertificateDigest: certificateDigest, PublicInvocations: 2, LedgerEntries: 2,
		Candidates: 1, DecisionReceipts: 1, AcceptedDecisions: 1, Certificates: 1, DigestContinuityEdgesExpected: 2,
		DigestContinuityEdgesObserved: 2, SemanticOperationsBefore: 1, SemanticOperationsAfter: 1,
		GeneratedBytesEqual: true, NormalizedSemanticEqual: true, ArtifactDenominator: 3, ArtifactCount: 3,
		WallMS: continuityWallMS(started), PeakRSSKib: readPeakRSSKib()}
	reportData, err := marshalContinuityJSON(report)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: encode report: %v\n", err)
		return exitFailure
	}
	human := []byte(renderContinuityReport(report, "The certificate is a typed conversion from a public-discovery candidate; no manual artifact transformation occurred."))
	if err := writeContinuityArtifacts(options.outputDir, []continuityArtifact{{name: "continuity-certificate.json", data: certificateData},
		{name: "continuity-certification-report.json", data: reportData}, {name: "continuity-certification-report.md", data: human}}); err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: output: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "continuity certificate: %s\nreport: %s\n", filepath.Join(options.outputDir, "continuity-certificate.json"), filepath.Join(options.outputDir, "continuity-certification-report.json"))
	return exitOK
}

const continuityCaseAccepted = "COMPLETE_ACCEPTED_CHAIN"

func decodeContinuityDecision(data []byte) (publiccontinuity.DecisionReceipt, error) {
	var decision publiccontinuity.DecisionReceipt
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decision); err != nil {
		return decision, err
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return decision, errors.New("decision receipt contains multiple JSON values")
	} else if err != io.EOF {
		return decision, fmt.Errorf("decode decision receipt trailer: %w", err)
	}
	if err := publiccontinuity.ValidateDecisionReceipt(decision); err != nil {
		return decision, err
	}
	return decision, nil
}

func parseContinuityCertifyArguments(args []string) (continuityCertifyOptions, error) {
	if len(args) == 0 {
		return continuityCertifyOptions{}, fmt.Errorf("%s", continuityCertifyUsage)
	}
	options := continuityCertifyOptions{contractFilename: args[0]}
	seen := map[string]bool{}
	for index := 1; index < len(args); index++ {
		if index+1 >= len(args) || args[index+1] == "" || seen[args[index]] {
			return continuityCertifyOptions{}, fmt.Errorf("%s", continuityCertifyUsage)
		}
		value := args[index+1]
		seen[args[index]] = true
		switch args[index] {
		case "--input":
			options.inputFilename = value
		case "--candidate":
			options.candidate = value
		case "--decision":
			options.decision = value
		case "--out":
			options.outputDir = value
		default:
			return continuityCertifyOptions{}, fmt.Errorf("%s", continuityCertifyUsage)
		}
		index++
	}
	if options.contractFilename == "" || options.inputFilename == "" || options.candidate == "" || options.decision == "" || options.outputDir == "" {
		return continuityCertifyOptions{}, fmt.Errorf("%s", continuityCertifyUsage)
	}
	return options, nil
}
