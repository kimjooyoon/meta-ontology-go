package generation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func prepareSemanticOperationOutput(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create caller-owned output directory: %w", err)
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("inspect caller-owned output directory: %w", err)
	}
	if len(entries) != 0 {
		return errors.New("caller-owned output directory must be empty")
	}
	return nil
}

func encodeEnvelopeJSON(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func encodeEnvelopeLines(value any) ([]byte, error) {
	if value == nil {
		return []byte{}, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func renderSemanticOperationReport(receipt SemanticOperationReceipt, receiptDigest string) string {
	metrics, _ := json.Marshal(receipt.Metrics)
	return strings.Join([]string{
		"# Semantic operation envelope",
		"",
		"schema: " + receipt.Schema,
		"scenario_id: " + receipt.ScenarioID,
		"decision: " + receipt.Decision.Decision,
		"reason: " + receipt.Decision.Reason,
		"external_user_utility: " + receipt.ExternalUserUtility,
		"activities: 8/8",
		"artifacts: 6/6",
		"repository_writes: 0",
		"local_test_executions: 0",
		"metrics_json: " + string(metrics),
		"receipt_digest: " + receiptDigest,
		"",
	}, "\n")
}

func envelopeSubset(values, allowed []string) bool {
	set := make(map[string]struct{}, len(allowed))
	for _, value := range allowed {
		set[value] = struct{}{}
	}
	for _, value := range values {
		if _, ok := set[value]; !ok {
			return false
		}
	}
	return true
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func countPhysicalLines(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	count := 1
	for _, value := range source {
		if value == '\n' {
			count++
		}
	}
	if source[len(source)-1] == '\n' {
		count--
	}
	return count
}

func envelopeDigestJSON(value any) string {
	data, _ := json.Marshal(value)
	return envelopeDigestBytes(data)
}

func envelopeDigestString(value string) string {
	return envelopeDigestBytes([]byte(value))
}

func envelopeDigestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
