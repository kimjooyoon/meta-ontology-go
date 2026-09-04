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

type retainedScenario struct {
	ID                string `json:"id"`
	Decision          string `json:"decision"`
	Reason            string `json:"reason"`
	EnvelopeScenario  string `json:"envelope_scenario"`
	EnvelopeDecision  string `json:"envelope_decision"`
	EnvelopeArtifacts int    `json:"envelope_artifacts"`
}

type retainedLoopReport struct {
	Schema                   string                                     `json:"schema"`
	CertificateDigest        string                                     `json:"certificate_digest"`
	CandidateStableID        string                                     `json:"candidate_stable_id"`
	CaseDenominator          int                                        `json:"case_denominator"`
	Counts                   map[string]int                             `json:"counts"`
	EnvelopeScenarioDenom    int                                        `json:"envelope_scenario_denominator"`
	EnvelopeArtifactsPerCase int                                        `json:"envelope_artifacts_per_case"`
	TotalArtifacts           int                                        `json:"total_artifacts"`
	BeforeMetrics            generation.SemanticRetentionRuntimeMetrics `json:"before_metrics"`
	AfterMetrics             generation.SemanticRetentionRuntimeMetrics `json:"after_metrics"`
	ReplayMetrics            generation.SemanticRetentionRuntimeMetrics `json:"replay_metrics"`
	GeneratedBytesEqual      bool                                       `json:"generated_bytes_equal"`
	NormalizedSemanticEqual  bool                                       `json:"normalized_semantic_equal"`
	Comparable               bool                                       `json:"comparable"`
	SuperiorityClaim         bool                                       `json:"superiority_claim"`
	NoSuperiorityClaim       bool                                       `json:"no_superiority_claim"`
	Cases                    []retainedScenario                         `json:"cases"`
	RepositoryWrites         int                                        `json:"repository_writes"`
	LocalTestExecutions      int                                        `json:"local_test_executions"`
	PolicyActivitiesExpected int                                        `json:"policy_activities_expected"`
	PolicyActivitiesObserved int                                        `json:"policy_activities_observed"`
	PolicyDecisionRows       int                                        `json:"policy_decision_rows"`
	GeneratedEvaluatorDigest string                                     `json:"generated_evaluator_digest"`
	GeneratedEvaluatorCases  int                                        `json:"generated_evaluator_cases_bound"`
	UnboundCases             int                                        `json:"unbound_cases"`
	DuplicateCases           int                                        `json:"duplicate_cases"`
	RegenerationMismatches   int                                        `json:"regeneration_byte_mismatches"`
}

type sourceRetentionPolicy struct {
	ActivityCount  int
	Cases          []string
	Decisions      map[string]string
	UnboundCases   int
	DuplicateCases int
}

// deriveRetentionPolicy is deliberately independent of the generated
// evaluator: it lowers the raw .gooo graph and reads the decision rows from
// the PublishOperationReceipt activity as the verifier's expectation.
func deriveRetentionPolicy(filename string, source []byte) (sourceRetentionPolicy, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return sourceRetentionPolicy{}, fmt.Errorf("independent retention policy parse: %w", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceRetentionPolicy{}, fmt.Errorf("independent retention policy lowering: %w", err)
	}
	policy := sourceRetentionPolicy{Decisions: make(map[string]string)}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.Name != retentionpolicy.PolicyActivity {
			continue
		}
		policy.ActivityCount++
		seenHeader := false
		for _, part := range strings.Split(node.ValueProgram, ";") {
			switch {
			case part == retentionpolicy.RetentionDecisionHead:
				if seenHeader {
					return sourceRetentionPolicy{}, errors.New("independent retention policy header is duplicated")
				}
				seenHeader = true
			case strings.HasPrefix(part, "retention-case="):
				id, decision, ok := strings.Cut(strings.TrimPrefix(part, "retention-case="), ":")
				if !ok || id == "" || decision == "" {
					return sourceRetentionPolicy{}, fmt.Errorf("independent retention row %q is malformed", part)
				}
				if _, exists := policy.Decisions[id]; exists {
					policy.DuplicateCases++
					continue
				}
				policy.Decisions[id] = decision
				policy.Cases = append(policy.Cases, id)
			}
		}
		if !seenHeader {
			return sourceRetentionPolicy{}, errors.New("independent retention policy header is missing")
		}
	}
	if policy.ActivityCount != 1 {
		return sourceRetentionPolicy{}, fmt.Errorf("independent retention policy activities = %d, want 1", policy.ActivityCount)
	}
	expectedIDs := generation.SemanticRetentionCaseIDs()
	if len(policy.Cases) != len(expectedIDs) {
		policy.UnboundCases = len(expectedIDs) - len(policy.Cases)
		if policy.UnboundCases < 0 {
			policy.UnboundCases = 0
		}
		return policy, fmt.Errorf("independent retention decision rows = %d, want %d", len(policy.Cases), len(expectedIDs))
	}
	for index, id := range expectedIDs {
		if policy.Cases[index] != id {
			return sourceRetentionPolicy{}, fmt.Errorf("independent retention row %d = %q, want %q", index+1, policy.Cases[index], id)
		}
		if decision := policy.Decisions[id]; decision != "CLOSED" && decision != "UNKNOWN" && decision != "REFUTED" {
			return sourceRetentionPolicy{}, fmt.Errorf("independent retention decision %q is outside the fixed set", decision)
		}
	}
	return policy, nil
}

func main() {
	contractPath := flag.String("contract", "", "observation authority .gooo")
	inputPath := flag.String("input", "", "exact input .gooo")
	observationPath := flag.String("observation", "", "compiler observation JSON")
	proposalPath := flag.String("proposal", "", "adoption proposal JSON")
	authorizationPath := flag.String("authorization", "", "explicit authorization JSON")
	adoptionPath := flag.String("adoption", "", "authorized adoption result JSON")
	certificatePath := flag.String("certificate", "", "immutable retained-knowledge certificate JSON")
	beforePath := flag.String("before", "", "first independent invocation result JSON")
	afterPath := flag.String("after", "", "later certificate-hit result JSON")
	replayPath := flag.String("replay", "", "later replay result JSON")
	missingCertificatePath := flag.String("missing-certificate", "", "missing-certificate UNKNOWN result JSON")
	missingAuthorizationPath := flag.String("missing-authorization", "", "missing-authorization UNKNOWN result JSON")
	staleInputPath := flag.String("stale-input", "", "stale-input REFUTED result JSON")
	mismatchedCertificatePath := flag.String("mismatched-certificate", "", "mismatched-certificate REFUTED result JSON")
	envelopeRoot := flag.String("envelope-root", "", "caller-owned envelope output root")
	outputPath := flag.String("output", "", "caller-owned retained loop report JSON")
	flag.Parse()
	if *contractPath == "" || *inputPath == "" || *observationPath == "" || *proposalPath == "" || *authorizationPath == "" ||
		*adoptionPath == "" || *certificatePath == "" || *beforePath == "" || *afterPath == "" || *replayPath == "" ||
		*missingCertificatePath == "" || *missingAuthorizationPath == "" || *staleInputPath == "" || *mismatchedCertificatePath == "" ||
		*envelopeRoot == "" || *outputPath == "" {
		fail(errors.New("self-improvement-retained-knowledge requires all evidence, six case, envelope, and output paths"))
	}
	if err := run(*contractPath, *inputPath, *observationPath, *proposalPath, *authorizationPath, *adoptionPath, *certificatePath, *beforePath, *afterPath, *replayPath, *missingCertificatePath, *missingAuthorizationPath, *staleInputPath, *mismatchedCertificatePath, *envelopeRoot, *outputPath); err != nil {
		fail(err)
	}
}

func run(contractPath, inputPath, observationPath, proposalPath, authorizationPath, adoptionPath, certificatePath, beforePath, afterPath, replayPath, missingCertificatePath, missingAuthorizationPath, staleInputPath, mismatchedCertificatePath, envelopeRoot, outputPath string) error {
	contract, err := os.ReadFile(contractPath)
	if err != nil {
		return fmt.Errorf("read contract: %w", err)
	}
	sourcePolicy, err := deriveRetentionPolicy(contractPath, contract)
	if err != nil {
		return err
	}
	generatedPolicy, generatedEvaluator, err := retentionpolicy.GenerateNamed(contractPath, contract)
	if err != nil {
		return fmt.Errorf("generate retention evaluator: %w", err)
	}
	checkedInEvaluator, err := os.ReadFile("internal/meta/retentionpolicy/generated/evaluator.go")
	if err != nil {
		return fmt.Errorf("read checked-in retention evaluator: %w", err)
	}
	regenerationMismatches := 0
	if !bytes.Equal(generatedEvaluator, checkedInEvaluator) {
		regenerationMismatches = 1
	}
	if generatedPolicy.EvaluatorDigest != retentionpolicy.GeneratedEvaluatorDigest() ||
		!reflect.DeepEqual(sourcePolicy.Cases, retentionpolicy.GeneratedEvaluatorCaseIDs()) {
		return errors.New("generated retention evaluator is not bound to the independently lowered policy")
	}
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	observationData, observation, err := readObservation(observationPath)
	if err != nil {
		return err
	}
	proposalData, proposal, err := readProposal(proposalPath)
	if err != nil {
		return err
	}
	authorizationData, authorization, err := readAuthorization(authorizationPath)
	if err != nil {
		return err
	}
	adoptionData, adoption, err := readAdoption(adoptionPath)
	if err != nil {
		return err
	}
	certificateData, certificate, err := readCertificate(certificatePath)
	if err != nil {
		return err
	}
	before, err := readResult(beforePath)
	if err != nil {
		return err
	}
	after, err := readResult(afterPath)
	if err != nil {
		return err
	}
	replay, err := readResult(replayPath)
	if err != nil {
		return err
	}
	missingCertificate, err := readResult(missingCertificatePath)
	if err != nil {
		return err
	}
	missingAuthorization, err := readResult(missingAuthorizationPath)
	if err != nil {
		return err
	}
	staleInput, err := readResult(staleInputPath)
	if err != nil {
		return err
	}
	mismatchedCertificate, err := readResult(mismatchedCertificatePath)
	if err != nil {
		return err
	}

	if err := verifySourceEvidence(contract, input, observationData, observation, proposalData, proposal, authorizationData, authorization, adoptionData, adoption); err != nil {
		return err
	}
	compilerDigest, err := generation.SemanticRetentionCompilerDigest(os.ReadFile)
	if err != nil {
		return fmt.Errorf("compiler manifest: %w", err)
	}
	verifierDigest, err := generation.SemanticRetentionVerifierDigest(os.ReadFile)
	if err != nil {
		return fmt.Errorf("verifier manifest: %w", err)
	}
	bindings := generation.SemanticRetentionBindings{
		AdoptionReportDigest: cache.HashBytes(adoptionData).String(),
		ObservationDigest:    cache.HashBytes(observationData).String(),
		ProposalDigest:       cache.HashBytes(proposalData).String(),
		AuthorizationDigest:  cache.HashBytes(authorizationData).String(),
		CandidateStableID:    proposal.Candidate.StableID,
		ContractSourceDigest: cache.HashBytes(contract).String(),
		InputSourceDigest:    cache.HashBytes(input).String(),
		NormalizedIRDigest:   proposal.Candidate.InputDigest,
		CompilerDigest:       compilerDigest,
		ToolchainDigest:      generation.SemanticRetentionToolchainDigest(),
		VerifierDigest:       verifierDigest,
		PolicyDigest:         cache.HashBytes(contract).String(),
		EvaluatorDigest:      retentionpolicy.GeneratedEvaluatorDigest(),
	}
	if err := generation.VerifySemanticRetentionCertificate(certificate, bindings); err != nil {
		return fmt.Errorf("independent certificate verification: %w", err)
	}
	if certificate.EvaluatorDigest != generatedPolicy.EvaluatorDigest {
		return errors.New("retained certificate evaluator digest is not bound to regenerated policy")
	}
	if certificate.GeneratedOutputDigest != adoption.Evidence.BeforeOutputDigest || certificate.NormalizedIRDigest != adoption.Evidence.BeforeSemanticDigest {
		return errors.New("retained certificate is not bound to the independently verified adoption result")
	}
	computedDigest, err := independentlyComputeInputDigest(inputPath, input)
	if err != nil {
		return err
	}
	if computedDigest != proposal.Candidate.InputDigest || computedDigest != certificate.NormalizedIRDigest {
		return errors.New("independent normalized semantic digest does not match the retained certificate")
	}
	certificateDigest := cache.HashBytes(certificateData).String()
	if err := verifyClosedPair(certificate, certificateDigest, before, after, replay); err != nil {
		return err
	}

	unknownObservation := unknownObservation(observation)
	refutedObservation := refutedObservation(observation)
	caseResults := map[string]generation.SemanticRetentionResult{
		"CERTIFICATE_HIT":        after,
		"CERTIFICATE_REPLAY":     replay,
		"MISSING_CERTIFICATE":    missingCertificate,
		"MISSING_AUTHORIZATION":  missingAuthorization,
		"STALE_INPUT":            staleInput,
		"MISMATCHED_CERTIFICATE": mismatchedCertificate,
	}
	caseObservations := map[string]generation.SemanticObservation{
		"CERTIFICATE_HIT":        adoption.Observation,
		"CERTIFICATE_REPLAY":     adoption.Observation,
		"MISSING_CERTIFICATE":    unknownObservation,
		"MISSING_AUTHORIZATION":  unknownObservation,
		"STALE_INPUT":            refutedObservation,
		"MISMATCHED_CERTIFICATE": refutedObservation,
	}
	envelopeIDs := map[string]string{
		"CERTIFICATE_HIT": "C1", "CERTIFICATE_REPLAY": "C2", "MISSING_CERTIFICATE": "U1",
		"MISSING_AUTHORIZATION": "U2", "STALE_INPUT": "R1", "MISMATCHED_CERTIFICATE": "R2",
	}
	if err := os.MkdirAll(envelopeRoot, 0o755); err != nil {
		return fmt.Errorf("create envelope root: %w", err)
	}
	counts := map[string]int{"CLOSED": 0, "UNKNOWN": 0, "REFUTED": 0}
	scenarios := make([]retainedScenario, 0, generation.SemanticRetentionCaseDenominator)
	totalArtifacts := 0
	for _, caseID := range sourcePolicy.Cases {
		result := caseResults[caseID]
		expectedDecision, ok := sourcePolicy.Decisions[caseID]
		if !ok || expectedDecision == "" {
			return fmt.Errorf("unknown retained case %q", caseID)
		}
		if err := verifyCaseResult(caseID, expectedDecision, result); err != nil {
			return err
		}
		if len(caseObservations[caseID].Candidates) == 0 {
			return fmt.Errorf("case %s has no observed candidate", caseID)
		}
		counts[result.Decision]++
		envelopeID := envelopeIDs[caseID]
		envelopeDecision, artifacts, _, err := generateRetentionEnvelope(contract, filepath.Join(envelopeRoot, caseID), envelopeID, caseObservations[caseID])
		if err != nil {
			return err
		}
		if envelopeDecision != expectedDecision || artifacts != generation.SemanticRetentionArtifactsPerCase {
			return fmt.Errorf("retained case %s envelope = %s/%d, want %s/%d", caseID, envelopeDecision, artifacts, expectedDecision, generation.SemanticRetentionArtifactsPerCase)
		}
		totalArtifacts += artifacts
		scenarios = append(scenarios, retainedScenario{ID: caseID, Decision: result.Decision, Reason: result.Reason, EnvelopeScenario: envelopeID, EnvelopeDecision: envelopeDecision, EnvelopeArtifacts: artifacts})
	}
	if len(scenarios) != generation.SemanticRetentionCaseDenominator || counts["CLOSED"] != 2 || counts["UNKNOWN"] != 2 || counts["REFUTED"] != 2 || totalArtifacts != generation.SemanticRetentionCaseDenominator*generation.SemanticRetentionArtifactsPerCase {
		return errors.New("retained knowledge denominator or outcome counts changed")
	}
	generatedBytesEqual := bytes.Equal(before.GeneratedSource, after.GeneratedSource) && bytes.Equal(after.GeneratedSource, replay.GeneratedSource)
	normalizedSemanticEqual := before.NormalizedIRDigest == after.NormalizedIRDigest && after.NormalizedIRDigest == replay.NormalizedIRDigest
	if !generatedBytesEqual || !normalizedSemanticEqual {
		return errors.New("independent retained result comparison changed bytes or normalized semantic identity")
	}
	comparable := before.Metrics.PeakRSSKib > 0 && after.Metrics.PeakRSSKib > 0 &&
		before.CompilerDigest == after.CompilerDigest && before.ToolchainDigest == after.ToolchainDigest &&
		before.VerifierDigest == after.VerifierDigest
	report := retainedLoopReport{
		Schema: generation.SemanticRetentionReportSchema, CertificateDigest: certificateDigest,
		CandidateStableID: proposal.Candidate.StableID, CaseDenominator: generation.SemanticRetentionCaseDenominator,
		Counts: counts, EnvelopeScenarioDenom: generation.SemanticRetentionCaseDenominator,
		EnvelopeArtifactsPerCase: generation.SemanticRetentionArtifactsPerCase, TotalArtifacts: totalArtifacts,
		BeforeMetrics: before.Metrics, AfterMetrics: after.Metrics, ReplayMetrics: replay.Metrics,
		GeneratedBytesEqual: generatedBytesEqual, NormalizedSemanticEqual: normalizedSemanticEqual,
		Comparable: comparable, SuperiorityClaim: false, NoSuperiorityClaim: true, Cases: scenarios,
		RepositoryWrites: 0, LocalTestExecutions: 0,
		PolicyActivitiesExpected: 1, PolicyActivitiesObserved: sourcePolicy.ActivityCount,
		PolicyDecisionRows: len(sourcePolicy.Cases), GeneratedEvaluatorDigest: generatedPolicy.EvaluatorDigest,
		GeneratedEvaluatorCases: len(retentionpolicy.GeneratedEvaluatorCaseIDs()),
		UnboundCases:            sourcePolicy.UnboundCases, DuplicateCases: sourcePolicy.DuplicateCases,
		RegenerationMismatches: regenerationMismatches,
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode retained loop report: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("write retained loop report: %w", err)
	}
	return nil
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

func verifySourceEvidence(contract, input, observationData []byte, observation generation.SemanticObservation, proposalData []byte, proposal generation.SemanticAdoptionProposal, authorizationData []byte, authorization generation.SemanticAdoptionAuthorization, adoptionData []byte, adoption generation.SemanticAdoptionReport) error {
	if err := generation.VerifySemanticObservation(observation); err != nil {
		return fmt.Errorf("observation: %w", err)
	}
	if len(observation.Candidates) != 1 {
		return errors.New("observation must contain exactly one candidate")
	}
	if observation.ContractDigest != cache.HashBytes(contract).String() || observation.InputSourceDigest != cache.HashBytes(input).String() {
		return errors.New("observation is not bound to the exact source and input")
	}
	if err := generation.ValidateSemanticAdoptionProposal(proposal); err != nil {
		return fmt.Errorf("proposal: %w", err)
	}
	observationDigest := cache.HashBytes(observationData).String()
	proposalDigest := cache.HashBytes(proposalData).String()
	authorizationDigest := cache.HashBytes(authorizationData).String()
	if proposal.ObservationDigest != observationDigest || proposal.ContractDigest != observation.ContractDigest || proposal.InputSourceDigest != observation.InputSourceDigest || !reflect.DeepEqual(proposal.Candidate, observation.Candidates[0]) {
		return errors.New("proposal is not bound to exact observed evidence")
	}
	if err := generation.ValidateSemanticAdoptionAuthorization(authorization); err != nil {
		return fmt.Errorf("authorization: %w", err)
	}
	if authorization.ProposalDigest != proposalDigest || authorization.CandidateStableID != proposal.Candidate.StableID || authorization.CandidateInputDigest != proposal.Candidate.InputDigest || authorization.ContractDigest != proposal.ContractDigest || authorization.InputSourceDigest != proposal.InputSourceDigest || !authorization.Authorized {
		return errors.New("authorization is not the exact authorized candidate input")
	}
	if adoption.Schema != generation.SemanticAdoptionReportSchema || adoption.Lifecycle != "AUTHORIZED_ADOPTION" || adoption.ObservationDigest != observationDigest || adoption.ProposalDigest != proposalDigest || adoption.AuthorizationDigest != authorizationDigest || !reflect.DeepEqual(adoption.Proposal, proposal) || !reflect.DeepEqual(adoption.Authorization, authorization) || adoption.IndependentDecision != "CLOSED" || adoption.IndependentReason != adoption.Evidence.Reason {
		return errors.New("adoption report is not bound to exact evidence")
	}
	decision, reason, _, err := generation.VerifySemanticAdoption(proposal, proposalDigest, authorization, authorizationDigest, adoption.Evidence)
	if err != nil || decision != "CLOSED" || reason != generation.SemanticAdoptionClosedReason {
		return fmt.Errorf("adoption independent verification = %s/%s: %v", decision, reason, err)
	}
	if err := generation.ValidateBoundSemanticAdoption(adoption.Observation, proposal, adoption.Evidence); err != nil {
		return err
	}
	return generation.VerifySemanticObservation(adoption.Observation)
}

func verifyClosedPair(certificate generation.SemanticRetentionCertificate, certificateDigest string, before, after, replay generation.SemanticRetentionResult) error {
	for name, result := range map[string]generation.SemanticRetentionResult{"before": before, "after": after, "replay": replay} {
		if err := generation.ValidateSemanticRetentionResult(result); err != nil {
			return fmt.Errorf("%s result: %w", name, err)
		}
		if result.CertificateDigest != certificateDigest || result.NormalizedIRDigest != certificate.NormalizedIRDigest || result.GeneratedOutputDigest != certificate.GeneratedOutputDigest || cache.HashBytes(result.GeneratedSource).String() != result.GeneratedOutputDigest {
			return fmt.Errorf("%s result is not bound to the certificate bytes and semantic digest", name)
		}
	}
	if before.Reason != generation.SemanticRetentionCertifiedReason || before.Metrics.SemanticOperationCount != 1 || before.Metrics.CertificateHits != 0 || before.Metrics.CertificateMisses != 1 {
		return errors.New("before invocation is not the measured one-operation certificate miss")
	}
	for name, result := range map[string]generation.SemanticRetentionResult{"after": after, "replay": replay} {
		if result.Reason != generation.SemanticRetentionHitReason || result.Metrics.SemanticOperationCount != 0 || result.Metrics.CertificateHits != 1 || result.Metrics.CertificateMisses != 0 || !result.Metrics.GeneratedBytesEqual || !result.Metrics.NormalizedSemanticEqual {
			return fmt.Errorf("%s invocation is not the measured zero-operation certificate hit", name)
		}
	}
	return nil
}

func verifyCaseResult(caseID, expected string, result generation.SemanticRetentionResult) error {
	if err := generation.ValidateSemanticRetentionResult(result); err != nil {
		return fmt.Errorf("case %s: %w", caseID, err)
	}
	if result.Decision != expected {
		return fmt.Errorf("case %s decision = %s, want %s", caseID, result.Decision, expected)
	}
	switch expected {
	case "CLOSED":
		if result.Reason != generation.SemanticRetentionHitReason || result.Metrics.CertificateHits != 1 || result.Metrics.SemanticOperationCount != 0 {
			return fmt.Errorf("case %s is not an exact certificate hit", caseID)
		}
	case "UNKNOWN":
		wantReason := generation.SemanticRetentionUnknownCertificateReason
		if caseID == "MISSING_AUTHORIZATION" {
			wantReason = generation.SemanticRetentionUnknownAuthorizationReason
		}
		if result.Reason != wantReason || result.Unknown == nil || !generation.SameSemanticRetentionUnknown(result.Unknown, generation.SemanticRetentionUnknownState(wantReason)) {
			return fmt.Errorf("case %s UNKNOWN state is not causal", caseID)
		}
	case generation.SemanticRetentionRefuted:
		if result.Reason != generation.SemanticRetentionRefutedReason || result.Unknown != nil || result.Metrics.CertificateMisses != 1 {
			return fmt.Errorf("case %s REFUTED state is not causal", caseID)
		}
	}
	return nil
}

func generateRetentionEnvelope(contract []byte, outputDir, scenarioID string, observation generation.SemanticObservation) (string, int, []byte, error) {
	run, err := generation.GenerateSemanticOperationEnvelopeWithObservation(contract, scenarioID, outputDir, observation)
	if err != nil {
		return "", 0, nil, fmt.Errorf("generate retained envelope %s: %w", scenarioID, err)
	}
	verification, err := generation.VerifySemanticOperationEnvelope(outputDir)
	if err != nil {
		return "", 0, nil, fmt.Errorf("verify retained envelope %s: %w", scenarioID, err)
	}
	if verification.ObservationDecision != observation.Decision {
		return "", 0, nil, fmt.Errorf("retained envelope %s observation decision = %s, want %s", scenarioID, verification.ObservationDecision, observation.Decision)
	}
	return verification.Decision, len(run.Artifacts), append([]byte(nil), run.Artifacts[0].Contents...), nil
}

func unknownObservation(observation generation.SemanticObservation) generation.SemanticObservation {
	result := observation
	result.Decision = "UNKNOWN"
	result.Reason = generation.SemanticObservationUnknownReason
	result.Unknown = generation.SemanticObservationUnknownState()
	result.PairEvidence = generation.SemanticObservationPairEvidence{}
	result.Metrics.BeforeOperationCount = 0
	result.Metrics.AfterOperationCount = 0
	result.Adoption = nil
	return result
}

func refutedObservation(observation generation.SemanticObservation) generation.SemanticObservation {
	result := observation
	result.Decision = "REFUTED"
	result.Reason = generation.SemanticObservationContradiction
	result.Unknown = nil
	result.PairEvidence.Contradiction = generation.SemanticObservationContradiction
	result.Adoption = nil
	return result
}

func readObservation(path string) ([]byte, generation.SemanticObservation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, generation.SemanticObservation{}, fmt.Errorf("read observation: %w", err)
	}
	var value generation.SemanticObservation
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, value, fmt.Errorf("decode observation: %w", err)
	}
	return data, value, nil
}

func readProposal(path string) ([]byte, generation.SemanticAdoptionProposal, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, generation.SemanticAdoptionProposal{}, fmt.Errorf("read proposal: %w", err)
	}
	var value generation.SemanticAdoptionProposal
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, value, fmt.Errorf("decode proposal: %w", err)
	}
	return data, value, nil
}

func readAuthorization(path string) ([]byte, generation.SemanticAdoptionAuthorization, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, generation.SemanticAdoptionAuthorization{}, fmt.Errorf("read authorization: %w", err)
	}
	var value generation.SemanticAdoptionAuthorization
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, value, fmt.Errorf("decode authorization: %w", err)
	}
	return data, value, nil
}

func readAdoption(path string) ([]byte, generation.SemanticAdoptionReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, generation.SemanticAdoptionReport{}, fmt.Errorf("read adoption: %w", err)
	}
	var value generation.SemanticAdoptionReport
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, value, fmt.Errorf("decode adoption: %w", err)
	}
	return data, value, nil
}

func readCertificate(path string) ([]byte, generation.SemanticRetentionCertificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, generation.SemanticRetentionCertificate{}, fmt.Errorf("read certificate: %w", err)
	}
	var value generation.SemanticRetentionCertificate
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, value, fmt.Errorf("decode certificate: %w", err)
	}
	return data, value, nil
}

func readResult(path string) (generation.SemanticRetentionResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return generation.SemanticRetentionResult{}, fmt.Errorf("read result: %w", err)
	}
	var value generation.SemanticRetentionResult
	if err := json.Unmarshal(data, &value); err != nil {
		return value, fmt.Errorf("decode result: %w", err)
	}
	return value, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
