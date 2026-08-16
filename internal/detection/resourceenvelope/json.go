package resourceenvelope

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

// EncodeJSON returns the validated envelope as stable indented JSON.
func EncodeJSON(envelope Envelope) ([]byte, error) {
	if err := envelope.Validate(); err != nil {
		return nil, err
	}
	payload, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode resource envelope: %w", err)
	}
	return append(payload, '\n'), nil
}

func (w envelopeWire) envelope() (Envelope, error) {
	if w.SchemaVersion == nil || w.RunnerImageDigest == nil || w.AllocatedCPUCount == nil ||
		w.WarmupCount == nil || w.SampleCount == nil || w.Limits == nil || w.Samples == nil {
		return Envelope{}, fmt.Errorf("missing required field")
	}
	limits, err := w.Limits.limits()
	if err != nil {
		return Envelope{}, err
	}
	samples := make([]Sample, len(*w.Samples))
	for index, raw := range *w.Samples {
		samples[index], err = raw.sample()
		if err != nil {
			return Envelope{}, fmt.Errorf("sample %d: %w", index, err)
		}
	}
	envelope := Envelope{
		SchemaVersion: *w.SchemaVersion, RunnerImageDigest: *w.RunnerImageDigest,
		AllocatedCPUCount: *w.AllocatedCPUCount, WarmupCount: *w.WarmupCount,
		SampleCount: *w.SampleCount, Limits: limits, Samples: samples,
	}
	if err := envelope.Validate(); err != nil {
		return Envelope{}, err
	}
	return envelope, nil
}

func (w *limitsWire) limits() (Limits, error) {
	if w.CPUCoreNS == nil || w.PeakRSSBytes == nil || w.ReadBytes == nil || w.WriteBytes == nil {
		return Limits{}, fmt.Errorf("limits contain a missing required field")
	}
	return Limits{CPUCoreNS: *w.CPUCoreNS, PeakRSSBytes: *w.PeakRSSBytes,
		ReadBytes: *w.ReadBytes, WriteBytes: *w.WriteBytes}, nil
}

func (w sampleWire) sample() (Sample, error) {
	if w.CPUCoreNS == nil || w.WallNS == nil || w.PeakRSSBytes == nil ||
		w.ReadBytes == nil || w.WriteBytes == nil {
		return Sample{}, fmt.Errorf("missing required field")
	}
	return Sample{CPUCoreNS: *w.CPUCoreNS, WallNS: *w.WallNS,
		PeakRSSBytes: *w.PeakRSSBytes, ReadBytes: *w.ReadBytes,
		WriteBytes: *w.WriteBytes}, nil
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := walkJSON(decoder); err != nil {
		return err
	}
	return requireEOF(decoder)
}

func walkJSON(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, isDelim := token.(json.Delim)
	if !isDelim {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if _, exists := seen[name]; exists {
				return fmt.Errorf("duplicate object field %q", name)
			}
			seen[name] = struct{}{}
			if err := walkJSON(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := walkJSON(decoder); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
	_, err = decoder.Token()
	return err
}
