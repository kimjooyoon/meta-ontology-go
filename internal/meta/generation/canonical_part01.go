package generation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func normalizeIndicators(indicators []sourcepolicy.Indicator) []sourcepolicy.Indicator {
	result := append([]sourcepolicy.Indicator{}, indicators...)
	sort.Slice(result, func(i, j int) bool {
		return indicatorKey(result[i]) < indicatorKey(result[j])
	})
	return result
}

func indicatorKey(indicator sourcepolicy.Indicator) string {
	payload, _ := json.Marshal(indicator)
	return string(payload)
}

func indicatorID(indicator sourcepolicy.Indicator) string {
	return "sha256:" + digestJSON(indicator)
}

func digestJSON(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func digestPair(left, right string) string {
	digest := sha256.Sum256([]byte(left + "\x00" + right))
	return hex.EncodeToString(digest[:])
}

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, character := range value {
		if !('0' <= character && character <= '9') && !('a' <= character && character <= 'f') {
			return false
		}
	}
	return true
}
