package domaincapability

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type ReverseObservationStatus string

const (
	ReverseObservationUnknown  ReverseObservationStatus = "UNKNOWN"
	ReverseObservationDeferred ReverseObservationStatus = "DEFERRED"
	ReverseObservationObserved ReverseObservationStatus = "OBSERVED"
	ReverseObservationMismatch ReverseObservationStatus = "MISMATCH"
)

type ReverseObservationInput struct {
	SourceDigest            string
	DeclarationDigest       string
	IRDigest                string
	GeneratedArtifactDigest string
	ObservedArtifactDigest  string
	EvidencePrefixDigest    string
}

type ReverseObservationReceipt struct {
	Status                  ReverseObservationStatus
	SourceDigest            string
	DeclarationDigest       string
	IRDigest                string
	GeneratedArtifactDigest string
	ObservedArtifactDigest  string
	EvidencePrefixDigest    string
	ObservedStage           string
	FirstMismatch           string
	MissingStage            string
	ReadOnly                bool
	ReceiptDigest           string
}

func ObserveReverseObservation(input ReverseObservationInput) ReverseObservationReceipt {
	receipt := ReverseObservationReceipt{
		Status:                  ReverseObservationUnknown,
		SourceDigest:            input.SourceDigest,
		DeclarationDigest:       input.DeclarationDigest,
		IRDigest:                input.IRDigest,
		GeneratedArtifactDigest: input.GeneratedArtifactDigest,
		ObservedArtifactDigest:  input.ObservedArtifactDigest,
		EvidencePrefixDigest:    input.EvidencePrefixDigest,
		FirstMismatch:           "source_digest",
		MissingStage:            "source_digest",
		ReadOnly:                true,
	}
	for _, stage := range []struct {
		name  string
		digest string
	}{
		{name: "source_digest", digest: receipt.SourceDigest},
		{name: "declaration_digest", digest: receipt.DeclarationDigest},
		{name: "ir_digest", digest: receipt.IRDigest},
		{name: "generated_artifact_digest", digest: receipt.GeneratedArtifactDigest},
		{name: "evidence_prefix_digest", digest: receipt.EvidencePrefixDigest},
	} {
		if !validReverseObservationDigest(stage.digest) {
			receipt.FirstMismatch = stage.name
			receipt.MissingStage = stage.name
			receipt.ReceiptDigest = receipt.digest()
			return receipt
		}
	}
	if !validReverseObservationDigest(receipt.ObservedArtifactDigest) {
		receipt.Status = ReverseObservationDeferred
		receipt.FirstMismatch = "reverse_observation"
		receipt.MissingStage = "reverse_observation"
		receipt.ReceiptDigest = receipt.digest()
		return receipt
	}
	if receipt.ObservedArtifactDigest != receipt.GeneratedArtifactDigest {
		receipt.Status = ReverseObservationMismatch
		receipt.ObservedStage = "reverse_observation"
		receipt.FirstMismatch = "reverse_observation"
		receipt.MissingStage = ""
		receipt.ReceiptDigest = receipt.digest()
		return receipt
	}
	receipt.Status = ReverseObservationObserved
	receipt.ObservedStage = "reverse_observation"
	receipt.FirstMismatch = ""
	receipt.MissingStage = ""
	receipt.ReceiptDigest = receipt.digest()
	return receipt
}

func (receipt ReverseObservationReceipt) Validate() error {
	if !receipt.ReadOnly {
		return fmt.Errorf("reverse observation receipt is not read-only")
	}
	if !validReverseObservationDigest(receipt.ReceiptDigest) {
		return fmt.Errorf("reverse observation receipt digest is invalid")
	}
	switch receipt.Status {
	case ReverseObservationUnknown:
		if receipt.FirstMismatch == "" || receipt.MissingStage == "" {
			return fmt.Errorf("unknown reverse observation must preserve its first boundary")
		}
	case ReverseObservationDeferred:
		if err := receipt.validateBoundChain(); err != nil {
			return err
		}
		if receipt.FirstMismatch != "reverse_observation" || receipt.MissingStage != "reverse_observation" || validReverseObservationDigest(receipt.ObservedArtifactDigest) {
			return fmt.Errorf("deferred reverse observation boundary is inconsistent")
		}
	case ReverseObservationObserved, ReverseObservationMismatch:
		if err := receipt.validateBoundChain(); err != nil {
			return err
		}
		if !validReverseObservationDigest(receipt.ObservedArtifactDigest) || receipt.ObservedStage != "reverse_observation" || receipt.MissingStage != "" {
			return fmt.Errorf("reverse observation result is incomplete")
		}
		if receipt.Status == ReverseObservationObserved && receipt.ObservedArtifactDigest != receipt.GeneratedArtifactDigest {
			return fmt.Errorf("observed reverse observation does not match generated artifact")
		}
		if receipt.Status == ReverseObservationMismatch && receipt.ObservedArtifactDigest == receipt.GeneratedArtifactDigest {
			return fmt.Errorf("mismatch reverse observation unexpectedly matches generated artifact")
		}
	default:
		return fmt.Errorf("reverse observation status %q is invalid", receipt.Status)
	}
	if receipt.digest() != receipt.ReceiptDigest {
		return fmt.Errorf("reverse observation receipt digest does not match")
	}
	return nil
}

func (receipt ReverseObservationReceipt) validateBoundChain() error {
	for name, value := range map[string]string{
		"source": receipt.SourceDigest,
		"declaration": receipt.DeclarationDigest,
		"ir": receipt.IRDigest,
		"generated": receipt.GeneratedArtifactDigest,
		"prefix": receipt.EvidencePrefixDigest,
	} {
		if !validReverseObservationDigest(value) {
			return fmt.Errorf("reverse observation %s digest is invalid", name)
		}
	}
	return nil
}

func validReverseObservationDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func (receipt ReverseObservationReceipt) digest() string {
	parts := []string{
		"meta-ontology-reverse-observation",
		string(receipt.Status),
		receipt.SourceDigest,
		receipt.DeclarationDigest,
		receipt.IRDigest,
		receipt.GeneratedArtifactDigest,
		receipt.ObservedArtifactDigest,
		receipt.EvidencePrefixDigest,
		receipt.ObservedStage,
		receipt.FirstMismatch,
		receipt.MissingStage,
		fmt.Sprintf("%t", receipt.ReadOnly),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "sha256:" + hex.EncodeToString(sum[:])
}
