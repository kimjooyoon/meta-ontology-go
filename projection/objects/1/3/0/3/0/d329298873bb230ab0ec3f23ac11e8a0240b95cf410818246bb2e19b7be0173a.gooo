package coupling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// New decodes an immutable explanation byte slice and validates all
// identity-bearing locations before any protocol response can be produced.
// The input is copied so later caller mutation cannot change an adapter.
func New(data []byte) (*Adapter, error) {
	raw := append([]byte(nil), data...)
	if err := validateJSONDocument(raw); err != nil {
		return nil, fmt.Errorf("coupling explanation: decode: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var envelope Envelope
	if err := decoder.Decode(&envelope); err != nil {
		return nil, fmt.Errorf("coupling explanation: decode: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("coupling explanation: trailing JSON")
		}
		return nil, fmt.Errorf("coupling explanation: trailing JSON: %w", err)
	}
	if err := validateEnvelope(envelope); err != nil {
		return nil, fmt.Errorf("coupling explanation: %w", err)
	}
	return &Adapter{envelope: envelope, raw: raw}, nil
}

// RawBytes returns a copy for transcript tests and diagnostics. It cannot be
// used to mutate the adapter's snapshot.
func (a *Adapter) RawBytes() []byte { return append([]byte(nil), a.raw...) }
func validateJSONDocument(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON")
		}
		return fmt.Errorf("trailing JSON: %w", err)
	}
	return nil
}
