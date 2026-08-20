package adapter

import (
	"fmt"
	"strings"
)

// Stage identifies the trust boundary that produced evidence.
type Stage uint8

const (
	StageGoBaseline Stage = iota
	StageDualEvidence
	StageGoooFallback
	StageGoooAuthoritative
)

// EvidenceFact is a stable-ID assertion projected from a response.
type EvidenceFact struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// EvidenceBundle is independent of producer identity and is comparison input.
type EvidenceBundle struct {
	Schema   string         `json:"schema"`
	Stage    Stage          `json:"stage"`
	Fixture  string         `json:"fixture"`
	Decision string         `json:"decision"`
	Facts    []EvidenceFact `json:"facts"`
}

// EvidenceArtifact identifies the producer while retaining a neutral bundle.
type EvidenceArtifact struct {
	Producer string         `json:"producer"`
	Bundle   EvidenceBundle `json:"bundle"`
}

// EvidenceManifest is an append-only, independently checkable digest record.
type EvidenceManifest struct {
	Schema        string `json:"schema"`
	Producer      string `json:"producer"`
	Stage         Stage  `json:"stage"`
	Fixture       string `json:"fixture"`
	Decision      string `json:"decision"`
	PayloadSHA256 string `json:"payload_sha256"`
}

// Validate checks only protocol invariants; it never promotes a result.
func (a EvidenceArtifact) Validate() error {
	if a.Producer != "go" && a.Producer != "gooo" {
		return fmt.Errorf("unsupported evidence producer %q", a.Producer)
	}
	bundle := a.Bundle
	if bundle.Schema != EvidenceSchema {
		return fmt.Errorf("unsupported evidence schema %q", bundle.Schema)
	}
	if bundle.Stage > StageGoooAuthoritative {
		return fmt.Errorf("unsupported evidence stage %d", bundle.Stage)
	}
	if strings.TrimSpace(bundle.Fixture) == "" || strings.TrimSpace(bundle.Decision) == "" {
		return fmt.Errorf("evidence fixture and decision are required")
	}
	for _, fact := range bundle.Facts {
		if strings.TrimSpace(fact.ID) == "" || strings.TrimSpace(fact.Kind) == "" {
			return fmt.Errorf("evidence facts require stable id and kind")
		}
	}
	return nil
}
