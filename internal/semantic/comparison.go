package semantic

// IRComparison is a diagnostic comparison, not an authorization decision.
// A promotion policy must still be enforced by the independent Go verifier.
type IRComparison struct {
	LeftValid           bool
	RightValid          bool
	SemanticEqual       bool
	ProvenanceEqual     bool
	ExactEvidenceEqual  bool
	LeftSemanticHash    string
	RightSemanticHash   string
	LeftProvenanceHash  string
	RightProvenanceHash string
	LeftError           string
	RightError          string
}

// CompareIR compares two independently produced IR snapshots. Invalid input
// never reports equivalence, even when its raw canonical strings match.
func CompareIR(left, right IR) IRComparison {
	leftValid, leftError := comparisonValidity(left)
	rightValid, rightError := comparisonValidity(right)
	result := IRComparison{
		LeftValid:           leftValid,
		RightValid:          rightValid,
		LeftSemanticHash:    left.StableHash(),
		RightSemanticHash:   right.StableHash(),
		LeftProvenanceHash:  left.ProvenanceHash(),
		RightProvenanceHash: right.ProvenanceHash(),
		LeftError:           leftError,
		RightError:          rightError,
	}
	result.SemanticEqual = leftValid && rightValid && result.LeftSemanticHash == result.RightSemanticHash
	result.ProvenanceEqual = leftValid && rightValid && result.LeftProvenanceHash == result.RightProvenanceHash
	result.ExactEvidenceEqual = leftValid && rightValid && left.EvidenceHash() == right.EvidenceHash()
	return result
}

// Equivalent reports comparable valid meaning and provenance only. It does
// not promote a gooo-hosted verifier or establish CI authority.
func (c IRComparison) Equivalent() bool {
	return c.SemanticEqual && c.ProvenanceEqual
}

func comparisonValidity(ir IR) (bool, string) {
	if err := ir.Validate(); err != nil {
		return false, err.Error()
	}
	return true, ""
}
