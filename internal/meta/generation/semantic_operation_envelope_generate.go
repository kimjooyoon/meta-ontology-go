package generation

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateSemanticOperationEnvelope parses the .gooo authority, constructs the
// semantic IR, and writes exactly six artifacts to the caller-owned directory.
func GenerateSemanticOperationEnvelope(source []byte, scenarioID, outputDir string) (SemanticOperationRun, error) {
	return generateSemanticOperationEnvelope(source, scenarioID, outputDir, nil)
}

// GenerateSemanticOperationEnvelopeWithObservation binds a compiler-produced
// self-observation to the existing semantic operation envelope. The observer
// is evidence only: no repository change is applied here.
func GenerateSemanticOperationEnvelopeWithObservation(source []byte, scenarioID, outputDir string, observation SemanticObservation) (SemanticOperationRun, error) {
	return generateSemanticOperationEnvelope(source, scenarioID, outputDir, &observation)
}

func generateSemanticOperationEnvelope(source []byte, scenarioID, outputDir string, observation *SemanticObservation) (SemanticOperationRun, error) {
	var run SemanticOperationRun
	if len(source) == 0 {
		return run, errors.New(".gooo authority is empty")
	}
	if outputDir == "" {
		return run, errors.New("caller-owned output directory is required")
	}
	if err := validateSemanticOperationAuthority(source); err != nil {
		return run, err
	}
	if observation != nil {
		if err := validateBoundSemanticObservation(source, *observation); err != nil {
			return run, err
		}
	}
	if err := prepareSemanticOperationOutput(outputDir); err != nil {
		return run, err
	}

	ir, metrics, err := buildSemanticOperationIR(source, scenarioID)
	if err != nil {
		return run, err
	}
	ir.Observation = observation
	if observation != nil {
		metrics = mergeSemanticObservationMetrics(metrics, observation.Metrics)
	}
	manifest := semanticOperationManifest{
		Schema:           SemanticOperationEnvelopeSchema,
		ScenarioID:       scenarioID,
		AuthorityDigest:  ir.AuthorityDigest,
		ToolchainDigest:  ir.ToolchainDigest,
		SourceRevision:   ir.Source,
		Intent:           ir.Intent,
		Grant:            ir.Grant,
		Activities:       append([]string(nil), ir.Activities...),
		ReplayIdentity:   ir.Replay.Identity,
		ExpectedDecision: ir.Decision.Decision,
	}
	if observation != nil {
		manifest.ObservationDigest = envelopeDigestJSON(*observation)
	}
	requestBytes := []byte{}
	if ir.Request != nil {
		requestBytes, err = encodeEnvelopeLines(*ir.Request)
		if err != nil {
			return run, err
		}
	}
	resultBytes := []byte{}
	if ir.Result != nil {
		resultBytes, err = encodeEnvelopeLines(*ir.Result)
		if err != nil {
			return run, err
		}
	}
	manifestBytes, err := encodeEnvelopeJSON(manifest)
	if err != nil {
		return run, err
	}
	patch := SemanticPatch{
		Schema:           SemanticOperationEnvelopeSchema,
		ScenarioID:       scenarioID,
		Changed:          false,
		Operations:       []string{},
		RepositoryWrites: 0,
		Observation:      observation,
	}
	patchBytes, err := encodeEnvelopeJSON(patch)
	if err != nil {
		return run, err
	}
	receipt := SemanticOperationReceipt{
		Schema:              SemanticOperationEnvelopeSchema,
		ScenarioID:          scenarioID,
		AuthorityDigest:     ir.AuthorityDigest,
		SourceRevision:      ir.Source,
		Decision:            ir.Decision,
		Replay:              ir.Replay,
		Activities:          append([]string(nil), ir.Activities...),
		ManifestDigest:      envelopeDigestBytes(manifestBytes),
		RequestDigest:       envelopeDigestBytes(requestBytes),
		ResultDigest:        envelopeDigestBytes(resultBytes),
		SemanticPatchDigest: envelopeDigestBytes(patchBytes),
		Metrics:             metrics,
		ExternalUserUtility: "UNKNOWN",
		Observation:         observation,
	}
	receiptBytes, err := encodeEnvelopeJSON(receipt)
	if err != nil {
		return run, err
	}
	receiptDigest := envelopeDigestBytes(receiptBytes)
	reportBytes := []byte(renderSemanticOperationReport(receipt, receiptDigest))

	artifacts := []SemanticOperationArtifact{
		{Name: "operation-manifest.json", Contents: manifestBytes},
		{Name: "effect-requests.ndjson", Contents: requestBytes},
		{Name: "effect-results.ndjson", Contents: resultBytes},
		{Name: "semantic-patch.json", Contents: patchBytes},
		{Name: "operation-receipt.json", Contents: receiptBytes},
		{Name: "operation-report.md", Contents: reportBytes},
	}
	for _, artifact := range artifacts {
		if err := os.WriteFile(filepath.Join(outputDir, artifact.Name), artifact.Contents, 0o644); err != nil {
			return run, fmt.Errorf("write %s: %w", artifact.Name, err)
		}
	}
	return SemanticOperationRun{IR: ir, Receipt: receipt, Artifacts: artifacts, ReceiptDigest: receiptDigest}, nil
}

func mergeSemanticObservationMetrics(metrics EnvelopeMetrics, observation SemanticObservationMetrics) EnvelopeMetrics {
	metrics.ObservedOperations = observation.ObservedOperations
	metrics.DistinctInputDigests = observation.DistinctInputDigests
	metrics.DuplicateEvaluations = observation.DuplicateEvaluations
	metrics.CandidatesEmitted = observation.CandidatesEmitted
	metrics.BeforeOperationCount = observation.BeforeOperationCount
	metrics.AfterOperationCount = observation.AfterOperationCount
	metrics.AllocationCount = observation.AllocationCount
	metrics.AllocationBytes = observation.AllocationBytes
	metrics.WallMS = int(observation.WallMS)
	metrics.PeakRSSKib = int(observation.PeakRSSKib)
	metrics.BuildMS = observation.BuildMS
	metrics.TestMS = observation.TestMS
	metrics.ExecutedTests = observation.ExecutedTests
	metrics.ReusedTests = observation.ReusedTests
	return metrics
}

type semanticOperationManifest struct {
	Schema            string          `json:"schema"`
	ScenarioID        string          `json:"scenario_id"`
	AuthorityDigest   string          `json:"authority_digest"`
	ToolchainDigest   string          `json:"toolchain_digest"`
	SourceRevision    SourceRevision  `json:"source_revision"`
	Intent            OperationIntent `json:"intent"`
	Grant             EffectGrant     `json:"grant"`
	Activities        []string        `json:"activities"`
	ReplayIdentity    string          `json:"replay_identity"`
	ExpectedDecision  string          `json:"expected_decision"`
	ObservationDigest string          `json:"observation_digest,omitempty"`
}

func validateBoundSemanticObservation(source []byte, observation SemanticObservation) error {
	contract, err := ParseSemanticObservationContract(source)
	if err != nil {
		return err
	}
	if observation.Schema != SemanticObservationSchema {
		return fmt.Errorf("semantic observation schema mismatch: %q", observation.Schema)
	}
	if observation.ContractDigest != envelopeDigestBytes(source) {
		return errors.New("semantic observation contract digest does not match authority")
	}
	if observation.Contract.Activity != contract.Activity ||
		observation.Contract.Phase != contract.Phase ||
		observation.Contract.OperationID != contract.OperationID ||
		observation.Contract.CanonicalInputIdentity != contract.CanonicalInputIdentity ||
		observation.Contract.Pure != contract.Pure ||
		observation.Contract.CandidateRule != contract.CandidateRule ||
		strings.Join(observation.Contract.AllowedEffects, "\x00") != strings.Join(contract.AllowedEffects, "\x00") {
		return errors.New("semantic observation contract does not match authority")
	}
	if observation.Metrics.RepositoryWrites != 0 || observation.Metrics.LocalTestExecutions != 0 {
		return errors.New("semantic observation metrics violate the no-write/no-local-test contract")
	}
	return nil
}

func validateSemanticOperationAuthority(source []byte) error {
	seen := make(map[string]int, len(semanticOperationActivities))
	scanner := bufio.NewScanner(strings.NewReader(string(source)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "activity ") {
			continue
		}
		name := strings.TrimSpace(strings.TrimPrefix(line, "activity "))
		if cut := strings.IndexByte(name, '('); cut >= 0 {
			name = name[:cut]
		}
		if _, ok := semanticOperationActivityIndex(name); !ok {
			return fmt.Errorf("unrecognized operation envelope activity %q", name)
		}
		seen[name]++
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan .gooo authority: %w", err)
	}
	for _, name := range semanticOperationActivities {
		if seen[name] != 1 {
			return fmt.Errorf("activity %s appears %d times; expected once", name, seen[name])
		}
	}
	if len(seen) != len(semanticOperationActivities) {
		return errors.New(".gooo authority does not bind exactly eight activities")
	}
	return nil
}

func semanticOperationActivityIndex(name string) (int, bool) {
	for index, expected := range semanticOperationActivities {
		if name == expected {
			return index, true
		}
	}
	return 0, false
}

type semanticOperationScenarioPlan struct {
	ID                    string
	RequestedEffects      []string
	GrantedEffects        []string
	ResultEffects         []string
	ResultSourceRevision  string
	ResultPresent         bool
	ReplayCompared        bool
	PreviousRequestDigest string
}

func semanticOperationScenario(id string) (semanticOperationScenarioPlan, error) {
	plan := semanticOperationScenarioPlan{ID: id, ResultSourceRevision: "source-r1"}
	switch id {
	case "C1":
		plan.ResultPresent = true
	case "C2":
		plan.RequestedEffects = []string{"read:source"}
		plan.GrantedEffects = []string{"read:source"}
		plan.ResultEffects = []string{"read:source"}
		plan.ResultPresent = true
		plan.ReplayCompared = true
	case "U1":
		plan.RequestedEffects = []string{"read:source"}
		plan.GrantedEffects = []string{"read:source"}
	case "U2":
		plan.RequestedEffects = []string{"read:source"}
		plan.GrantedEffects = []string{"read:source"}
		plan.ResultEffects = []string{"read:source"}
		plan.ResultSourceRevision = "source-r0"
		plan.ResultPresent = true
	case "R1":
		plan.RequestedEffects = []string{"write:repository"}
		plan.GrantedEffects = []string{"read:source"}
		plan.ResultEffects = []string{"write:repository"}
		plan.ResultPresent = true
	case "R2":
		plan.RequestedEffects = []string{"read:source"}
		plan.GrantedEffects = []string{"read:source"}
		plan.ResultEffects = []string{"read:source"}
		plan.ResultPresent = true
		plan.ReplayCompared = true
		plan.PreviousRequestDigest = envelopeDigestString("different-canonical-request")
	default:
		return semanticOperationScenarioPlan{}, fmt.Errorf("unknown semantic operation scenario %q", id)
	}
	return plan, nil
}
