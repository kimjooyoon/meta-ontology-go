package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

type adoptionScenario struct {
	ID                  string `json:"id"`
	Decision            string `json:"decision"`
	Reason              string `json:"reason"`
	ObservationDecision string `json:"observation_decision"`
	EnvelopeDecision    string `json:"envelope_decision"`
	EnvelopeArtifacts   int    `json:"envelope_artifacts"`
}

type adoptionLoopReport struct {
	Schema                        string             `json:"schema"`
	ObservationDigest             string             `json:"observation_digest"`
	ProposalDigest                string             `json:"proposal_digest"`
	CandidateStableID             string             `json:"candidate_stable_id"`
	ScenarioDenominator           int                `json:"scenario_denominator"`
	Counts                        map[string]int     `json:"counts"`
	EnvelopeScenarioDenom         int                `json:"envelope_scenario_denominator"`
	EnvelopeArtifactsPerAdoption  int                `json:"envelope_artifacts_per_adoption_scenario"`
	EnvelopeArtifactsPerOperation int                `json:"envelope_artifacts_per_operation"`
	ReplayEqual                   bool               `json:"replay_equal"`
	Scenarios                     []adoptionScenario `json:"scenarios"`
	RepositoryWrites              int                `json:"repository_writes"`
	LocalTestExecutions           int                `json:"local_test_executions"`
}

func main() {
	contractPath := flag.String("contract", "", "observation authority .gooo")
	observationPath := flag.String("observation", "", "pre-adoption observation JSON")
	proposalPath := flag.String("proposal", "", "caller-owned adoption proposal JSON")
	authorizationPath := flag.String("authorization", "", "explicit authorization JSON")
	unknownPath := flag.String("unknown", "", "authorization-denied adoption result JSON")
	unknownAuthorizationPath := flag.String("unknown-authorization", "", "authorization-denied input JSON")
	adoptionPath := flag.String("adoption", "", "authorized adoption result JSON")
	envelopeRoot := flag.String("envelope-root", "", "caller-owned envelope output root")
	outputPath := flag.String("output", "", "caller-owned loop report JSON")
	flag.Parse()
	if *contractPath == "" || *observationPath == "" || *proposalPath == "" || *authorizationPath == "" || *unknownPath == "" || *unknownAuthorizationPath == "" || *adoptionPath == "" || *envelopeRoot == "" || *outputPath == "" {
		fail(errors.New("self-improvement-adoption requires -contract, -observation, -proposal, -authorization, -unknown, -unknown-authorization, -adoption, -envelope-root, and -output"))
	}
	if err := run(*contractPath, *observationPath, *proposalPath, *authorizationPath, *unknownPath, *unknownAuthorizationPath, *adoptionPath, *envelopeRoot, *outputPath); err != nil {
		fail(err)
	}
}

func run(contractPath, observationPath, proposalPath, authorizationPath, unknownPath, unknownAuthorizationPath, adoptionPath, envelopeRoot, outputPath string) error {
	contract, err := os.ReadFile(contractPath)
	if err != nil {
		return fmt.Errorf("read contract: %w", err)
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
	_, unknownReport, err := readReport(unknownPath)
	if err != nil {
		return err
	}
	unknownAuthorizationData, unknownAuthorization, err := readAuthorization(unknownAuthorizationPath)
	if err != nil {
		return err
	}
	_, adoptedReport, err := readReport(adoptionPath)
	if err != nil {
		return err
	}
	if err := generation.VerifySemanticObservation(observation); err != nil {
		return fmt.Errorf("pre-adoption observation: %w", err)
	}
	proposalDigest := cache.HashBytes(proposalData).String()
	authorizationDigest := cache.HashBytes(authorizationData).String()
	observationDigest := cache.HashBytes(observationData).String()
	if proposal.ObservationDigest != observationDigest {
		return errors.New("proposal does not bind the pre-adoption observation")
	}
	if err := validateAdoptionReportBinding(adoptedReport, proposal, authorization, observationDigest, proposalDigest, authorizationDigest, "AUTHORIZED_ADOPTION"); err != nil {
		return fmt.Errorf("authorized adoption report binding: %w", err)
	}
	if decision, reason, _, err := generation.VerifySemanticAdoption(proposal, proposalDigest, authorization, authorizationDigest, adoptedReport.Evidence); err != nil || decision != "CLOSED" || reason != generation.SemanticAdoptionClosedReason {
		if err != nil {
			return fmt.Errorf("authorized adoption verification: %w", err)
		}
		return fmt.Errorf("authorized adoption verification = %s/%s", decision, reason)
	}
	if err := generation.ValidateBoundSemanticAdoption(adoptedReport.Observation, proposal, adoptedReport.Evidence); err != nil {
		return err
	}
	if err := generation.VerifySemanticObservation(adoptedReport.Observation); err != nil {
		return fmt.Errorf("adopted observation: %w", err)
	}
	if unknownReport.Authorization != unknownAuthorization {
		return errors.New("unknown adoption report is not bound to its authorization input")
	}
	unknownAuthorizationDigest := cache.HashBytes(unknownAuthorizationData).String()
	if err := validateAdoptionReportBinding(unknownReport, proposal, unknownAuthorization, observationDigest, proposalDigest, unknownAuthorizationDigest, "AUTHORIZATION_REQUIRED"); err != nil {
		return fmt.Errorf("unknown adoption report binding: %w", err)
	}
	unknownDecision, unknownReason, unknownState, err := generation.VerifySemanticAdoption(proposal, proposalDigest, unknownAuthorization, unknownAuthorizationDigest, unknownReport.Evidence)
	if err != nil {
		return fmt.Errorf("unknown adoption verification: %w", err)
	}
	if unknownDecision != "UNKNOWN" || unknownReason != generation.SemanticAdoptionUnknownReason || !generation.SameAdoptionUnknown(unknownState, generation.AdoptionUnknownState()) {
		return fmt.Errorf("unknown adoption verification = %s/%s", unknownDecision, unknownReason)
	}
	if err := generation.VerifySemanticObservation(unknownReport.Observation); err != nil {
		return fmt.Errorf("unknown observation: %w", err)
	}
	refutedEvidence := adoptedReport.Evidence
	refutedEvidence.BehaviorEqual = false
	refutedEvidence.Decision = generation.SemanticAdoptionRefuted
	refutedEvidence.Reason = generation.SemanticAdoptionRefutedReason
	refutedDecision, refutedReason, _, err := generation.VerifySemanticAdoption(proposal, proposalDigest, authorization, authorizationDigest, refutedEvidence)
	if err != nil {
		return fmt.Errorf("refuted adoption verification: %w", err)
	}
	if refutedDecision != generation.SemanticAdoptionRefuted || refutedReason != generation.SemanticAdoptionRefutedReason {
		return fmt.Errorf("refuted adoption verification = %s/%s", refutedDecision, refutedReason)
	}
	refutedObservation := observation
	refutedObservation.Decision = "REFUTED"
	refutedObservation.Reason = generation.SemanticObservationContradiction
	refutedObservation.Unknown = nil
	refutedObservation.PairEvidence.Contradiction = generation.SemanticObservationContradiction
	refutedObservation.Adoption = nil
	if err := generation.VerifySemanticObservation(refutedObservation); err != nil {
		return fmt.Errorf("refuted observation: %w", err)
	}
	if err := os.MkdirAll(envelopeRoot, 0o755); err != nil {
		return fmt.Errorf("create envelope root: %w", err)
	}
	observations := map[string]generation.SemanticObservation{
		"NORMAL":  adoptedReport.Observation,
		"UNKNOWN": unknownReport.Observation,
		"REFUTED": refutedObservation,
		"REPLAY":  adoptedReport.Observation,
	}
	scenarios := make([]adoptionScenario, 0, len(observations))
	counts := map[string]int{"CLOSED": 0, "UNKNOWN": 0, "REFUTED": 0}
	var replayBytes []byte
	var replayEqual = true
	for _, id := range []string{"NORMAL", "UNKNOWN", "REFUTED", "REPLAY"} {
		envelopeDecision, artifacts, firstBytes, err := generateEnvelopeSet(contract, filepath.Join(envelopeRoot, id), id, observations[id])
		if err != nil {
			return err
		}
		if id == "NORMAL" {
			replayBytes = firstBytes
		} else if id == "REPLAY" {
			replayEqual = bytesEqual(replayBytes, firstBytes)
		}
		decision, reason := adoptionScenarioDecision(id, adoptedReport, unknownReport, refutedDecision, refutedReason)
		if envelopeDecision != observations[id].Decision {
			return fmt.Errorf("envelope %s observation decision = %s, want %s", id, envelopeDecision, observations[id].Decision)
		}
		counts[decision]++
		scenarios = append(scenarios, adoptionScenario{ID: id, Decision: decision, Reason: reason, ObservationDecision: observations[id].Decision, EnvelopeDecision: envelopeDecision, EnvelopeArtifacts: artifacts})
	}
	if !replayEqual {
		return errors.New("adoption envelope replay changed bytes")
	}
	report := adoptionLoopReport{
		Schema: generation.SemanticAdoptionReportSchema, ObservationDigest: cache.HashBytes(observationData).String(),
		ProposalDigest: proposalDigest, CandidateStableID: proposal.Candidate.StableID, ScenarioDenominator: 4,
		Counts: counts, EnvelopeScenarioDenom: len(generation.SemanticOperationScenarioIDs()), EnvelopeArtifactsPerAdoption: len(generation.SemanticOperationScenarioIDs()) * 6, EnvelopeArtifactsPerOperation: 6,
		ReplayEqual: replayEqual, Scenarios: scenarios, RepositoryWrites: 0, LocalTestExecutions: 0,
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode adoption loop report: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("write adoption loop report: %w", err)
	}
	return nil
}

func generateEnvelopeSet(contract []byte, root, scenarioID string, observation generation.SemanticObservation) (string, int, []byte, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", 0, nil, err
	}
	var first []byte
	decision := ""
	for _, envelopeID := range generation.SemanticOperationScenarioIDs() {
		directory := filepath.Join(root, envelopeID)
		run, err := generation.GenerateSemanticOperationEnvelopeWithObservation(contract, envelopeID, directory, observation)
		if err != nil {
			return "", 0, nil, fmt.Errorf("generate %s/%s envelope: %w", scenarioID, envelopeID, err)
		}
		verification, err := generation.VerifySemanticOperationEnvelope(directory)
		if err != nil {
			return "", 0, nil, fmt.Errorf("verify %s/%s envelope: %w", scenarioID, envelopeID, err)
		}
		if verification.ObservationDecision != observation.Decision {
			return "", 0, nil, fmt.Errorf("verify %s/%s observation decision = %s, want %s", scenarioID, envelopeID, verification.ObservationDecision, observation.Decision)
		}
		if envelopeID == "C1" {
			first = append([]byte(nil), run.Artifacts[0].Contents...)
			decision = verification.ObservationDecision
		}
	}
	return decision, len(generation.SemanticOperationScenarioIDs()) * 6, first, nil
}

func adoptionScenarioDecision(id string, adopted, unknown generation.SemanticAdoptionReport, refutedDecision, refutedReason string) (string, string) {
	switch id {
	case "UNKNOWN":
		return unknown.IndependentDecision, unknown.IndependentReason
	case "REFUTED":
		return refutedDecision, refutedReason
	default:
		return adopted.IndependentDecision, adopted.IndependentReason
	}
}

func validateAdoptionReportBinding(report generation.SemanticAdoptionReport, proposal generation.SemanticAdoptionProposal, authorization generation.SemanticAdoptionAuthorization, observationDigest, proposalDigest, authorizationDigest, lifecycle string) error {
	if report.Schema != generation.SemanticAdoptionReportSchema || report.Lifecycle != lifecycle || report.ObservationDigest != observationDigest ||
		report.ProposalDigest != proposalDigest || report.AuthorizationDigest != authorizationDigest || !reflect.DeepEqual(report.Proposal, proposal) ||
		!reflect.DeepEqual(report.Authorization, authorization) || report.IndependentDecision != report.Evidence.Decision || report.IndependentReason != report.Evidence.Reason {
		return errors.New("report is not bound to the exact proposal, authorization, and observation")
	}
	if authorization.Authorized {
		if report.Observation.Adoption == nil || !reflect.DeepEqual(*report.Observation.Adoption, report.Evidence) {
			return errors.New("authorized report does not bind adoption evidence into its observation")
		}
	} else if report.Observation.Adoption != nil {
		return errors.New("authorization-denied report must not claim adoption evidence")
	}
	return nil
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

func readReport(path string) ([]byte, generation.SemanticAdoptionReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, generation.SemanticAdoptionReport{}, fmt.Errorf("read adoption report: %w", err)
	}
	var value generation.SemanticAdoptionReport
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, value, fmt.Errorf("decode adoption report: %w", err)
	}
	return data, value, nil
}

func bytesEqual(left, right []byte) bool {
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

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
