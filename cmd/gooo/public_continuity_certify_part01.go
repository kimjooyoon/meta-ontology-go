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
)

const continuityCertifyUsage = "usage: gooo certify-discovery CONTRACT.gooo --input FILE.gooo --candidate FILE --decision FILE --out DIRECTORY"

type continuityCertifyOptions struct {
	contractFilename string
	inputFilename    string
	candidate        string
	decision         string
	outputDir        string
}

func runContinuityCertify(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	options, err := parseContinuityCertifyArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	started := time.Now()
	contractSource, err := reader.ReadFile(options.contractFilename)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: read contract: %v\n", err)
		return exitFailure
	}
	policy, err := discoverypolicy.Load(options.contractFilename, contractSource)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: policy: %v\n", err)
		return exitFailure
	}
	if policy.SourceDigest != discoverypolicy.PolicySourceDigest() || policy.EvaluatorDigest != discoverypolicy.GeneratedEvaluatorDigest() {
		fmt.Fprintln(stderr, "gooo: discovery certification: policy is not bound to the generated evaluator")
		return exitFailure
	}
	candidateData, err := reader.ReadFile(options.candidate)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: read candidate: %v\n", err)
		return exitFailure
	}
	candidate, err := publiccontinuity.DecodeCandidate(candidateData)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: %v\n", err)
		return exitFailure
	}
	candidateDigest := cache.HashBytes(candidateData).String()
	decisionData, err := reader.ReadFile(options.decision)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: read decision: %v\n", err)
		return exitFailure
	}
	decision, err := decodeContinuityDecision(decisionData)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: %v\n", err)
		return exitFailure
	}
	decisionDigest := cache.HashBytes(decisionData).String()
	if decision.Decision != publiccontinuity.DecisionAccept {
		fmt.Fprintln(stderr, "gooo: discovery certification: explicit human rejection is terminal and cannot certify")
		return exitFailure
	}
	if err := publiccontinuity.ValidateBinding(decision.Binding, candidate, candidateDigest); err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: decision binding: %v\n", err)
		return exitFailure
	}
	if decision.Binding.ContractDigest != policy.SourceDigest || decision.Binding.EvaluatorDigest != policy.EvaluatorDigest {
		fmt.Fprintln(stderr, "gooo: discovery certification: decision policy binding mismatch")
		return exitFailure
	}
	input, code := readGenerateInput(generateOptions{filename: options.inputFilename}, reader, parser, false, stdout, stderr, time.Now().Add(commandDeadline))
	if code != exitOK {
		return code
	}
	if cache.HashBytes(input.source).String() != candidate.SourceDigest {
		fmt.Fprintln(stderr, "gooo: discovery certification: candidate source binding is stale")
		return exitFailure
	}
	compilerDigest, err := publiccontinuity.CompilerDigest(reader.ReadFile)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: compiler binding: %v\n", err)
		return exitFailure
	}
	verifierDigest, err := publiccontinuity.VerifierDigest(reader.ReadFile)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: verifier binding: %v\n", err)
		return exitFailure
	}
	generated, err := generateWithDeadlineCore(input.file, nil, commandDeadline)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: compiler execution: %v\n", err)
		return exitFailure
	}
	normalizedDigest, err := cache.SemanticDigest(generated.ir)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: semantic digest: %v\n", err)
		return exitFailure
	}
	manifest, err := buildProjectionManifest(options.inputFilename, generatedFileName, input.source, nil, generated.ir, generated.result)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: manifest: %v\n", err)
		return exitFailure
	}
	manifestData, err := jsonManifestBytes(manifest)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: encode manifest: %v\n", err)
		return exitFailure
	}
	manifestDigest := cache.HashBytes(manifestData).String()
	outputDigest := cache.HashBytes(generated.result.Source).String()
	if normalizedDigest.String() != candidate.InputSemanticDigest || normalizedDigest.String() != candidate.GeneratedSemanticDigest ||
		outputDigest != candidate.GeneratedOutputDigest || manifestDigest != candidate.GeneratedManifestDigest ||
		candidate.ToolchainDigest != generation.SemanticRetentionToolchainDigest() || candidate.ContractDigest != policy.SourceDigest ||
		candidate.EvaluatorDigest != policy.EvaluatorDigest {
		fmt.Fprintln(stderr, "gooo: discovery certification: compiler replay does not match the exact candidate")
		return exitFailure
	}
	certificate := publiccontinuity.Certificate{
		Schema: publiccontinuity.CertificateSchema, Mode: publiccontinuity.CertificateMode,
		ConversionSchema: publiccontinuity.ConversionSchema, SourceOperation: publiccontinuity.Operation,
		TargetOperation:       "gooo.generate.public-self-observation-consumption",
		DecisionReceiptDigest: decisionDigest,
		Binding:               publiccontinuity.BindingFromCandidate(candidate, candidateDigest),
		ContractSourceDigest:  cache.HashBytes(contractSource).String(), InputSourceDigest: cache.HashBytes(input.source).String(),
		CompilerDigest: compilerDigest, VerifierDigest: verifierDigest, PolicyDigest: policy.SourceDigest,
		EvaluatorDigest: policy.EvaluatorDigest, GeneratedSource: append([]byte(nil), generated.result.Source...),
		GeneratedManifest: append([]byte(nil), manifestData...), GeneratedManifestDigest: manifestDigest,
		ManualTransformations: 0, RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0,
	}
	certificate.CertificateID, err = publiccontinuity.CertificateContentDigest(certificate)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: certificate identity: %v\n", err)
		return exitFailure
	}
	if err := publiccontinuity.ValidateCertificate(certificate); err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: certificate validation: %v\n", err)
		return exitFailure
	}
	certificateData, err := marshalContinuityJSON(certificate)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: encode certificate: %v\n", err)
		return exitFailure
	}
	certificateDigest := cache.HashBytes(certificateData).String()
	report := publiccontinuity.Report{
		Schema: publiccontinuity.ReportSchema, Lifecycle: "CERTIFICATION", Decision: "CLOSED",
		Reason: "EXACT_ACCEPTED_CANDIDATE_CERTIFIED", CaseID: continuityCaseAccepted,
		Binding: certificate.Binding, DecisionReceiptDigest: decisionDigest, CertificateDigest: certificateDigest,
		PublicInvocations: 2, LedgerEntries: 2, Candidates: 1, DecisionReceipts: 1, AcceptedDecisions: 1,
		Certificates: 1, DigestContinuityEdgesExpected: 2, DigestContinuityEdgesObserved: 2,
		ManualTransformations: 0, SemanticOperationsBefore: 1, SemanticOperationsAfter: 1,
		CandidateCertificateByteReplayMismatches: 0, GeneratedBytesEqual: true, NormalizedSemanticEqual: true,
		ArtifactDenominator: 3, ArtifactCount: 3, RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0,
		WallMS: continuityWallMS(started), PeakRSSKib: readPeakRSSKib(),
	}
	reportData, err := marshalContinuityJSON(report)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: discovery certification: encode report: %v\n", err)
		return exitFailure
	}
	human := []byte(renderContinuityReport(report, "The certificate is a typed conversion from a public-discovery candidate; no manual artifact transformation occurred."))
	if err := writeContinuityArtifacts(options.outputDir, []continuityArtifact{
		{name: "continuity-certificate.json", data: certificateData},
		{name: "continuity-certification-report.json", data: reportData},
		{name: "continuity-certification-report.md", data: human},
	}); err != nil {
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
