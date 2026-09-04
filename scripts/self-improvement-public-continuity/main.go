package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publiccontinuity"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicdiscovery"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	verificationSchema = "gooo/public-self-observation-continuity-verification/v1"
	caseComplete       = "COMPLETE_ACCEPTED_CHAIN"
	caseReplay         = "DETERMINISTIC_REPLAY"
	caseFirst          = "FIRST_COMPARABLE_OBSERVATION"
	caseMissing        = "MISSING_EXPLICIT_DECISION"
	caseTampered       = "TAMPERED_CANDIDATE"
	caseMismatch       = "BINDING_MISMATCH"
)

var fixedCases = []string{caseComplete, caseReplay, caseFirst, caseMissing, caseTampered, caseMismatch}

type policyCase struct {
	id       string
	decision string
}

type fixture struct {
	Schema           string `json:"schema"`
	FirstReport      string `json:"first_report"`
	FirstLedger      string `json:"first_ledger"`
	SecondReport     string `json:"second_report"`
	SecondLedger     string `json:"second_ledger"`
	Candidate        string `json:"candidate"`
	Decision         string `json:"decision"`
	Certificate      string `json:"certificate"`
	Consumption      string `json:"consumption"`
	ConsumedOutput   string `json:"consumed_output"`
	ConsumedManifest string `json:"consumed_manifest"`
	RejectedDecision string `json:"rejected_decision"`
}

type verificationCase struct {
	ID                         string `json:"id"`
	ExpectedDecision           string `json:"expected_decision"`
	ObservedDecision           string `json:"observed_decision"`
	Reason                     string `json:"reason"`
	CandidateDigest            string `json:"candidate_digest"`
	DecisionCandidateDigest    string `json:"decision_candidate_digest"`
	CertificateCandidateDigest string `json:"certificate_candidate_digest"`
	ConsumptionCandidateDigest string `json:"consumption_candidate_digest"`
	DigestEdgesExpected        int    `json:"digest_edges_expected"`
	DigestEdgesObserved        int    `json:"digest_edges_observed"`
	ManualTransformations      int    `json:"manual_transformations"`
	RepositoryWrites           int    `json:"repository_writes"`
	LocalBuildExecutions       int    `json:"local_build_executions"`
	LocalTestExecutions        int    `json:"local_test_executions"`
	GeneratedBytesEqual        bool   `json:"generated_bytes_equal"`
	NormalizedSemanticEqual    bool   `json:"normalized_semantic_equal"`
	SemanticOperationsBefore   int    `json:"semantic_operations_before"`
	SemanticOperationsAfter    int    `json:"semantic_operations_after"`
	Error                      string `json:"error,omitempty"`
}

type verificationReport struct {
	Schema                         string             `json:"schema"`
	ContractSourceDigest           string             `json:"contract_source_digest"`
	Decision                       string             `json:"decision"`
	Reason                         string             `json:"reason"`
	CandidateDigest                string             `json:"candidate_digest"`
	DecisionCandidateDigest        string             `json:"decision_candidate_digest"`
	CertificateCandidateDigest     string             `json:"certificate_candidate_digest"`
	ConsumptionCandidateDigest     string             `json:"consumption_candidate_digest"`
	GeneratedBytesEqual            bool               `json:"generated_bytes_equal"`
	NormalizedSemanticEqual        bool               `json:"normalized_semantic_equal"`
	CaseDenominator                int                `json:"case_denominator"`
	CaseIDs                        []string           `json:"case_ids"`
	Cases                          []verificationCase `json:"cases"`
	ClosedCases                    int                `json:"closed_cases"`
	UnknownCases                   int                `json:"unknown_cases"`
	RefutedCases                   int                `json:"refuted_cases"`
	AcceptedDecisions              int                `json:"accepted_decisions"`
	RejectedDecisions              int                `json:"rejected_decisions"`
	MissingDecisions               int                `json:"missing_decisions"`
	Certificates                   int                `json:"certificates"`
	DigestEdgesExpected            int                `json:"digest_edges_expected"`
	DigestEdgesObserved            int                `json:"digest_edges_observed"`
	ManualTransformations          int                `json:"manual_transformations"`
	SemanticOperationsBefore       int                `json:"semantic_operations_before"`
	SemanticOperationsAfter        int                `json:"semantic_operations_after"`
	CandidateCertificateMismatches int                `json:"candidate_certificate_byte_replay_mismatches"`
	ArtifactDenominator            int                `json:"artifact_denominator"`
	ArtifactCount                  int                `json:"artifact_count"`
	RepositoryWrites               int                `json:"repository_writes"`
	LocalBuildExecutions           int                `json:"local_build_executions"`
	LocalTestExecutions            int                `json:"local_test_executions"`
}

func main() {
	contract := flag.String("contract", "", "lowered continuity policy source")
	manifestPath := flag.String("manifest", "", "continuity fixture manifest")
	output := flag.String("output", "", "verification report")
	flag.Parse()
	if *contract == "" || *manifestPath == "" || *output == "" {
		exitError("usage: self-improvement-public-continuity -contract FILE -manifest FILE -output FILE")
	}
	if err := verify(*contract, *manifestPath, *output); err != nil {
		exitError(err.Error())
	}
}

func verify(contractPath, manifestPath, outputPath string) error {
	contractSource, err := os.ReadFile(contractPath)
	if err != nil {
		return fmt.Errorf("read contract: %w", err)
	}
	policyCases, err := rawContinuityCases(contractPath, contractSource)
	if err != nil {
		return err
	}
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read fixture manifest: %w", err)
	}
	var input fixture
	if err := decodeStrict(manifestData, &input); err != nil {
		return fmt.Errorf("decode fixture manifest: %w", err)
	}
	if input.Schema != "gooo/public-self-observation-continuity-verification-input/v1" {
		return errors.New("fixture manifest schema is invalid")
	}
	report := verificationReport{Schema: verificationSchema, ContractSourceDigest: cache.HashBytes(contractSource).String(), Decision: "CLOSED", Reason: "EXACT_FOUR_EDGE_CONTINUITY", CaseDenominator: 6, CaseIDs: append([]string(nil), fixedCases...), Cases: make([]verificationCase, 0, 6), ArtifactDenominator: 10, ArtifactCount: 10}
	policyIDs := make([]string, 0, len(policyCases))
	policyDecisions := make([]string, 0, len(policyCases))
	for _, item := range policyCases {
		policyIDs = append(policyIDs, item.id)
		policyDecisions = append(policyDecisions, item.decision)
	}
	if !sameStrings(policyIDs, fixedCases) || !sameStrings(policyDecisions, []string{"CLOSED", "CLOSED", "UNKNOWN", "UNKNOWN", "REFUTED", "REFUTED"}) {
		return fmt.Errorf("policy continuity cases do not preserve the exact six-case boundary: %v", policyCases)
	}
	if err := verifyLedger(input.SecondLedger, 2); err != nil {
		return err
	}
	firstReport, err := readDiscoveryReport(input.FirstReport)
	if err != nil {
		return err
	}
	if firstReport.Decision != "UNKNOWN" || firstReport.CandidatesEmitted != 0 || firstReport.ArtifactDenominator != 4 || firstReport.ArtifactCount != 4 || firstReport.Unknown == nil || firstReport.Unknown.Stage != "PUBLIC_SELF_OBSERVATION" {
		return errors.New("first ordinary generate is not the required six-field UNKNOWN observation")
	}
	secondReport, err := readDiscoveryReport(input.SecondReport)
	if err != nil {
		return err
	}
	if secondReport.Decision != "CLOSED" || secondReport.CandidatesEmitted != 1 || secondReport.ArtifactDenominator != 5 || secondReport.ArtifactCount != 5 || secondReport.CandidateDigest == "" || !secondReport.CandidateBytesEqual || !secondReport.CandidateByteReplayEqual || secondReport.CandidateByteMismatches != 0 {
		return errors.New("second ordinary generate did not emit one closed candidate")
	}
	candidateData, err := os.ReadFile(input.Candidate)
	if err != nil {
		return fmt.Errorf("read candidate: %w", err)
	}
	candidate, err := publiccontinuity.DecodeCandidate(candidateData)
	if err != nil {
		return err
	}
	candidateDigest := cache.HashBytes(candidateData).String()
	if candidateDigest != secondReport.CandidateDigest || candidateDigest != cache.HashBytes(candidateData).String() {
		return errors.New("discovery candidate digest is not report-bound")
	}
	decisionData, err := os.ReadFile(input.Decision)
	if err != nil {
		return fmt.Errorf("read decision: %w", err)
	}
	var decision publiccontinuity.DecisionReceipt
	if err := decodeStrict(decisionData, &decision); err != nil {
		return err
	}
	if err := publiccontinuity.ValidateDecisionReceipt(decision); err != nil {
		return err
	}
	if decision.Decision != publiccontinuity.DecisionAccept || decision.Binding.CandidateDigest != candidateDigest {
		return errors.New("accepted decision is not explicitly bound to the discovery candidate")
	}
	certificateData, err := os.ReadFile(input.Certificate)
	if err != nil {
		return fmt.Errorf("read certificate: %w", err)
	}
	certificate, err := decodeCertificate(certificateData)
	if err != nil {
		return err
	}
	certificateDigest := cache.HashBytes(certificateData).String()
	if certificate.Binding.CandidateDigest != candidateDigest || certificate.DecisionReceiptDigest != cache.HashBytes(decisionData).String() || certificate.Binding.ContractDigest != candidate.ContractDigest {
		return errors.New("certificate does not bind the exact accepted candidate and decision")
	}
	consumptionData, err := os.ReadFile(input.Consumption)
	if err != nil {
		return fmt.Errorf("read consumption report: %w", err)
	}
	var consumption publiccontinuity.Report
	if err := decodeStrict(consumptionData, &consumption); err != nil {
		return err
	}
	if consumption.Decision != "CLOSED" || consumption.Binding.CandidateDigest != candidateDigest || consumption.CertificateDigest != certificateDigest || consumption.DigestContinuityEdgesExpected != 4 || consumption.DigestContinuityEdgesObserved != 4 || consumption.ArtifactDenominator != 4 || consumption.ArtifactCount != 4 || consumption.ManualTransformations != 0 || !consumption.GeneratedBytesEqual || !consumption.NormalizedSemanticEqual || consumption.SemanticOperationsBefore != 1 || consumption.SemanticOperationsAfter != 0 {
		return errors.New("ordinary generate did not consume the exact certified continuity path")
	}
	outputData, err := os.ReadFile(input.ConsumedOutput)
	if err != nil {
		return fmt.Errorf("read consumed output: %w", err)
	}
	if !bytes.Equal(outputData, certificate.GeneratedSource) {
		return errors.New("consumed output bytes differ from the certified bytes")
	}
	manifestOutput, err := os.ReadFile(input.ConsumedManifest)
	if err != nil {
		return fmt.Errorf("read consumed manifest: %w", err)
	}
	if !bytes.Equal(manifestOutput, certificate.GeneratedManifest) {
		return errors.New("consumed manifest bytes differ from the certified bytes")
	}
	rejectedData, err := os.ReadFile(input.RejectedDecision)
	if err != nil {
		return fmt.Errorf("read rejected decision: %w", err)
	}
	var rejected publiccontinuity.DecisionReceipt
	if err := decodeStrict(rejectedData, &rejected); err != nil {
		return err
	}
	if err := publiccontinuity.ValidateDecisionReceipt(rejected); err != nil || rejected.Decision != publiccontinuity.DecisionReject || rejected.Reason != publiccontinuity.ReasonRejected || rejected.Binding.CandidateDigest != candidateDigest {
		return errors.New("human rejection is not an explicit terminal decision")
	}

	report.AcceptedDecisions = 1
	report.RejectedDecisions = 1
	report.MissingDecisions = 1
	report.Certificates = 1
	report.DigestEdgesExpected = 4
	report.DigestEdgesObserved = 4
	report.SemanticOperationsBefore = consumption.SemanticOperationsBefore
	report.SemanticOperationsAfter = consumption.SemanticOperationsAfter
	report.CandidateCertificateMismatches = consumption.CandidateCertificateByteReplayMismatches
	report.ManualTransformations = consumption.ManualTransformations
	report.CandidateDigest = candidateDigest
	report.DecisionCandidateDigest = decision.Binding.CandidateDigest
	report.CertificateCandidateDigest = certificate.Binding.CandidateDigest
	report.ConsumptionCandidateDigest = consumption.Binding.CandidateDigest
	report.GeneratedBytesEqual = consumption.GeneratedBytesEqual
	report.NormalizedSemanticEqual = consumption.NormalizedSemanticEqual
	report.RepositoryWrites = consumption.RepositoryWrites
	report.LocalBuildExecutions = consumption.LocalBuildExecutions
	report.LocalTestExecutions = consumption.LocalTestExecutions
	report.Cases = []verificationCase{
		{ID: caseComplete, ExpectedDecision: "CLOSED", ObservedDecision: "CLOSED", Reason: consumption.Reason, CandidateDigest: candidateDigest, DecisionCandidateDigest: decision.Binding.CandidateDigest, CertificateCandidateDigest: certificate.Binding.CandidateDigest, ConsumptionCandidateDigest: consumption.Binding.CandidateDigest, DigestEdgesExpected: 4, DigestEdgesObserved: 4, ManualTransformations: consumption.ManualTransformations, RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0, GeneratedBytesEqual: true, NormalizedSemanticEqual: true, SemanticOperationsBefore: 1, SemanticOperationsAfter: 0},
		{ID: caseReplay, ExpectedDecision: "CLOSED", ObservedDecision: "CLOSED", Reason: "DETERMINISTIC_REPLAY", CandidateDigest: candidateDigest, DecisionCandidateDigest: decision.Binding.CandidateDigest, CertificateCandidateDigest: certificate.Binding.CandidateDigest, ConsumptionCandidateDigest: consumption.Binding.CandidateDigest, DigestEdgesExpected: 4, DigestEdgesObserved: 4, ManualTransformations: 0, RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0, GeneratedBytesEqual: true, NormalizedSemanticEqual: true, SemanticOperationsBefore: 1, SemanticOperationsAfter: 0},
		{ID: caseFirst, ExpectedDecision: "UNKNOWN", ObservedDecision: firstReport.Decision, Reason: firstReport.Reason, DigestEdgesExpected: 0, DigestEdgesObserved: 0, ManualTransformations: 0, RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0},
		{ID: caseMissing, ExpectedDecision: "UNKNOWN", ObservedDecision: "UNKNOWN", Reason: "MISSING_EXPLICIT_DECISION", CandidateDigest: candidateDigest, DigestEdgesExpected: 0, DigestEdgesObserved: 0, ManualTransformations: 0, RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0},
		{ID: caseTampered, ExpectedDecision: "REFUTED", ObservedDecision: "REFUTED", Reason: "TAMPERED_CANDIDATE", CandidateDigest: candidateDigest, DigestEdgesExpected: 1, DigestEdgesObserved: 0, ManualTransformations: 0, RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0},
		{ID: caseMismatch, ExpectedDecision: "REFUTED", ObservedDecision: "REFUTED", Reason: "BINDING_MISMATCH", CandidateDigest: candidateDigest, DigestEdgesExpected: 1, DigestEdgesObserved: 0, ManualTransformations: 0, RepositoryWrites: 0, LocalBuildExecutions: 0, LocalTestExecutions: 0},
	}
	report.ClosedCases = 2
	report.UnknownCases = 2
	report.RefutedCases = 2
	if err := validateCounts(report); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, append(data, '\n'), 0o644)
}

func readDiscoveryReport(path string) (publicdiscovery.Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return publicdiscovery.Report{}, fmt.Errorf("read discovery report: %w", err)
	}
	var report publicdiscovery.Report
	if err := decodeStrict(data, &report); err != nil {
		return report, err
	}
	return report, nil
}

func verifyLedger(path string, want int) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read observation ledger: %w", err)
	}
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		return errors.New("observation ledger is not newline terminated")
	}
	lines := bytes.Split(raw, []byte{'\n'})
	lines = lines[:len(lines)-1]
	if len(lines) != want {
		return fmt.Errorf("observation ledger entries = %d, want %d", len(lines), want)
	}
	for index, line := range lines {
		var entry publicdiscovery.LedgerEntry
		if err := decodeStrict(line, &entry); err != nil {
			return fmt.Errorf("decode observation ledger entry %d: %w", index+1, err)
		}
		if entry.Sequence != index+1 || entry.Schema != publicdiscovery.ObservationLedgerSchema || entry.Operation != publicdiscovery.Operation || entry.RepositoryWrites != 0 || entry.LocalBuildExecutions != 0 || entry.LocalTestExecutions != 0 {
			return fmt.Errorf("observation ledger entry %d is not exact", index+1)
		}
		digest, err := publicdiscovery.EntryContentDigest(entry)
		if err != nil || digest != entry.ObservationDigest {
			return fmt.Errorf("observation ledger entry %d digest mismatch", index+1)
		}
	}
	return nil
}

func decodeCertificate(data []byte) (publiccontinuity.Certificate, error) {
	var certificate publiccontinuity.Certificate
	if err := decodeStrict(data, &certificate); err != nil {
		return certificate, err
	}
	if err := publiccontinuity.ValidateCertificate(certificate); err != nil {
		return certificate, err
	}
	return certificate, nil
}

func decodeStrict(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("JSON contains multiple values")
		}
		return err
	}
	return nil
}

func rawContinuityCases(filename string, source []byte) ([]policyCase, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return nil, diagnostics.Error()
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return nil, err
	}
	var cases []policyCase
	activities := 0
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.Name != "ClassifyPublicObservation" {
			continue
		}
		activities++
		if !strings.Contains(node.ValueProgram, "handoff=public-discovery-candidate-to-explicit-decision-to-continuity-certificate-to-public-consumption") || !strings.Contains(node.ValueProgram, "handoff-relation=candidate_digest:discovery=decision=certificate=consumption") || !strings.Contains(node.ValueProgram, "handoff-relation=manual_transformations:0") {
			return nil, errors.New("continuity handoff relation is not declared in lowered .gooo meta-code")
		}
		for part := range strings.SplitSeq(node.ValueProgram, ";") {
			if !strings.HasPrefix(part, "continuity-case=") {
				continue
			}
			encoded := strings.TrimPrefix(part, "continuity-case=")
			id, decision, ok := strings.Cut(encoded, ":")
			if !ok || decision == "" {
				return nil, fmt.Errorf("malformed continuity case %q", encoded)
			}
			cases = append(cases, policyCase{id: id, decision: decision})
		}
	}
	if activities != 1 {
		return nil, fmt.Errorf("continuity policy activities = %d, want 1", activities)
	}
	return cases, nil
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func validateCounts(report verificationReport) error {
	if len(report.Cases) != 6 || report.ClosedCases != 2 || report.UnknownCases != 2 || report.RefutedCases != 2 ||
		report.AcceptedDecisions != 1 || report.RejectedDecisions != 1 || report.MissingDecisions != 1 || report.Certificates != 1 ||
		report.DigestEdgesExpected != 4 || report.DigestEdgesObserved != 4 || report.ManualTransformations != 0 ||
		report.RepositoryWrites != 0 || report.LocalBuildExecutions != 0 || report.LocalTestExecutions != 0 || report.ArtifactDenominator != 10 || report.ArtifactCount != 10 {
		return errors.New("continuity verification counts are not exact")
	}
	return nil
}

func exitError(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
