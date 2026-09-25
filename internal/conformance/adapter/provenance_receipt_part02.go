package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// CanonicalJSON serializes a validated receipt as one canonical JSONL record.
func (r ProvenanceReceipt) CanonicalJSON() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	normalized := r
	if normalized.Predecessors == nil {
		normalized.Predecessors = []ReceiptPredecessor{}
	}
	return jsonLine(normalized)
}

// Digest returns the SHA-256 digest of the canonical receipt JSONL.
func (r ProvenanceReceipt) Digest() (string, error) {
	payload, err := r.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(payload), nil
}

// ParseProvenanceReceipt accepts only canonical JSONL with known fields.
func ParseProvenanceReceipt(payload []byte) (ProvenanceReceipt, error) {
	var receipt ProvenanceReceipt
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return ProvenanceReceipt{}, fmt.Errorf("decode receipt: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return ProvenanceReceipt{}, fmt.Errorf("receipt has trailing JSON")
		}
		return ProvenanceReceipt{}, fmt.Errorf("receipt has trailing data: %w", err)
	}
	canonical, err := receipt.CanonicalJSON()
	if err != nil {
		return ProvenanceReceipt{}, err
	}
	if !bytes.Equal(payload, canonical) {
		return ProvenanceReceipt{}, fmt.Errorf("receipt is not canonical JSONL")
	}
	return receipt, nil
}

// AppendPredecessor returns a new receipt without mutating the existing record.
func (r ProvenanceReceipt) AppendPredecessor(predecessor ReceiptPredecessor) (ProvenanceReceipt, error) {
	if err := r.Validate(); err != nil {
		return ProvenanceReceipt{}, err
	}
	updated := r
	updated.Predecessors = append([]ReceiptPredecessor{}, r.Predecessors...)
	updated.Predecessors = append(updated.Predecessors, predecessor)
	if err := updated.Validate(); err != nil {
		return ProvenanceReceipt{}, err
	}
	return updated, nil
}
