package proposalpredecessor

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func ValidateObservationEvidence(evidence ObservationEvidence) error {
	if evidence.Schema != ObservationSchema || evidence.CachePath == "" {
		return fmt.Errorf("proposal observation evidence identity is invalid")
	}
	if evidence.CacheBytes <= 0 || evidence.ResponseTotal <= 0 || evidence.ResponseConsumed != evidence.ResponseTotal {
		return fmt.Errorf("proposal observation evidence coverage is incomplete")
	}
	if len(evidence.CacheDigest) != len("sha256:")+64 || !strings.HasPrefix(evidence.CacheDigest, "sha256:") {
		return fmt.Errorf("proposal observation evidence digest is invalid")
	}
	if _, err := hex.DecodeString(strings.TrimPrefix(evidence.CacheDigest, "sha256:")); err != nil {
		return fmt.Errorf("proposal observation evidence digest is invalid")
	}
	return nil
}
