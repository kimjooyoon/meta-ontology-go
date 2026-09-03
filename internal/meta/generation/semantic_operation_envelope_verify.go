package generation

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// VerifySemanticOperationEnvelope independently re-reads all six artifacts,
// recomputes their digests, and reclassifies the evidence without using the
// generator's decision function.
func VerifySemanticOperationEnvelope(outputDir string) (SemanticOperationVerification, error) {
	var verification SemanticOperationVerification
	expectedNames := []string{
		"operation-manifest.json", "effect-requests.ndjson", "effect-results.ndjson",
		"semantic-patch.json", "operation-receipt.json", "operation-report.md",
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return verification, fmt.Errorf("read generated output: %w", err)
	}
	if len(entries) != len(expectedNames) {
		return verification, fmt.Errorf("expected six output artifacts, found %d", len(entries))
	}
	contents := make(map[string][]byte, len(expectedNames))
	for _, name := range expectedNames {
		data, readErr := os.ReadFile(filepath.Join(outputDir, name))
		if readErr != nil {
			return verification, fmt.Errorf("read %s: %w", name, readErr)
		}
		contents[name] = data
	}
	var manifest semanticOperationManifest
	var patch SemanticPatch
	var receipt SemanticOperationReceipt
	if err := json.Unmarshal(contents["operation-manifest.json"], &manifest); err != nil {
		return verification, fmt.Errorf("decode manifest: %w", err)
	}
	if err := json.Unmarshal(contents["semantic-patch.json"], &patch); err != nil {
		return verification, fmt.Errorf("decode semantic patch: %w", err)
	}
	if err := json.Unmarshal(contents["operation-receipt.json"], &receipt); err != nil {
		return verification, fmt.Errorf("decode receipt: %w", err)
	}
	if manifest.Schema != SemanticOperationEnvelopeSchema || patch.Schema != SemanticOperationEnvelopeSchema || receipt.Schema != SemanticOperationEnvelopeSchema {
		return verification, errors.New("artifact schema mismatch")
	}
	if manifest.ScenarioID == "" || manifest.ScenarioID != receipt.ScenarioID || patch.ScenarioID != receipt.ScenarioID {
		return verification, errors.New("artifact scenario mismatch")
	}
	if !sameEnvelopeActivities(manifest.Activities) || !sameEnvelopeActivities(receipt.Activities) {
		return verification, errors.New("released activity graph is not exactly eight activities")
	}
	if patch.Changed || patch.RepositoryWrites != 0 || receipt.ExternalUserUtility != "UNKNOWN" {
		return verification, errors.New("semantic patch or utility state is not fail-closed")
	}
	if receipt.ManifestDigest != envelopeDigestBytes(contents["operation-manifest.json"]) ||
		receipt.RequestDigest != envelopeDigestBytes(contents["effect-requests.ndjson"]) ||
		receipt.ResultDigest != envelopeDigestBytes(contents["effect-results.ndjson"]) ||
		receipt.SemanticPatchDigest != envelopeDigestBytes(contents["semantic-patch.json"]) {
		return verification, errors.New("artifact digest mismatch")
	}
	requests, err := decodeEnvelopeRequests(contents["effect-requests.ndjson"])
	if err != nil {
		return verification, err
	}
	results, err := decodeEnvelopeResults(contents["effect-results.ndjson"])
	if err != nil {
		return verification, err
	}
	decision, reason, err := independentlyClassifyEnvelope(manifest, requests, results, receipt.Replay)
	if err != nil {
		return verification, err
	}
	if decision != receipt.Decision.Decision || reason != receipt.Decision.Reason || manifest.ExpectedDecision != decision {
		return verification, fmt.Errorf("independent decision mismatch: got %s/%s, receipt %s/%s", decision, reason, receipt.Decision.Decision, receipt.Decision.Reason)
	}
	if receipt.Decision.Decision == "UNKNOWN" && !validEnvelopeUnknown(receipt.Decision.Unknown) {
		return verification, errors.New("unknown decision does not contain the six required fields")
	}
	if receipt.Decision.Decision != "UNKNOWN" && receipt.Decision.Unknown != nil {
		return verification, errors.New("non-unknown decision contains unknown evidence")
	}
	if receipt.Metrics.OutputArtifactFiles != 6 || receipt.Metrics.RepositoryWrites != 0 || receipt.Metrics.LocalTestExecutions != 0 {
		return verification, errors.New("metrics violate output or write contract")
	}
	receiptDigest := envelopeDigestBytes(contents["operation-receipt.json"])
	if string(contents["operation-report.md"]) != renderSemanticOperationReport(receipt, receiptDigest) {
		return verification, errors.New("report replay mismatch")
	}
	return SemanticOperationVerification{
		ScenarioID:    receipt.ScenarioID,
		Decision:      receipt.Decision.Decision,
		Reason:        receipt.Decision.Reason,
		ReceiptDigest: receiptDigest,
		Metrics:       receipt.Metrics,
	}, nil
}

func sameEnvelopeActivities(actual []string) bool {
	return len(actual) == len(semanticOperationActivities) && strings.Join(actual, "\x00") == strings.Join(semanticOperationActivities[:], "\x00")
}

func decodeEnvelopeRequests(data []byte) ([]EffectRequest, error) {
	return decodeEnvelopeLines[EffectRequest](data, "effect request")
}

func decodeEnvelopeResults(data []byte) ([]EffectResult, error) {
	return decodeEnvelopeLines[EffectResult](data, "effect result")
}

func decodeEnvelopeLines[T any](data []byte, kind string) ([]T, error) {
	if len(data) == 0 {
		return nil, nil
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	values := make([]T, 0, len(lines))
	for _, line := range lines {
		var value T
		if err := json.Unmarshal([]byte(line), &value); err != nil {
			return nil, fmt.Errorf("decode %s: %w", kind, err)
		}
		values = append(values, value)
	}
	return values, nil
}

func independentlyClassifyEnvelope(manifest semanticOperationManifest, requests []EffectRequest, results []EffectResult, replay ReplayIdentity) (string, string, error) {
	if len(requests) > 1 || len(results) > 1 {
		return "REFUTED", "MULTIPLE_RESULTS", nil
	}
	if len(requests) == 1 && len(results) == 1 {
		request, result := requests[0], results[0]
		if !independentSubset(result.Effects, manifest.Grant.Effects) {
			return "REFUTED", "EFFECT_ESCALATION", nil
		}
		if replay.Compared && replay.CurrentRequestDigest != replay.PreviousRequestDigest {
			return "REFUTED", "REPLAY_COLLISION", nil
		}
		if result.RequestID != request.RequestID || result.SourceRevision != manifest.SourceRevision.ID {
			return "UNKNOWN", "STALE", nil
		}
		return "CLOSED", "EXACT_RESULT", nil
	}
	if len(requests) == 1 && len(results) == 0 {
		return "UNKNOWN", "DIRECT_MISSING", nil
	}
	if len(requests) == 0 && len(results) == 1 && len(results[0].Effects) == 0 {
		return "CLOSED", "EXACT_RESULT", nil
	}
	return "REFUTED", "EVIDENCE_RELATION", nil
}

func independentSubset(values, allowed []string) bool {
	for _, value := range values {
		found := slices.Contains(allowed, value)
		if !found {
			return false
		}
	}
	return true
}

func validEnvelopeUnknown(unknown *EnvelopeUnknownState) bool {
	return unknown != nil && unknown.Stage != "" && unknown.Step != "" && unknown.Reason != "" &&
		unknown.UnknownClass != "" && unknown.NextOperation != "" && len(unknown.BlockedBy) > 0
}
