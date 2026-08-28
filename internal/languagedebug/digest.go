package languagedebug

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func seal(receipt Receipt) Receipt {
	receipt.Digest = ""
	data, _ := json.Marshal(receipt)
	sum := sha256.Sum256(data)
	receipt.Digest = "sha256:" + hex.EncodeToString(sum[:])
	return receipt
}

const DeterministicPayloadSchema = "gooo/language-debug-deterministic-payload/v1"

// DeterministicPayload is the semantic part of a debug receipt. Breakpoint
// projection and runtime fields intentionally do not participate in it.
type DeterministicPayload struct {
	Schema           string            `json:"schema"`
	Filename         string            `json:"filename"`
	SourceDigest     string            `json:"source_digest"`
	SemanticDigest   string            `json:"semantic_digest"`
	ExecutionDigest  string            `json:"execution_digest"`
	Entry            json.RawMessage   `json:"entry"`
	Diagnostics      []json.RawMessage `json:"diagnostics"`
	Effects          Effects           `json:"effects"`
	NonClaims        []string          `json:"non_claims"`
}

func (receipt Receipt) DeterministicPayload() DeterministicPayload {
	return DeterministicPayload{
		Schema: DeterministicPayloadSchema, Filename: receipt.Filename,
		SourceDigest: receipt.SourceDigest, SemanticDigest: receipt.SemanticDigest,
		ExecutionDigest: receipt.ExecutionDigest, Entry: receipt.Entry,
		Diagnostics: receipt.Diagnostics, Effects: receipt.Effects,
		NonClaims: receipt.NonClaims,
	}
}

func (receipt Receipt) DeterministicDigest() string {
	data, _ := json.Marshal(receipt.DeterministicPayload())
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DeterministicExcludedFields() []string {
	return []string{"decision", "reason", "resolution", "state", "breakpoint", "current_event", "trace", "remaining_events", "digest"}
}

func validDigest(value string) bool {
	if len(value) != 71 || value[:7] != "sha256:" {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil
}
