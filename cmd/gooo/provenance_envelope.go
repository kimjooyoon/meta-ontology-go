package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

type provenanceCLIRecord struct {
	ID         string `json:"id"`
	SemanticID string `json:"semantic_id"`
	Producer   string `json:"producer"`
	Kind       string `json:"kind"`
	Status     string `json:"status"`
	Sequence   uint64 `json:"sequence"`
	Hash       string `json:"hash"`
}

type provenanceCLIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// provenancePublishResponse is a projection of the durable ledger result.
// The ledger remains the only provenance store and authority boundary.
type provenancePublishResponse struct {
	Schema          string                `json:"schema"`
	Status          string                `json:"status"`
	CheckStatus     string                `json:"check_status,omitempty"`
	SourceDigest    string                `json:"source_digest,omitempty"`
	SemanticDigest  string                `json:"semantic_digest,omitempty"`
	GraphDigest     string                `json:"graph_digest,omitempty"`
	StoreDigest     string                `json:"store_digest,omitempty"`
	Records         []provenanceCLIRecord `json:"records"`
	Error           *provenanceCLIError   `json:"error,omitempty"`
	CanonicalDigest string                `json:"canonical_digest"`
}

type provenanceEvidenceEnvelope struct {
	Records []provenance.Evidence `json:"records"`
}

func decodeProvenanceEvidence(data []byte) ([]provenance.Evidence, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, errors.New("evidence input is empty")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	if trimmed[0] == '[' {
		var records []provenance.Evidence
		if err := decoder.Decode(&records); err != nil {
			return nil, err
		}
		if err := requireJSONEOF(decoder); err != nil {
			return nil, err
		}
		return records, nil
	}
	var envelope provenanceEvidenceEnvelope
	if err := decoder.Decode(&envelope); err != nil {
		return nil, err
	}
	if err := requireJSONEOF(decoder); err != nil {
		return nil, err
	}
	return envelope.Records, nil
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("evidence input must contain one JSON value")
		}
		return err
	}
	return nil
}

func writeProvenanceFailure(writer io.Writer, response provenancePublishResponse, code, message string) int {
	response.Status = provenanceStatusRejected
	response.StoreDigest = ""
	response.Records = []provenanceCLIRecord{}
	response.Error = &provenanceCLIError{Code: code, Message: message}
	if err := sealProvenanceResponse(&response); err != nil {
		return exitFailure
	}
	if err := writeProvenanceJSON(writer, response); err != nil {
		return exitFailure
	}
	return exitFailure
}

func sealProvenanceResponse(response *provenancePublishResponse) error {
	response.CanonicalDigest = ""
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	response.CanonicalDigest = sha256Digest(payload)
	return nil
}

func writeProvenanceResponse(writer io.Writer, response provenancePublishResponse, deadline time.Time) error {
	if err := sealProvenanceResponse(&response); err != nil {
		return err
	}
	return writeProvenanceJSONWithDeadline(writer, response, deadline)
}

func writeProvenanceJSON(writer io.Writer, response provenancePublishResponse) error {
	return writeProvenanceJSONWithDeadline(writer, response, time.Time{})
}

func writeProvenanceJSONWithDeadline(writer io.Writer, response provenancePublishResponse, deadline time.Time) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	if len(payload) > maxDiagnosticBytes {
		return errDiagnosticLimit
	}
	return writeInspectOutput(writer, payload, deadline)
}

func sha256Digest(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
