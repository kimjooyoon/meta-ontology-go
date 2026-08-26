package claimledger

import (
	"bytes"
	"encoding/json"
	"strings"
)

type InputRecord struct {
	Name   string `json:"name"`
	Digest string `json:"digest,omitempty"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type sourceState struct {
	InputRecord
	Value map[string]any
}

func buildSources(observationData, runtimeData []byte, subject string) (map[string]sourceState, []InputRecord) {
	observation := decodeSource("observation", observationData)
	runtime := validateRuntimeSource(decodeSource("runtime", runtimeData), observation, subject)
	sources := map[string]sourceState{"observation": observation, "runtime": runtime}
	inputs := []InputRecord{observation.InputRecord, runtime.InputRecord}
	return sources, inputs
}

func decodeSource(name string, data []byte) sourceState {
	state := sourceState{InputRecord: InputRecord{Name: name}, Value: map[string]any{}}
	if len(data) == 0 {
		state.Status, state.Reason = "MISSING", strings.ToUpper(name)+"_EVIDENCE_MISSING"
		return state
	}
	state.Digest = digestBytes(data)
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&state.Value); err != nil {
		state.Status, state.Reason = "INVALID", strings.ToUpper(name)+"_EVIDENCE_JSON_INVALID"
		return state
	}
	state.Status, state.Reason = "VERIFIED", "SOURCE_DECODED"
	return state
}

func sourceDigest(sources map[string]sourceState, name string) string {
	if source, ok := sources[name]; ok {
		return source.Digest
	}
	return ""
}
