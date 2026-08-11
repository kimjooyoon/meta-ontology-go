package adapter

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

// CanonicalJSON returns a stable JSONL request suitable for hashing or replay.
func (r Request) CanonicalJSON() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	normalized := r
	var err error
	normalized.Input.IR, err = canonicalRaw(r.Input.IR)
	if err != nil {
		return nil, fmt.Errorf("canonical request IR: %w", err)
	}
	return jsonLine(normalized)
}

// Digest returns the SHA-256 digest of the canonical request JSONL.
func (r Request) Digest() (string, error) {
	payload, err := r.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digest(payload), nil
}

// CanonicalJSON returns a stable JSONL response suitable for comparison.
func (r Response) CanonicalJSON() ([]byte, error) {
	normalized, err := r.Normalized()
	if err != nil {
		return nil, err
	}
	return jsonLine(normalized)
}

// Digest returns the SHA-256 digest of the canonical response JSONL.
func (r Response) Digest() (string, error) {
	payload, err := r.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digest(payload), nil
}

func canonicalRaw(raw json.RawMessage) (json.RawMessage, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return nil, fmt.Errorf("trailing JSON value")
	} else if err != io.EOF {
		return nil, fmt.Errorf("trailing JSON: %w", err)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(encoded), nil
}

func jsonLine(value any) ([]byte, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

func digest(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
