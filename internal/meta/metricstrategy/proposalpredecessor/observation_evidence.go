package proposalpredecessor

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func ValidateObservationEvidence(evidence ObservationEvidence) error {
	if err := validateObservationEvidenceIdentity(evidence); err != nil {
		return err
	}
	if evidence.ResponseConsumed != evidence.ResponseTotal {
		return fmt.Errorf("FAIL_CLOSED: proposal observation evidence coverage is incomplete")
	}
	return nil
}

func validateObservationEvidenceIdentity(evidence ObservationEvidence) error {
	if evidence.Schema != ObservationSchema || evidence.CachePath != ObservationMemberPath || evidence.CacheRole != ObservationRole {
		return fmt.Errorf("FAIL_CLOSED: proposal observation evidence identity is invalid")
	}
	if evidence.CacheBytes <= 0 || evidence.ResponseTotal <= 0 {
		return fmt.Errorf("FAIL_CLOSED: proposal observation evidence coverage is incomplete")
	}
	if len(evidence.CacheDigest) != len("sha256:")+64 || !strings.HasPrefix(evidence.CacheDigest, "sha256:") {
		return fmt.Errorf("FAIL_CLOSED: proposal observation evidence digest is invalid")
	}
	if _, err := hex.DecodeString(strings.TrimPrefix(evidence.CacheDigest, "sha256:")); err != nil {
		return fmt.Errorf("FAIL_CLOSED: proposal observation evidence digest is invalid")
	}
	return nil
}

func ValidateRawObservationCache(evidence ObservationEvidence, raw []byte) error {
	if err := validateObservationEvidenceIdentity(evidence); err != nil {
		return err
	}
	var cache ObservationCache
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cache); err != nil {
		return fmt.Errorf("FAIL_CLOSED: proposal observation cache malformed: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("FAIL_CLOSED: proposal observation cache has trailing content")
	}
	if cache.Schema != ObservationSchema {
		return fmt.Errorf("FAIL_CLOSED: proposal observation cache schema mismatch")
	}
	if len(raw) != evidence.CacheBytes {
		return fmt.Errorf("FAIL_CLOSED: proposal observation cache byte count mismatch")
	}
	sum := sha256.Sum256(raw)
	if evidence.CacheDigest != "sha256:"+hex.EncodeToString(sum[:]) {
		return fmt.Errorf("FAIL_CLOSED: proposal observation cache digest mismatch")
	}
	if len(cache.Responses) != evidence.ResponseTotal {
		return fmt.Errorf("FAIL_CLOSED: proposal observation cache response count mismatch")
	}
	return nil
}

func ValidateRawObservationEvidence(evidence ObservationEvidence, raw []byte, actualConsumed int) error {
	if err := ValidateObservationEvidence(evidence); err != nil {
		return err
	}
	if err := ValidateRawObservationCache(evidence, raw); err != nil {
		return err
	}
	if actualConsumed != evidence.ResponseConsumed || actualConsumed != evidence.ResponseTotal {
		return fmt.Errorf("FAIL_CLOSED: proposal observation consumed count mismatch")
	}
	return nil
}
