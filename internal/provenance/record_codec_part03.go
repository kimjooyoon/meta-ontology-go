package provenance

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

func decodeEvidence(raw []byte) (Evidence, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var wire wireEvidence
	if err := decoder.Decode(&wire); err != nil {
		return Evidence{}, &lineError{kind: "invalid-json", err: err}
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Evidence{}, &lineError{kind: "multiple-values", err: fmt.Errorf("multiple JSON values")}
		}
		return Evidence{}, &lineError{kind: "invalid-json", err: err}
	}
	if wire.ID != "" && wire.EventID != "" && wire.ID != wire.EventID {
		return Evidence{}, &lineError{kind: "conflicting-id", err: fmt.Errorf("id and event_id differ")}
	}
	id := wire.ID
	if id == "" {
		id = wire.EventID
	}
	return Evidence{
		Schema: wire.Schema, ID: id, SemanticID: wire.SemanticID,
		Producer: wire.Producer, Kind: wire.Kind, Status: wire.Status,
		SourceSpan: SourceSpan{
			URI: wire.SourceSpan.URI, File: wire.SourceSpan.File,
			Start: wire.SourceSpan.Start, End: wire.SourceSpan.End,
		},
		SourceDigest: wire.SourceDigest, SemanticDigest: wire.SemanticDigest,
		GraphDigest: wire.GraphDigest, Sequence: wire.Sequence,
		Predecessor: wire.Predecessor, Attributes: wire.Attributes,
		Freshness: wire.Freshness, Hash: wire.Hash,
		Type: wire.Type, Subject: wire.Subject, GeneratedBy: wire.GeneratedBy,
	}, nil
}
func digestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
