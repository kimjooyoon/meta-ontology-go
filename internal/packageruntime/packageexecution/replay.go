package packageexecution

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

const ReplayReceiptSchema = `gooo/package-source-execution-replay-receipt/v1`

type ReplayReceipt struct {
	Schema                 string       `json:"schema"`
	Scope                  string       `json:"scope"`
	Decision               string       `json:"decision"`
	Reason                 string       `json:"reason"`
	Resolution             string       `json:"resolution"`
	ArtifactDigest         string       `json:"artifact_digest,omitempty"`
	ExpectedSourceDigest   string       `json:"expected_source_digest,omitempty"`
	ObservedSourceDigest   string       `json:"observed_source_digest,omitempty"`
	ExpectedSemanticDigest string       `json:"expected_semantic_digest,omitempty"`
	ObservedSemanticDigest string       `json:"observed_semantic_digest,omitempty"`
	Replay                 *Receipt     `json:"replay,omitempty"`
	Diagnostics            []Diagnostic `json:"diagnostics"`
	Digest                 string       `json:"digest"`
}

// Replay consumes a sealed package receipt and independently executes the
// supplied clean source set. It authorizes success only when the sealed
// artifact, source identity, semantic identity, and nested execution identity
// all agree.
func Replay(request Request, artifact Receipt) ReplayReceipt {
	receipt := replayBase(artifact)
	if err := Validate(artifact); err != nil {
		return rejectReplay(receipt, "PACKAGE_ARTIFACT_INVALID", fmt.Sprintf("sealed package artifact is invalid: %v", err), "EXACT")
	}
	if artifact.Decision != "PASS" {
		return rejectReplay(receipt, "PACKAGE_ARTIFACT_NOT_REPLAYABLE", "a fail-closed package artifact cannot authorize replay", "LOWER_RESOLUTION")
	}

	observed := Execute(request)
	receipt.Replay = &observed
	receipt.ObservedSourceDigest = observed.CombinedSourceDigest
	receipt.ObservedSemanticDigest = observed.SemanticDigest
	if observed.Decision != "PASS" {
		return rejectReplay(receipt, "PACKAGE_ARTIFACT_REPLAY_EXECUTION_REJECTED", "clean-source replay did not produce a passing package receipt", "EXACT")
	}
	if artifact.CombinedSourceDigest != observed.CombinedSourceDigest {
		return rejectReplay(receipt, "PACKAGE_ARTIFACT_SOURCE_DIGEST_MISMATCH", "sealed artifact source digest differs from clean-source replay", "EXACT")
	}
	if artifact.SemanticDigest != observed.SemanticDigest {
		return rejectReplay(receipt, "PACKAGE_ARTIFACT_SEMANTIC_DIGEST_MISMATCH", "sealed artifact semantic digest differs from clean-source replay", "EXACT")
	}
	if artifact.Execution == nil || observed.Execution == nil || artifact.Execution.Digest != observed.Execution.Digest {
		return rejectReplay(receipt, "PACKAGE_ARTIFACT_EXECUTION_DIGEST_MISMATCH", "sealed artifact nested execution identity differs from clean-source replay", "EXACT")
	}

	receipt.Decision = "PASS"
	receipt.Reason = "PACKAGE_ARTIFACT_REPLAYED"
	receipt.Resolution = "EXACT"
	sealReplay(&receipt)
	return receipt
}

func replayBase(artifact Receipt) ReplayReceipt {
	return ReplayReceipt{
		Schema:                 ReplayReceiptSchema,
		Scope:                  sourceexecution.DeclarationResolutionScope,
		ArtifactDigest:         artifact.Digest,
		ExpectedSourceDigest:   artifact.CombinedSourceDigest,
		ExpectedSemanticDigest: artifact.SemanticDigest,
		Diagnostics:            []Diagnostic{},
	}
}

func rejectReplay(receipt ReplayReceipt, code, message, resolution string) ReplayReceipt {
	receipt.Decision = "FAIL_CLOSED"
	receipt.Reason = code
	receipt.Resolution = resolution
	receipt.Diagnostics = []Diagnostic{{Stage: "ARTIFACT_REPLAY", Code: code, Message: message}}
	sealReplay(&receipt)
	return receipt
}

func ValidateReplay(receipt ReplayReceipt) error {
	if receipt.Schema != ReplayReceiptSchema {
		return fmt.Errorf("packageexecution: replay schema mismatch")
	}
	if receipt.Scope != sourceexecution.DeclarationResolutionScope {
		return fmt.Errorf("packageexecution: replay scope mismatch")
	}
	if receipt.Decision != "PASS" && receipt.Decision != "FAIL_CLOSED" {
		return fmt.Errorf("packageexecution: unknown replay decision %q", receipt.Decision)
	}
	if receipt.Resolution != "EXACT" && receipt.Resolution != "LOWER_RESOLUTION" {
		return fmt.Errorf("packageexecution: unknown replay resolution %q", receipt.Resolution)
	}
	if receipt.Replay != nil {
		if err := Validate(*receipt.Replay); err != nil {
			return fmt.Errorf("packageexecution: nested replay receipt: %w", err)
		}
	}
	if receipt.Decision == "PASS" {
		if receipt.Reason != "PACKAGE_ARTIFACT_REPLAYED" || receipt.Replay == nil || receipt.Replay.Decision != "PASS" {
			return fmt.Errorf("packageexecution: passing replay contradicts nested execution")
		}
		if !strings.HasPrefix(receipt.ArtifactDigest, "sha256:") ||
			receipt.ExpectedSourceDigest == "" ||
			receipt.ExpectedSourceDigest != receipt.ObservedSourceDigest ||
			receipt.ExpectedSemanticDigest == "" ||
			receipt.ExpectedSemanticDigest != receipt.ObservedSemanticDigest ||
			len(receipt.Diagnostics) != 0 {
			return fmt.Errorf("packageexecution: passing replay is incomplete")
		}
	}
	if receipt.Decision == "FAIL_CLOSED" && len(receipt.Diagnostics) == 0 {
		return fmt.Errorf("packageexecution: failed replay has no diagnostic")
	}
	want := receipt.Digest
	sealReplay(&receipt)
	if want == "" || receipt.Digest != want {
		return fmt.Errorf("packageexecution: replay receipt digest mismatch")
	}
	return nil
}

func sealReplay(receipt *ReplayReceipt) {
	receipt.Digest = ""
	receipt.Digest = digestValue(*receipt)
}

func MarshalReplay(receipt ReplayReceipt) ([]byte, error) {
	if err := ValidateReplay(receipt); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
