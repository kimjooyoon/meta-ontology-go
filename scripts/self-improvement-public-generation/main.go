package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/retentionpolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type publicCase struct {
	ID        string `json:"id"`
	Decision  string `json:"decision"`
	Reason    string `json:"reason"`
	Artifacts int    `json:"artifacts"`
}

type publicLoopReport struct {
	Schema                  string                                     `json:"schema"`
	CaseDenominator         int                                        `json:"case_denominator"`
	Counts                  map[string]int                             `json:"counts"`
	PublicArtifactsExpected int                                        `json:"public_artifacts_expected"`
	PublicArtifacts         int                                        `json:"public_artifacts"`
	BaselineArtifacts       int                                        `json:"baseline_artifacts"`
	GeneratedBytesEqual     bool                                       `json:"generated_bytes_equal"`
	NormalizedSemanticEqual bool                                       `json:"normalized_semantic_equal"`
	BaselineMetrics         generation.SemanticRetentionRuntimeMetrics `json:"baseline_metrics"`
	AppliedMetrics          generation.SemanticRetentionRuntimeMetrics `json:"applied_metrics"`
	ReplayMetrics           generation.SemanticRetentionRuntimeMetrics `json:"replay_metrics"`
	Cases                   []publicCase                               `json:"cases"`
	RepositoryWrites        int                                        `json:"repository_writes"`
	LocalTestExecutions     int                                        `json:"local_test_executions"`
}

type publicRunInputs struct {
	input           []byte
	certificateData []byte
	certificate     generation.SemanticRetentionCertificate
	contract        []byte
	evidence        publicEvidence
}

func main() {
	contract := flag.String("contract", "", "policy .gooo")
	input := flag.String("input", "", "input .gooo")
	certificate := flag.String("certificate", "", "retained certificate")
	observation := flag.String("observation", "", "compiler observation")
	proposal := flag.String("proposal", "", "adoption proposal")
	authorization := flag.String("authorization", "", "authorization")
	adoption := flag.String("adoption", "", "adoption report")
	baseline := flag.String("baseline", "", "baseline output directory")
	applied := flag.String("applied", "", "certificate output directory")
	replay := flag.String("replay", "", "replay output directory")
	cases := flag.String("cases", "", "six public case output directories")
	output := flag.String("output", "", "report output")
	flag.Parse()
	if *contract == "" || *input == "" || *certificate == "" || *observation == "" || *proposal == "" || *authorization == "" || *adoption == "" || *baseline == "" || *applied == "" || *replay == "" || *cases == "" || *output == "" {
		fail(errors.New("self-improvement-public-generation requires contract, input, certificate, baseline, applied, replay, cases, and output"))
	}
	if err := run(*contract, *input, *certificate, *observation, *proposal, *authorization, *adoption, *baseline, *applied, *replay, *cases, *output); err != nil {
		fail(err)
	}
}

func run(contractPath, inputPath, certificatePath, observationPath, proposalPath, authorizationPath, adoptionPath, baselineDir, appliedDir, replayDir, casesRoot, outputPath string) error {
	inputs, err := loadPublicRunInputs(contractPath, inputPath, certificatePath, observationPath, proposalPath, authorizationPath, adoptionPath)
	if err != nil {
		return err
	}
	if err := verifyPublicPolicy(inputs.contract); err != nil {
		return err
	}
	if err := verifyPublicEvidence(inputs.contract, inputs.input, inputs.certificate, inputs.evidence); err != nil {
		return err
	}
	inputDigest, err := independentlyComputeInputDigest(inputPath, inputs.input)
	if err != nil {
		return err
	}
	return verifyPublicLoop(inputs, inputDigest, baselineDir, appliedDir, replayDir, casesRoot, outputPath)
}

func loadPublicRunInputs(contractPath, inputPath, certificatePath, observationPath, proposalPath, authorizationPath, adoptionPath string) (publicRunInputs, error) {
	var inputs publicRunInputs
	var err error
	if inputs.input, err = os.ReadFile(inputPath); err != nil {
		return inputs, fmt.Errorf("read input: %w", err)
	}
	if inputs.certificateData, inputs.certificate, err = readCertificate(certificatePath); err != nil {
		return inputs, err
	}
	if err := generation.ValidateSemanticRetentionCertificate(inputs.certificate); err != nil {
		return inputs, fmt.Errorf("certificate: %w", err)
	}
	if inputs.contract, err = os.ReadFile(contractPath); err != nil {
		return inputs, fmt.Errorf("read contract: %w", err)
	}
	if inputs.evidence, err = readPublicEvidence(observationPath, proposalPath, authorizationPath, adoptionPath); err != nil {
		return inputs, err
	}
	return inputs, nil
}

func verifyPublicLoop(inputs publicRunInputs, normalizedDigest, baselineDir, appliedDir, replayDir, casesRoot, outputPath string) error {
	certificateDigest := cache.HashBytes(inputs.certificateData).String()
	sourceDigest := cache.HashBytes(inputs.input).String()
	baseline, err := readReport(baselineDir)
	if err != nil {
		return fmt.Errorf("baseline: %w", err)
	}
	applied, err := readReport(appliedDir)
	if err != nil {
		return fmt.Errorf("applied: %w", err)
	}
	replay, err := readReport(replayDir)
	if err != nil {
		return fmt.Errorf("replay: %w", err)
	}
	if err := verifyClosedPublicResults(sourceDigest, normalizedDigest, inputs.certificate, certificateDigest, baseline, applied, replay); err != nil {
		return err
	}
	caseReports, counts, artifacts, err := verifyPublicCases(casesRoot, inputs.contract)
	if err != nil {
		return err
	}
	if artifacts != 16 || counts["CLOSED"] != 2 || counts["UNKNOWN"] != 2 || counts["REFUTED"] != 2 {
		return errors.New("public generation denominator or outcome counts changed")
	}
	report := publicLoopReport{
		Schema: "gooo/semantic-public-generation-loop/v1", CaseDenominator: generation.SemanticRetentionCaseDenominator,
		Counts: counts, PublicArtifactsExpected: 16, PublicArtifacts: artifacts,
		BaselineArtifacts: directoryFileCount(baselineDir), GeneratedBytesEqual: true, NormalizedSemanticEqual: true,
		BaselineMetrics: baseline.Metrics, AppliedMetrics: applied.Metrics, ReplayMetrics: replay.Metrics,
		Cases: caseReports, RepositoryWrites: 0, LocalTestExecutions: 0,
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode report: %w", err)
	}
	if err := os.WriteFile(outputPath, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

type publicEvidence struct {
	observationData   []byte
	observation       generation.SemanticObservation
	proposalData      []byte
	proposal          generation.SemanticAdoptionProposal
	authorizationData []byte
	authorization     generation.SemanticAdoptionAuthorization
	adoptionData      []byte
	adoption          generation.SemanticAdoptionReport
}

func readPublicEvidence(observationPath, proposalPath, authorizationPath, adoptionPath string) (publicEvidence, error) {
	var evidence publicEvidence
	var err error
	if evidence.observationData, err = os.ReadFile(observationPath); err != nil {
		return evidence, fmt.Errorf("read observation: %w", err)
	}
	if err := json.Unmarshal(evidence.observationData, &evidence.observation); err != nil {
		return evidence, fmt.Errorf("decode observation: %w", err)
	}
	if evidence.proposalData, err = os.ReadFile(proposalPath); err != nil {
		return evidence, fmt.Errorf("read proposal: %w", err)
	}
	if err := json.Unmarshal(evidence.proposalData, &evidence.proposal); err != nil {
		return evidence, fmt.Errorf("decode proposal: %w", err)
	}
	if evidence.authorizationData, err = os.ReadFile(authorizationPath); err != nil {
		return evidence, fmt.Errorf("read authorization: %w", err)
	}
	if err := json.Unmarshal(evidence.authorizationData, &evidence.authorization); err != nil {
		return evidence, fmt.Errorf("decode authorization: %w", err)
	}
	if evidence.adoptionData, err = os.ReadFile(adoptionPath); err != nil {
		return evidence, fmt.Errorf("read adoption: %w", err)
	}
	if err := json.Unmarshal(evidence.adoptionData, &evidence.adoption); err != nil {
		return evidence, fmt.Errorf("decode adoption: %w", err)
	}
	return evidence, nil
}

func verifyPublicPolicy(contract []byte) error {
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("contract.gooo", string(contract), syntax.EntityFieldsV1Support())
	if diagnostics.HasErrors() {
		return errors.New("public retention policy parsing failed")
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, syntax.EntityFieldsV1Support())
	if err != nil {
		return fmt.Errorf("public retention policy lowering failed: %w", err)
	}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.Name != "PublishOperationReceipt" {
			continue
		}
		for _, marker := range []string{"public-generate=baseline-or-retained", "certificate=explicit-only", "invalid=fail-closed", "fallback=none"} {
			if !strings.Contains(node.ValueProgram, marker) {
				return fmt.Errorf("public retention policy omits %s", marker)
			}
		}
		return nil
	}
	return errors.New("public retention policy activity is missing")
}

func verifyPublicEvidence(contract, input []byte, certificate generation.SemanticRetentionCertificate, evidence publicEvidence) error {
	if err := generation.VerifySemanticObservation(evidence.observation); err != nil {
		return fmt.Errorf("observation: %w", err)
	}
	if evidence.observation.ContractDigest != cache.HashBytes(contract).String() || evidence.observation.InputSourceDigest != cache.HashBytes(input).String() || len(evidence.observation.Candidates) != 1 {
		return errors.New("observation is not bound to the exact public compiler input")
	}
	if err := generation.ValidateSemanticAdoptionProposal(evidence.proposal); err != nil {
		return fmt.Errorf("proposal: %w", err)
	}
	observationDigest := cache.HashBytes(evidence.observationData).String()
	proposalDigest := cache.HashBytes(evidence.proposalData).String()
	authorizationDigest := cache.HashBytes(evidence.authorizationData).String()
	if evidence.proposal.ObservationDigest != observationDigest || evidence.proposal.ContractDigest != evidence.observation.ContractDigest || evidence.proposal.InputSourceDigest != evidence.observation.InputSourceDigest || !reflect.DeepEqual(evidence.proposal.Candidate, evidence.observation.Candidates[0]) {
		return errors.New("proposal is not bound to exact public observation")
	}
	if err := generation.ValidateSemanticAdoptionAuthorization(evidence.authorization); err != nil {
		return fmt.Errorf("authorization: %w", err)
	}
	if !evidence.authorization.Authorized || evidence.authorization.ProposalDigest != proposalDigest || evidence.authorization.CandidateStableID != evidence.proposal.Candidate.StableID || evidence.authorization.CandidateInputDigest != evidence.proposal.Candidate.InputDigest || evidence.authorization.ContractDigest != evidence.proposal.ContractDigest || evidence.authorization.InputSourceDigest != evidence.proposal.InputSourceDigest {
		return errors.New("authorization is not the exact authorized public candidate")
	}
	if evidence.adoption.Schema != generation.SemanticAdoptionReportSchema || evidence.adoption.Lifecycle != "AUTHORIZED_ADOPTION" || evidence.adoption.ObservationDigest != observationDigest || evidence.adoption.ProposalDigest != proposalDigest || evidence.adoption.AuthorizationDigest != authorizationDigest || !reflect.DeepEqual(evidence.adoption.Proposal, evidence.proposal) || !reflect.DeepEqual(evidence.adoption.Authorization, evidence.authorization) || evidence.adoption.IndependentDecision != "CLOSED" {
		return errors.New("adoption report is not bound to exact public evidence")
	}
	compilerDigest, err := generation.SemanticRetentionCompilerDigest(os.ReadFile)
	if err != nil {
		return fmt.Errorf("compiler binding: %w", err)
	}
	verifierDigest, err := generation.SemanticRetentionVerifierDigest(os.ReadFile)
	if err != nil {
		return fmt.Errorf("verifier binding: %w", err)
	}
	expected := generation.SemanticRetentionBindings{
		AdoptionReportDigest: cache.HashBytes(evidence.adoptionData).String(), ObservationDigest: observationDigest,
		ProposalDigest: proposalDigest, AuthorizationDigest: authorizationDigest, CandidateStableID: evidence.proposal.Candidate.StableID,
		ContractSourceDigest: cache.HashBytes(contract).String(), InputSourceDigest: cache.HashBytes(input).String(), NormalizedIRDigest: evidence.proposal.Candidate.InputDigest,
		CompilerDigest: compilerDigest, ToolchainDigest: generation.SemanticRetentionToolchainDigest(), VerifierDigest: verifierDigest, PolicyDigest: cache.HashBytes(contract).String(), EvaluatorDigest: retentionpolicy.GeneratedEvaluatorDigest(),
	}
	if err := generation.VerifySemanticRetentionCertificate(certificate, expected); err != nil {
		return fmt.Errorf("certificate bindings: %w", err)
	}
	if evidence.adoption.IndependentReason != evidence.adoption.Evidence.Reason {
		return errors.New("adoption independent reason is not bound to evidence")
	}
	if decision, reason, _, err := generation.VerifySemanticAdoption(evidence.proposal, proposalDigest, evidence.authorization, authorizationDigest, evidence.adoption.Evidence); err != nil || decision != "CLOSED" || reason != generation.SemanticAdoptionClosedReason {
		return fmt.Errorf("adoption verification = %s/%s: %v", decision, reason, err)
	}
	if err := generation.ValidateBoundSemanticAdoption(evidence.adoption.Observation, evidence.proposal, evidence.adoption.Evidence); err != nil {
		return err
	}
	if err := generation.VerifySemanticObservation(evidence.adoption.Observation); err != nil {
		return fmt.Errorf("adopted observation: %w", err)
	}
	return nil
}

func verifyClosedPublicResults(sourceDigest, normalizedDigest string, certificate generation.SemanticRetentionCertificate, certificateDigest string, baseline, applied, replay generation.SemanticPublicGenerationReport) error {
	if err := validateClosedPublicReports(baseline, applied, replay); err != nil {
		return err
	}
	if err := verifyClosedPublicBindings(sourceDigest, certificate, certificateDigest, baseline, applied, replay); err != nil {
		return err
	}
	if err := verifyClosedPublicOutputs(normalizedDigest, certificate, baseline, applied, replay); err != nil {
		return err
	}
	return verifyClosedPublicAccounting(baseline, applied, replay)
}

func validateClosedPublicReports(baseline, applied, replay generation.SemanticPublicGenerationReport) error {
	for name, report := range map[string]generation.SemanticPublicGenerationReport{"baseline": baseline, "applied": applied, "replay": replay} {
		if err := generation.ValidateSemanticPublicGenerationReport(report); err != nil {
			return fmt.Errorf("%s report: %w", name, err)
		}
	}
	if baseline.Reason != generation.SemanticPublicGenerationBaselineReason || applied.Reason != generation.SemanticPublicGenerationHitReason || replay.Reason != generation.SemanticPublicGenerationHitReason {
		return errors.New("public generation reports do not identify baseline and retained paths")
	}
	return nil
}

func verifyClosedPublicBindings(sourceDigest string, certificate generation.SemanticRetentionCertificate, certificateDigest string, baseline, applied, replay generation.SemanticPublicGenerationReport) error {
	if baseline.InputSourceDigest != sourceDigest || applied.InputSourceDigest != sourceDigest || replay.InputSourceDigest != sourceDigest {
		return errors.New("public generation reports are not bound to the exact input")
	}
	if applied.CertificateDigest != certificateDigest || replay.CertificateDigest != certificateDigest || applied.GeneratedOutputDigest != certificate.GeneratedOutputDigest || replay.GeneratedOutputDigest != certificate.GeneratedOutputDigest {
		return errors.New("public generation reports are not bound to the certificate")
	}
	for _, report := range []generation.SemanticPublicGenerationReport{applied, replay} {
		if report.AdoptionReportDigest != certificate.AdoptionReportDigest || report.ObservationDigest != certificate.ObservationDigest || report.ProposalDigest != certificate.ProposalDigest || report.AuthorizationDigest != certificate.AuthorizationDigest || report.CandidateStableID != certificate.CandidateStableID || report.ContractSourceDigest != certificate.ContractSourceDigest || report.CompilerDigest != certificate.CompilerDigest || report.ToolchainDigest != certificate.ToolchainDigest || report.VerifierDigest != certificate.VerifierDigest || report.PolicyDigest != certificate.PolicyDigest || report.EvaluatorDigest != certificate.EvaluatorDigest || report.GeneratedManifestDigest != certificate.GeneratedManifestDigest {
			return errors.New("public generation report dropped an evidence binding")
		}
	}
	return nil
}

func verifyClosedPublicOutputs(normalizedDigest string, certificate generation.SemanticRetentionCertificate, baseline, applied, replay generation.SemanticPublicGenerationReport) error {
	baselineSource, appliedSource, replaySource, err := readClosedPublicSources(baseline, applied, replay)
	if err != nil {
		return err
	}
	if !bytes.Equal(baselineSource, certificate.GeneratedSource) || !bytes.Equal(baselineSource, appliedSource) || !bytes.Equal(appliedSource, replaySource) {
		return errors.New("public compiler output bytes changed across baseline, certificate, and replay")
	}
	if baseline.NormalizedIRDigest != normalizedDigest || baseline.NormalizedIRDigest != applied.NormalizedIRDigest || applied.NormalizedIRDigest != replay.NormalizedIRDigest || applied.NormalizedIRDigest != certificate.NormalizedIRDigest {
		return errors.New("public compiler normalized semantic identity changed across invocations")
	}
	if baseline.Metrics.SemanticOperationCount != 1 || applied.Metrics.SemanticOperationCount != 0 || replay.Metrics.SemanticOperationCount != 0 || applied.Metrics.CertificateHits != 1 || replay.Metrics.CertificateHits != 1 {
		return errors.New("public compiler operation or certificate counts changed")
	}
	manifest, err := os.ReadFile(applied.ManifestFile)
	if err != nil {
		return fmt.Errorf("read applied manifest: %w", err)
	}
	if !bytes.Equal(manifest, certificate.GeneratedManifest) {
		return errors.New("public compiler manifest bytes changed on certificate application")
	}
	return nil
}

func readClosedPublicSources(baseline, applied, replay generation.SemanticPublicGenerationReport) ([]byte, []byte, []byte, error) {
	baselineSource, err := os.ReadFile(baseline.OutputFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read baseline source: %w", err)
	}
	appliedSource, err := os.ReadFile(applied.OutputFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read applied source: %w", err)
	}
	replaySource, err := os.ReadFile(replay.OutputFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read replay source: %w", err)
	}
	return baselineSource, appliedSource, replaySource, nil
}

func verifyClosedPublicAccounting(baseline, applied, replay generation.SemanticPublicGenerationReport) error {
	if directoryFileCount(filepath.Dir(baseline.OutputFile)) != 4 || directoryFileCount(filepath.Dir(applied.OutputFile)) != 4 || directoryFileCount(filepath.Dir(replay.OutputFile)) != 4 {
		return errors.New("public compiler artifact denominator changed")
	}
	for _, report := range []generation.SemanticPublicGenerationReport{baseline, applied, replay} {
		if report.OutputBytes != directoryFileBytes(filepath.Dir(report.OutputFile)) {
			return errors.New("public compiler output byte count is not exact")
		}
	}
	return nil
}

func verifyPublicCases(root string, contract []byte) ([]publicCase, map[string]int, int, error) {
	ids, decisions, err := independentPublicDecisionRows(contract)
	if err != nil {
		return nil, nil, 0, err
	}
	directories := map[string]string{
		"CERTIFICATE_HIT": "certificate-hit", "CERTIFICATE_REPLAY": "certificate-replay",
		"MISSING_CERTIFICATE": "missing-certificate", "MISSING_AUTHORIZATION": "missing-authorization",
		"STALE_INPUT": "stale-input", "MISMATCHED_CERTIFICATE": "mismatched-certificate",
	}
	counts := map[string]int{"CLOSED": 0, "UNKNOWN": 0, "REFUTED": 0}
	results := make([]publicCase, 0, len(ids))
	total := 0
	for _, id := range ids {
		report, err := readReport(filepath.Join(root, directories[id]))
		if err != nil {
			return nil, nil, 0, fmt.Errorf("case %s: %w", id, err)
		}
		expectedDecision := decisions[id]
		expectedReason := publicDecisionReason(id, expectedDecision)
		if report.Decision != expectedDecision || report.Reason != expectedReason {
			return nil, nil, 0, fmt.Errorf("case %s = %s/%s, want %s/%s", id, report.Decision, report.Reason, expectedDecision, expectedReason)
		}
		caseDirectory := filepath.Join(root, directories[id])
		artifacts := directoryFileCount(caseDirectory)
		if artifacts != expectedArtifacts(report.Decision) {
			return nil, nil, 0, fmt.Errorf("case %s artifacts = %d", id, artifacts)
		}
		if report.OutputFileCount != artifacts || report.OutputBytes != directoryFileBytes(caseDirectory) {
			return nil, nil, 0, fmt.Errorf("case %s output accounting is not exact", id)
		}
		if report.Decision != "CLOSED" {
			for _, name := range []string{"semantic.gooo.go", "semantic.gooo.manifest.jsonl"} {
				if _, err := os.Stat(filepath.Join(caseDirectory, name)); !os.IsNotExist(err) {
					return nil, nil, 0, fmt.Errorf("case %s wrote generated artifact %s", id, name)
				}
			}
		}
		counts[report.Decision]++
		total += artifacts
		results = append(results, publicCase{ID: id, Decision: report.Decision, Reason: report.Reason, Artifacts: artifacts})
	}
	return results, counts, total, nil
}

func independentPublicDecisionRows(contract []byte) ([]string, map[string]string, error) {
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("contract.gooo", string(contract), syntax.EntityFieldsV1Support())
	if diagnostics.HasErrors() {
		return nil, nil, errors.New("independent public decision policy parsing failed")
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, syntax.EntityFieldsV1Support())
	if err != nil {
		return nil, nil, fmt.Errorf("independent public decision policy lowering failed: %w", err)
	}
	decisions := make(map[string]string, generation.SemanticRetentionCaseDenominator)
	ids := make([]string, 0, generation.SemanticRetentionCaseDenominator)
	activityCount := 0
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.Name != retentionpolicy.PolicyActivity {
			continue
		}
		activityCount++
		for _, part := range strings.Split(node.ValueProgram, ";") {
			if !strings.HasPrefix(part, "retention-case=") {
				continue
			}
			id, decision, ok := strings.Cut(strings.TrimPrefix(part, "retention-case="), ":")
			if !ok || id == "" || decision == "" || decisions[id] != "" {
				return nil, nil, fmt.Errorf("independent public decision row %q is malformed or duplicated", part)
			}
			decisions[id] = decision
			ids = append(ids, id)
		}
	}
	expected := generation.SemanticRetentionCaseIDs()
	if activityCount != 1 || len(ids) != len(expected) {
		return nil, nil, fmt.Errorf("independent public policy activities/rows = %d/%d, want 1/%d", activityCount, len(ids), len(expected))
	}
	for index, id := range expected {
		if ids[index] != id || (decisions[id] != "CLOSED" && decisions[id] != "UNKNOWN" && decisions[id] != "REFUTED") {
			return nil, nil, fmt.Errorf("independent public decision row %d is not fixed-policy compliant", index+1)
		}
	}
	return ids, decisions, nil
}

func publicDecisionReason(caseID, decision string) string {
	switch decision {
	case "CLOSED":
		return generation.SemanticPublicGenerationHitReason
	case "UNKNOWN":
		if caseID == retentionpolicy.CaseMissingAuthorization {
			return generation.SemanticRetentionUnknownAuthorizationReason
		}
		return generation.SemanticRetentionUnknownCertificateReason
	case "REFUTED":
		return generation.SemanticRetentionRefutedReason
	default:
		return ""
	}
}

func expectedArtifacts(decision string) int {
	if decision == "CLOSED" {
		return generation.SemanticPublicGenerationArtifactsClosed
	}
	return generation.SemanticPublicGenerationArtifactsFailClosed
}

func readReport(directory string) (generation.SemanticPublicGenerationReport, error) {
	data, err := os.ReadFile(filepath.Join(directory, "generation-report.json"))
	if err != nil {
		return generation.SemanticPublicGenerationReport{}, err
	}
	var report generation.SemanticPublicGenerationReport
	if err := json.Unmarshal(data, &report); err != nil {
		return report, err
	}
	return report, nil
}

func readCertificate(path string) ([]byte, generation.SemanticRetentionCertificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, generation.SemanticRetentionCertificate{}, fmt.Errorf("read certificate: %w", err)
	}
	var certificate generation.SemanticRetentionCertificate
	if err := json.Unmarshal(data, &certificate); err != nil {
		return nil, certificate, fmt.Errorf("decode certificate: %w", err)
	}
	return data, certificate, nil
}

func independentlyComputeInputDigest(filename string, source []byte) (string, error) {
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(filename, string(source), syntax.EntityFieldsV1Support())
	if diagnostics.HasErrors() {
		return "", errors.New("independent input parsing failed")
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, syntax.EntityFieldsV1Support())
	if err != nil {
		return "", fmt.Errorf("independent semantic lowering failed: %w", err)
	}
	digest, err := cache.SemanticDigest(ir)
	if err != nil {
		return "", fmt.Errorf("independent normalized semantic digest failed: %w", err)
	}
	return digest.String(), nil
}

func directoryFileCount(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	return len(entries)
}

func directoryFileBytes(path string) int64 {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	var total int64
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err == nil {
			total += info.Size()
		}
	}
	return total
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
