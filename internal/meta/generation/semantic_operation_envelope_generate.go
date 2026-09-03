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
	if err := prepareSemanticOperationOutput(outputDir); err != nil {
		return run, err
	}

	ir, metrics, err := buildSemanticOperationIR(source, scenarioID)
	if err != nil {
		return run, err
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

type semanticOperationManifest struct {
	Schema           string          `json:"schema"`
	ScenarioID       string          `json:"scenario_id"`
	AuthorityDigest  string          `json:"authority_digest"`
	ToolchainDigest  string          `json:"toolchain_digest"`
	SourceRevision   SourceRevision  `json:"source_revision"`
	Intent           OperationIntent `json:"intent"`
	Grant            EffectGrant     `json:"grant"`
	Activities       []string        `json:"activities"`
	ReplayIdentity   string          `json:"replay_identity"`
	ExpectedDecision string          `json:"expected_decision"`
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
