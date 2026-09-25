package resourceenvelope

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type envelopeWire struct {
	SchemaVersion     *string       `json:"schema_version"`
	RunnerImageDigest *string       `json:"runner_image_digest"`
	AllocatedCPUCount *uint64       `json:"allocated_cpu_count"`
	WarmupCount       *uint64       `json:"warmup_count"`
	SampleCount       *uint64       `json:"sample_count"`
	Limits            *limitsWire   `json:"limits"`
	Samples           *[]sampleWire `json:"samples"`
}
type limitsWire struct {
	CPUCoreNS    *uint64 `json:"cpu_core_ns"`
	PeakRSSBytes *uint64 `json:"peak_rss_bytes"`
	ReadBytes    *uint64 `json:"read_bytes"`
	WriteBytes   *uint64 `json:"write_bytes"`
}
type sampleWire struct {
	CPUCoreNS    *uint64 `json:"cpu_core_ns"`
	WallNS       *uint64 `json:"wall_ns"`
	PeakRSSBytes *uint64 `json:"peak_rss_bytes"`
	ReadBytes    *uint64 `json:"read_bytes"`
	WriteBytes   *uint64 `json:"write_bytes"`
}

// UnmarshalJSON keeps the public Envelope type strict even when callers use
// encoding/json directly instead of DecodeJSON.
func (e *Envelope) UnmarshalJSON(data []byte) error {
	decoded, err := DecodeJSON(data)
	if err != nil {
		return err
	}
	*e = decoded
	return nil
}

// DecodeJSON parses exactly one strict resource envelope.
func DecodeJSON(data []byte) (Envelope, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return Envelope{}, fmt.Errorf("resource envelope is empty")
	}
	if err := rejectDuplicateKeys(trimmed); err != nil {
		return Envelope{}, fmt.Errorf("decode resource envelope: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	var raw *envelopeWire
	if err := decoder.Decode(&raw); err != nil {
		return Envelope{}, fmt.Errorf("decode resource envelope: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return Envelope{}, fmt.Errorf("decode resource envelope: %w", err)
	}
	if raw == nil {
		return Envelope{}, fmt.Errorf("resource envelope must be an object")
	}
	envelope, err := raw.envelope()
	if err != nil {
		return Envelope{}, fmt.Errorf("decode resource envelope: %w", err)
	}
	return envelope, nil
}

// Decode is an alias for DecodeJSON.
func Decode(data []byte) (Envelope, error) { return DecodeJSON(data) }
