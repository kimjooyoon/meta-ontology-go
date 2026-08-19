package eligibilityregistry

import (
	"crypto/sha256"
)

func itemKindSpelling(kind ItemKind) []byte {
	switch kind {
	case ItemSemantic:
		return []byte("SEMANTIC")
	case ItemStructural:
		return []byte("STRUCTURAL")
	default:
		return nil
	}
}
func authorityKindSpelling(kind AuthorityKind) []byte {
	switch kind {
	case AuthorityBusinessDSL:
		return []byte("BUSINESS_DSL")
	case AuthoritySemanticIR:
		return []byte("SEMANTIC_IR")
	default:
		return nil
	}
}
func projectionKindSpelling(kind ProjectionKind) []byte {
	switch kind {
	case ProjectionSemanticIR:
		return []byte("SEMANTIC_IR")
	case ProjectionGeneratedGo:
		return []byte("GENERATED_GO")
	default:
		return nil
	}
}
func digestBytes(payload []byte) Digest {
	sum := sha256.Sum256(payload)
	encoded := make([]byte, 71)
	copy(encoded, "sha256:")
	const digits = "0123456789abcdef"
	for index, value := range sum {
		encoded[7+index*2] = digits[value>>4]
		encoded[8+index*2] = digits[value&0x0F]
	}
	return Digest(encoded)
}
