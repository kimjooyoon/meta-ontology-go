package resourceenvelope

import (
	"encoding/json"
	"fmt"
	"io"
)

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
