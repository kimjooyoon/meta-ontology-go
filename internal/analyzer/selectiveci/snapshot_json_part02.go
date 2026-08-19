package selectiveci

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

func decodeSnapshotWire(data []byte) (snapshotWire, error) {
	var wire snapshotWire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return snapshotWire{}, fail(CodeInvalidSchema, "decode snapshot JSON: %v", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return snapshotWire{}, fail(CodeInvalidSchema, "snapshot JSON has trailing values")
		}
		return snapshotWire{}, fail(CodeInvalidSchema, "decode snapshot JSON after object: %v", err)
	}
	return wire, nil
}

type snapshotWire struct {
	Schema            string   `json:"schema"`
	Status            Status   `json:"status"`
	FullSuiteFallback bool     `json:"full_suite_fallback"`
	SourceMapDigest   string   `json:"source_map_digest"`
	RegistryDigest    string   `json:"registry_digest"`
	RegisteredIDs     []string `json:"registered_ids"`
	Sources           []Source `json:"sources"`
	Digest            string   `json:"digest"`
}

func wireForSnapshot(s Snapshot) snapshotWire {
	return snapshotWire{
		Schema: s.Schema, Status: s.Status, FullSuiteFallback: s.FullSuiteFallback,
		SourceMapDigest: s.SourceMapDigest, RegistryDigest: s.RegistryDigest,
		RegisteredIDs: s.RegisteredIDs, Sources: s.Sources, Digest: s.Digest,
	}
}
func (s Snapshot) unsignedJSON() ([]byte, error) {
	type unsignedWire struct {
		Schema            string   `json:"schema"`
		Status            Status   `json:"status"`
		FullSuiteFallback bool     `json:"full_suite_fallback"`
		SourceMapDigest   string   `json:"source_map_digest"`
		RegistryDigest    string   `json:"registry_digest"`
		RegisteredIDs     []string `json:"registered_ids"`
		Sources           []Source `json:"sources"`
	}
	return json.Marshal(unsignedWire{
		Schema: s.Schema, Status: s.Status, FullSuiteFallback: s.FullSuiteFallback,
		SourceMapDigest: s.SourceMapDigest, RegistryDigest: s.RegistryDigest,
		RegisteredIDs: s.RegisteredIDs, Sources: s.Sources,
	})
}
