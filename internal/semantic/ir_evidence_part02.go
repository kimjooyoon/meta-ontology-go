package semantic

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
