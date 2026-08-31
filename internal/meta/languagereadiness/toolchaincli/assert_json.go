package toolchaincli

import (
	"encoding/json"
	"strings"
)

type versionPayload struct {
	SchemaVersion string `json:"schema_version"`
	Language      string `json:"language"`
	Version       string `json:"version"`
	Status        string `json:"status"`
	SemanticIR    string `json:"semantic_ir"`
	SemanticCheck string `json:"semantic_check"`
	Graph         string `json:"graph"`
	FixPlan       string `json:"fix_plan"`
}

type commandPayload struct {
	SchemaVersion string            `json:"schema_version"`
	Command       string            `json:"command"`
	Status        string            `json:"status"`
	File          string            `json:"file"`
	Diagnostics   []json.RawMessage `json:"diagnostics"`
	OriginalHash  string            `json:"original_semantic_hash"`
	RoundtripHash string            `json:"round_tripped_semantic_hash"`
	Equivalent    *bool             `json:"equivalent"`
	GetPut        *bool             `json:"get_put"`
	PutGet        *bool             `json:"put_get"`
}

func validVersionJSON(raw string) bool {
	value := versionPayload{}
	if json.Unmarshal([]byte(raw), &value) != nil {
		return false
	}
	return value.SchemaVersion == "gooo-version/v1" && value.Language == "gooo" &&
		value.Version == "0.2.0-dev" && value.Status == "development" &&
		value.SemanticIR != "" && value.SemanticCheck != "" && value.Graph != "" && value.FixPlan != ""
}

func decodeCommandJSON(raw string) (commandPayload, bool) {
	value := commandPayload{}
	err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &value)
	return value, err == nil
}
