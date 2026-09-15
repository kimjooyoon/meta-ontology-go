package bindingcoverage

import (
	"strings"
)

func validateWireToken(value, prefix string) Reason {
	if value == "" {
		return ReasonMissingInput
	}
	if !strings.HasPrefix(value, prefix) {
		return ReasonInvalidToken
	}
	token := value[len(prefix):]
	if token == "" || token[0] == '-' || token[len(token)-1] == '-' {
		return ReasonInvalidToken
	}
	for index := 0; index < len(token); index++ {
		char := token[index]
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
			return ReasonInvalidToken
		}
		if index > 0 && token[index-1] == '-' && char == '-' {
			return ReasonInvalidToken
		}
	}
	return ""
}
func validKind(kind BindingKind) bool {
	return kind == KindExactValue || kind == KindExactDigest || kind == KindSetEqual || kind == KindDerivedDigest
}
func validPolarity(polarity Polarity) bool {
	return polarity == PolarityMatch || polarity == PolarityMismatch
}
func expectedPair(stage, reason string) string { return stage + "\x00" + reason }
