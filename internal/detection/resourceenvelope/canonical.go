package resourceenvelope

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type canonicalResult struct {
	CanonicalDigest   string `json:"canonical_digest"`
	CPUCoreNS         uint64 `json:"cpu_core_ns"`
	CPUUtilizationPPM uint64 `json:"cpu_utilization_ppm"`
	PeakRSSBytes      uint64 `json:"peak_rss_bytes"`
	ReadBytes         uint64 `json:"read_bytes"`
	SchemaVersion     string `json:"schema_version"`
	Status            Status `json:"status"`
	WriteBytes        uint64 `json:"write_bytes"`
}

// MarshalJSON keeps direct JSON encoding on the same stable field order as
// CanonicalJSON while retaining the sealed digest value.
func (r Result) MarshalJSON() ([]byte, error) {
	return json.Marshal(canonicalResult{CanonicalDigest: r.CanonicalDigest,
		CPUCoreNS: r.CPUCoreNS, CPUUtilizationPPM: r.CPUUtilizationPPM,
		PeakRSSBytes: r.PeakRSSBytes, ReadBytes: r.ReadBytes,
		SchemaVersion: r.SchemaVersion, Status: r.Status,
		WriteBytes: r.WriteBytes})
}

// CanonicalJSON returns stable JSON for digesting a result. The digest field
// is retained and blanked, which makes the digest input unambiguous.
func (r Result) CanonicalJSON() ([]byte, error) {
	return json.Marshal(canonicalResult{CPUCoreNS: r.CPUCoreNS,
		CPUUtilizationPPM: r.CPUUtilizationPPM, PeakRSSBytes: r.PeakRSSBytes,
		ReadBytes: r.ReadBytes, SchemaVersion: r.SchemaVersion, Status: r.Status,
		WriteBytes: r.WriteBytes})
}

// Canonical returns the stable digest input as a string.
func (r Result) Canonical() string {
	payload, err := r.CanonicalJSON()
	if err != nil {
		return ""
	}
	return string(payload)
}

// CanonicalDigestValue computes the SHA-256 digest of the canonical result.
func (r Result) CanonicalDigestValue() string {
	payload, err := r.CanonicalJSON()
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

// DigestValue is a short alias for callers that do not need the field name.
func (r Result) DigestValue() string { return r.CanonicalDigestValue() }

func (r Result) computedDigest() string { return r.CanonicalDigestValue() }

func sealResult(result Result, reason string) Result {
	result.ReasonCode = reason
	result.Digest = ""
	result.CanonicalDigest = result.computedDigest()
	result.Digest = result.CanonicalDigest
	return result
}

// EncodeResultJSON returns one stable JSON result followed by LF.
func EncodeResultJSON(result Result) ([]byte, error) {
	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("encode resource result: %w", err)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("encode resource result: %w", err)
	}
	return append(payload, '\n'), nil
}
