package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementobservation"
)

const loopSchema = "gooo/compiler-self-observation-loop/v1"

type loopScenario struct {
	ID                  string `json:"id"`
	GeneratedDecision   string `json:"generated_decision"`
	VerifiedDecision    string `json:"verified_decision"`
	ObservationDecision string `json:"observation_decision"`
	ObservationReason   string `json:"observation_reason"`
	ReceiptDigest       string `json:"receipt_digest"`
	Artifacts           int    `json:"artifacts"`
}

type loopReport struct {
	Schema               string                                `json:"schema"`
	AuthorityProgram     string                                `json:"authority_program"`
	AuthorityDigest      string                                `json:"authority_digest"`
	ObservationDigest    string                                `json:"observation_digest"`
	ObservedOperations   int                                   `json:"observed_operations"`
	DistinctInputDigests int                                   `json:"distinct_input_digests"`
	DuplicateEvaluations int                                   `json:"duplicate_evaluations"`
	CandidatesEmitted    int                                   `json:"candidates_emitted"`
	ObservationMetrics   generation.SemanticObservationMetrics `json:"observation_metrics"`
	Scenarios            []loopScenario                        `json:"scenarios"`
	Counts               map[string]int                        `json:"counts"`
	ScenarioDenominator  int                                   `json:"scenario_denominator"`
	ArtifactsPerScenario int                                   `json:"artifacts_per_scenario"`
	RepositoryWrites     int                                   `json:"repository_writes"`
	LocalTestExecutions  int                                   `json:"local_test_executions"`
}

func main() {
	mode := flag.String("mode", "language", "language or compiler")
	head := flag.String("head-sha", "", "exact language experiment commit")
	runID := flag.Int64("source-run-id", 0, "language experiment workflow run ID")
	languageReport := flag.String("report", "", "language experiment report JSON")
	counterexamples := flag.String("counterexamples", "", "language counterexample summary JSON")
	languageContract := flag.String("contract", "", "compiled Gooo self-improvement contract JSON")
	output := flag.String("output", "", "observation JSON output path; stdout when empty")
	check := flag.Bool("check", false, "exit non-zero unless the input is exactly observed")
	compilerContract := flag.String("compiler-contract", "", "compiler observation authority .gooo")
	compilerObservation := flag.String("compiler-observation", "", "compiler observation JSON")
	envelopeRoot := flag.String("envelope-root", "", "caller-owned envelope output directory")
	compilerOutput := flag.String("compiler-output", "", "caller-owned compiler loop report JSON")
	flag.Parse()
	if *mode == "compiler" {
		if *compilerContract == "" || *compilerObservation == "" || *envelopeRoot == "" || *compilerOutput == "" {
			exitError(errors.New("usage: self-improvement-observation -mode compiler -compiler-contract FILE -compiler-observation FILE -envelope-root DIR -compiler-output FILE"))
		}
		if err := runCompilerObservationLoop(*compilerContract, *compilerObservation, *envelopeRoot, *compilerOutput); err != nil {
			exitError(err)
		}
		return
	}
	if *mode != "language" {
		exitError(errors.New("mode must be language or compiler"))
	}
	in, err := selfimprovementobservation.LoadInputs(*languageReport, *counterexamples, *languageContract)
	if err != nil {
		exitError(err)
	}
	observation := selfimprovementobservation.Build(in, selfimprovementobservation.Options{HeadSHA: *head, SourceRunID: *runID})
	data, err := json.MarshalIndent(observation, "", "  ")
	if err != nil {
		exitError(err)
	}
	data = append(data, '\n')
	if *output == "" {
		_, err = os.Stdout.Write(data)
	} else {
		err = os.WriteFile(*output, data, 0o644)
	}
	if err != nil {
		exitError(err)
	}
	if *check && observation.Decision != "OBSERVED" {
		os.Exit(1)
	}
}

func runCompilerObservationLoop(contractPath, observationPath, envelopeRoot, reportPath string) error {
	contractSource, err := os.ReadFile(contractPath)
	if err != nil {
		return err
	}
	observationData, err := os.ReadFile(observationPath)
	if err != nil {
		return err
	}
	var observation generation.SemanticObservation
	if err := json.Unmarshal(observationData, &observation); err != nil {
		return fmt.Errorf("decode compiler observation: %w", err)
	}
	if err := ensureEmptyDirectory(envelopeRoot); err != nil {
		return err
	}

	observations := map[string]generation.SemanticObservation{
		"NORMAL":  observation,
		"UNKNOWN": unknownObservation(observation),
		"REFUTED": refutedObservation(observation),
		"REPLAY":  observation,
	}
	order := generation.SemanticObservationScenarioIDs()
	report := loopReport{
		Schema:               loopSchema,
		AuthorityProgram:     contractPath,
		AuthorityDigest:      observation.ContractDigest,
		ObservationDigest:    digestJSON(observation),
		ObservedOperations:   observation.ObservedOperations,
		DistinctInputDigests: observation.DistinctInputDigests,
		DuplicateEvaluations: observation.DuplicateEvaluations,
		CandidatesEmitted:    observation.CandidatesEmitted,
		ObservationMetrics:   observation.Metrics,
		Counts:               map[string]int{"CLOSED": 0, "UNKNOWN": 0, "REFUTED": 0},
		ScenarioDenominator:  len(order),
		ArtifactsPerScenario: 6,
		RepositoryWrites:     0,
		LocalTestExecutions:  0,
	}
	for _, scenarioID := range order {
		outputDir := filepath.Join(envelopeRoot, strings.ToLower(scenarioID))
		run, err := generation.GenerateSemanticOperationEnvelopeWithObservation(contractSource, "C2", outputDir, observations[scenarioID])
		if err != nil {
			return fmt.Errorf("generate %s envelope: %w", scenarioID, err)
		}
		verified, err := generation.VerifySemanticOperationEnvelope(outputDir)
		if err != nil {
			return fmt.Errorf("verify %s envelope: %w", scenarioID, err)
		}
		if verified.Decision != "CLOSED" || verified.ObservationDecision != observations[scenarioID].Decision || verified.ObservationReason != observations[scenarioID].Reason {
			return fmt.Errorf("%s verifier decision mismatch", scenarioID)
		}
		report.Counts[verified.ObservationDecision]++
		report.Scenarios = append(report.Scenarios, loopScenario{
			ID:                  scenarioID,
			GeneratedDecision:   run.Receipt.Observation.Decision,
			VerifiedDecision:    verified.ObservationDecision,
			ObservationDecision: verified.ObservationDecision,
			ObservationReason:   verified.ObservationReason,
			ReceiptDigest:       verified.ReceiptDigest,
			Artifacts:           len(run.Artifacts),
		})
	}
	if report.Counts["CLOSED"] != 2 || report.Counts["UNKNOWN"] != 1 || report.Counts["REFUTED"] != 1 {
		return fmt.Errorf("observation scenario denominator is not CLOSED/UNKNOWN/REFUTED = 2/1/1: %v", report.Counts)
	}
	if err := verifyDeterministicReplay(contractSource, observations, order, envelopeRoot); err != nil {
		return err
	}
	reportData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(reportPath, append(reportData, '\n'), 0o644); err != nil {
		return err
	}
	return nil
}

func unknownObservation(observation generation.SemanticObservation) generation.SemanticObservation {
	observation.Decision = "UNKNOWN"
	observation.Reason = generation.SemanticObservationUnknownReason
	observation.Unknown = &generation.EnvelopeUnknownState{
		Stage:         generation.SemanticObservationUnknownStage,
		Step:          generation.SemanticObservationUnknownStep,
		Reason:        generation.SemanticObservationUnknownReason,
		UnknownClass:  generation.SemanticObservationUnknownClass,
		NextOperation: generation.SemanticObservationUnknownNext,
		BlockedBy:     []string{"behavior_determinism_pair"},
	}
	observation.PairEvidence.EvidenceAvailable = false
	return observation
}

func refutedObservation(observation generation.SemanticObservation) generation.SemanticObservation {
	observation.Decision = "REFUTED"
	observation.Reason = generation.SemanticObservationContradiction
	observation.PairEvidence.Contradiction = "fixed scenario contradiction: declared evidence is marked contradictory"
	return observation
}

func verifyDeterministicReplay(contractSource []byte, observations map[string]generation.SemanticObservation, order []string, firstRoot string) error {
	secondRoot, err := os.MkdirTemp("", "gooo-self-observation-replay-")
	if err != nil {
		return fmt.Errorf("create replay output: %w", err)
	}
	defer os.RemoveAll(secondRoot)
	for _, scenarioID := range order {
		firstDir := filepath.Join(firstRoot, strings.ToLower(scenarioID))
		secondDir := filepath.Join(secondRoot, strings.ToLower(scenarioID))
		if _, err := generation.GenerateSemanticOperationEnvelopeWithObservation(contractSource, "C2", secondDir, observations[scenarioID]); err != nil {
			return fmt.Errorf("replay generate %s envelope: %w", scenarioID, err)
		}
		firstEntries, err := os.ReadDir(firstDir)
		if err != nil {
			return err
		}
		secondEntries, err := os.ReadDir(secondDir)
		if err != nil {
			return err
		}
		if len(firstEntries) != len(secondEntries) {
			return fmt.Errorf("replay artifact count mismatch for %s", scenarioID)
		}
		for _, entry := range firstEntries {
			first, err := os.ReadFile(filepath.Join(firstDir, entry.Name()))
			if err != nil {
				return err
			}
			second, err := os.ReadFile(filepath.Join(secondDir, entry.Name()))
			if err != nil {
				return err
			}
			if !bytes.Equal(first, second) {
				return fmt.Errorf("replay artifact %s/%s is not deterministic", scenarioID, entry.Name())
			}
		}
	}
	return nil
}

func ensureEmptyDirectory(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("caller-owned envelope directory must be empty")
	}
	return nil
}

func digestJSON(value any) string {
	data, _ := json.Marshal(value)
	return fmt.Sprintf("sha256:%x", sha256Sum(data))
}

func sha256Sum(data []byte) [32]byte {
	return sha256.Sum256(data)
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
