package freshness

// ValidDigest reports whether value is a lowercase SHA-256 hexadecimal
// digest, the format used by all freshness records.
func ValidDigest(value string) bool {
	if len(value) != digestLength {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}
