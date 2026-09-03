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
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	certificateData, certificate, err := readCertificate(certificatePath)
	if err != nil {
		return err
	}
	if err := generation.ValidateSemanticRetentionCertificate(certificate); err != nil {
		return fmt.Errorf("certificate: %w", err)
	}
	contract, err := os.ReadFile(contractPath)
	if err != nil {
		return fmt.Errorf("read contract: %w", err)
	}
	evidence, err := readPublicEvidence(observationPath, proposalPath, authorizationPath, adoptionPath)
	if err != nil {
		return err
	}
	if err := verifyPublicPolicy(contract); err != nil {
		return err
	}
	if err := verifyPublicEvidence(contract, input, certificate, evidence); err != nil {
		return err
	}
	certificateDigest := cache.HashBytes(certificateData).String()
	inputDigest, err := independentlyComputeInputDigest(inputPath, input)
	if err != nil {
		return err
	}
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
	if err := verifyClosedPublicResults(input, inputDigest, certificate, certificateDigest, baseline, applied, replay); err != nil {
		return err
	}
	caseReports, counts, artifacts, err := verifyPublicCases(casesRoot)
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
		CompilerDigest: compilerDigest, ToolchainDigest: generation.SemanticRetentionToolchainDigest(), VerifierDigest: verifierDigest, PolicyDigest: cache.HashBytes(contract).String(),
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

func verifyClosedPublicResults(input []byte, inputDigest string, certificate generation.SemanticRetentionCertificate, certificateDigest string, baseline, applied, replay generation.SemanticPublicGenerationReport) error {
	if err := generation.ValidateSemanticPublicGenerationReport(baseline); err != nil {
		return fmt.Errorf("baseline report: %w", err)
	}
	if err := generation.ValidateSemanticPublicGenerationReport(applied); err != nil {
		return fmt.Errorf("applied report: %w", err)
	}
	if err := generation.ValidateSemanticPublicGenerationReport(replay); err != nil {
		return fmt.Errorf("replay report: %w", err)
	}
	if baseline.Reason != generation.SemanticPublicGenerationBaselineReason || applied.Reason != generation.SemanticPublicGenerationHitReason || replay.Reason != generation.SemanticPublicGenerationHitReason {
		return errors.New("public generation reports do not identify baseline and retained paths")
	}
	if baseline.InputSourceDigest != inputDigest || applied.InputSourceDigest != inputDigest || replay.InputSourceDigest != inputDigest {
		return errors.New("public generation reports are not bound to the exact input")
	}
	if applied.CertificateDigest != certificateDigest || replay.CertificateDigest != certificateDigest || applied.GeneratedOutputDigest != certificate.GeneratedOutputDigest || replay.GeneratedOutputDigest != certificate.GeneratedOutputDigest {
		return errors.New("public generation reports are not bound to the certificate")
	}
	for _, report := range []generation.SemanticPublicGenerationReport{applied, replay} {
		if report.AdoptionReportDigest != certificate.AdoptionReportDigest || report.ObservationDigest != certificate.ObservationDigest || report.ProposalDigest != certificate.ProposalDigest || report.AuthorizationDigest != certificate.AuthorizationDigest || report.CandidateStableID != certificate.CandidateStableID || report.ContractSourceDigest != certificate.ContractSourceDigest || report.CompilerDigest != certificate.CompilerDigest || report.ToolchainDigest != certificate.ToolchainDigest || report.VerifierDigest != certificate.VerifierDigest || report.PolicyDigest != certificate.PolicyDigest || report.GeneratedManifestDigest != certificate.GeneratedManifestDigest {
			return errors.New("public generation report dropped an evidence binding")
		}
	}
	baselineSource, err := os.ReadFile(baseline.OutputFile)
	if err != nil {
		return fmt.Errorf("read baseline source: %w", err)
	}
	appliedSource, err := os.ReadFile(applied.OutputFile)
	if err != nil {
		return fmt.Errorf("read applied source: %w", err)
	}
	replaySource, err := os.ReadFile(replay.OutputFile)
	if err != nil {
		return fmt.Errorf("read replay source: %w", err)
	}
	if !bytes.Equal(baselineSource, certificate.GeneratedSource) || !bytes.Equal(baselineSource, appliedSource) || !bytes.Equal(appliedSource, replaySource) {
		return errors.New("public compiler output bytes changed across baseline, certificate, and replay")
	}
	if baseline.NormalizedIRDigest != applied.NormalizedIRDigest || applied.NormalizedIRDigest != replay.NormalizedIRDigest || applied.NormalizedIRDigest != certificate.NormalizedIRDigest {
		return errors.New("public compiler normalized semantic identity changed across invocations")
	}
	if baseline.Metrics.SemanticOperationCount != 1 || applied.Metrics.SemanticOperationCount != 0 || replay.Metrics.SemanticOperationCount != 0 || applied.Metrics.CertificateHits != 1 || replay.Metrics.CertificateHits != 1 {
		return errors.New("public compiler operation or certificate counts changed")
	}
	if directoryFileCount(filepath.Dir(baseline.OutputFile)) != 4 || directoryFileCount(filepath.Dir(applied.OutputFile)) != 4 || directoryFileCount(filepath.Dir(replay.OutputFile)) != 4 {
		return errors.New("public compiler artifact denominator changed")
	}
	for _, report := range []generation.SemanticPublicGenerationReport{baseline, applied, replay} {
		if report.OutputBytes != directoryFileBytes(filepath.Dir(report.OutputFile)) {
			return errors.New("public compiler output byte count is not exact")
		}
	}
	manifest, err := os.ReadFile(applied.ManifestFile)
	if err != nil {
		return fmt.Errorf("read applied manifest: %w", err)
	}
	if !bytes.Equal(manifest, certificate.GeneratedManifest) {
		return errors.New("public compiler manifest bytes changed on certificate application")
	}
	_ = input
	return nil
}

func verifyPublicCases(root string) ([]publicCase, map[string]int, int, error) {
	ids := []string{"CERTIFICATE_HIT", "CERTIFICATE_REPLAY", "MISSING_CERTIFICATE", "MISSING_AUTHORIZATION", "STALE_INPUT", "MISMATCHED_CERTIFICATE"}
	directories := []string{"certificate-hit", "certificate-replay", "missing-certificate", "missing-authorization", "stale-input", "mismatched-certificate"}
	want := map[string]struct{ decision, reason string }{
		"CERTIFICATE_HIT": {"CLOSED", generation.SemanticPublicGenerationHitReason}, "CERTIFICATE_REPLAY": {"CLOSED", generation.SemanticPublicGenerationHitReason},
		"MISSING_CERTIFICATE": {"UNKNOWN", generation.SemanticRetentionUnknownCertificateReason}, "MISSING_AUTHORIZATION": {"UNKNOWN", generation.SemanticRetentionUnknownAuthorizationReason},
		"STALE_INPUT": {generation.SemanticRetentionRefuted, generation.SemanticRetentionRefutedReason}, "MISMATCHED_CERTIFICATE": {generation.SemanticRetentionRefuted, generation.SemanticRetentionRefutedReason},
	}
	counts := map[string]int{"CLOSED": 0, "UNKNOWN": 0, "REFUTED": 0}
	results := make([]publicCase, 0, len(ids))
	total := 0
	for index, id := range ids {
		report, err := readReport(filepath.Join(root, directories[index]))
		if err != nil {
			return nil, nil, 0, fmt.Errorf("case %s: %w", id, err)
		}
		expected := want[id]
		if report.Decision != expected.decision || report.Reason != expected.reason {
			return nil, nil, 0, fmt.Errorf("case %s = %s/%s, want %s/%s", id, report.Decision, report.Reason, expected.decision, expected.reason)
		}
		artifacts := directoryFileCount(filepath.Join(root, directories[index]))
		if artifacts != expectedArtifacts(report.Decision) {
			return nil, nil, 0, fmt.Errorf("case %s artifacts = %d", id, artifacts)
		}
		caseDirectory := filepath.Join(root, directories[index])
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
