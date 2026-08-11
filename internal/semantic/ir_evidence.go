package semantic

import (
	"fmt"
	"strings"
)

func (ir *IR) ensureEvidence() {
	if ir.evidence == nil {
		ir.evidence = make(map[ID]Evidence)
	}
}

// AddEvidence appends a normalized record. Re-adding the same record is
// idempotent; reusing its stable ID for different content is rejected.
func (ir *IR) AddEvidence(evidence Evidence) error {
	normalized, err := evidence.Normalized()
	if err != nil {
		return err
	}
	ir.ensureEvidence()
	if existing, ok := ir.evidence[normalized.ID]; ok {
		if existing.Canonical() != normalized.Canonical() {
			return fmt.Errorf("%w: %s", ErrEvidenceConflict, normalized.ID)
		}
		return nil
	}
	ir.evidence[normalized.ID] = normalized
	return nil
}

func (ir IR) Evidence() []Evidence {
	evidence := make([]Evidence, 0, len(ir.evidence))
	for _, record := range ir.evidence {
		evidence = append(evidence, record)
	}
	sortEvidence(evidence)
	return evidence
}

func (ir IR) validateEvidence() error {
	for _, evidence := range ir.Evidence() {
		if err := evidence.ValidateAgainst(ir.Graph); err != nil {
			return fmt.Errorf("%w: evidence %s: %v", ErrGraphInvalid, evidence.ID, err)
		}
	}
	return nil
}

func (ir IR) EvidenceCanonical() string {
	var b strings.Builder
	for _, evidence := range ir.Evidence() {
		b.WriteString(evidence.Canonical())
		b.WriteByte('\n')
	}
	return b.String()
}

func (ir IR) ProvenanceCanonical() string {
	var b strings.Builder
	for _, evidence := range ir.Evidence() {
		b.WriteString(evidence.ComparisonCanonical())
		b.WriteByte('\n')
	}
	return b.String()
}

func (ir IR) EvidenceHash() string {
	return StableHashString(ir.EvidenceCanonical())
}

func (ir IR) ProvenanceHash() string {
	return StableHashString(ir.ProvenanceCanonical())
}

// SemanticallyEquivalent compares compiler meaning while ignoring evidence
// producers, source locations, and other audit metadata.
func (ir IR) SemanticallyEquivalent(other IR) bool {
	return ir.SemanticCanonical() == other.SemanticCanonical()
}

// ProvenanceEquivalent compares the normalized claims emitted by two hosts.
// Exact audit identity remains available through EvidenceHash.
func (ir IR) ProvenanceEquivalent(other IR) bool {
	return ir.ProvenanceCanonical() == other.ProvenanceCanonical()
}
