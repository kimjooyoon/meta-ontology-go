package verify

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func (bundle EvidenceBundle) normalized() (EvidenceBundle, error) {
	if bundle.Schema != EvidenceSchemaVersion {
		return EvidenceBundle{}, fmt.Errorf("unsupported evidence schema %q", bundle.Schema)
	}
	if bundle.Stage > StageGoooAuthoritative {
		return EvidenceBundle{}, fmt.Errorf("unsupported conformance stage %d", bundle.Stage)
	}
	if strings.TrimSpace(bundle.Fixture) == "" || strings.TrimSpace(bundle.Decision) == "" {
		return EvidenceBundle{}, fmt.Errorf("evidence fixture and decision are required")
	}
	facts := append([]EvidenceFact(nil), bundle.Facts...)
	for _, fact := range facts {
		if strings.TrimSpace(fact.ID) == "" || strings.TrimSpace(fact.Kind) == "" {
			return EvidenceBundle{}, fmt.Errorf("evidence facts require stable id and kind")
		}
	}
	sort.Slice(facts, func(i, j int) bool {
		if facts[i].ID != facts[j].ID {
			return facts[i].ID < facts[j].ID
		}
		if facts[i].Kind != facts[j].Kind {
			return facts[i].Kind < facts[j].Kind
		}
		return facts[i].Value < facts[j].Value
	})
	for i := 1; i < len(facts); i++ {
		if facts[i].ID == facts[i-1].ID {
			return EvidenceBundle{}, fmt.Errorf("duplicate evidence fact id %q", facts[i].ID)
		}
	}
	if facts == nil {
		facts = []EvidenceFact{}
	}
	bundle.Facts = facts
	return bundle, nil
}
func validateEvidenceProducer(producer string) error {
	if producer != EvidenceProducerGo && producer != EvidenceProducerGooo {
		return fmt.Errorf("unsupported evidence producer %q", producer)
	}
	return nil
}

// CanonicalPayload returns producer-independent JSONL suitable for comparison.
func (artifact EvidenceArtifact) CanonicalPayload() ([]byte, error) {
	if err := validateEvidenceProducer(artifact.Producer); err != nil {
		return nil, err
	}
	bundle, err := artifact.Bundle.normalized()
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(bundle)
	if err != nil {
		return nil, fmt.Errorf("marshal evidence bundle: %w", err)
	}
	return append(payload, '\n'), nil
}
