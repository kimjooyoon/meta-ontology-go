package metarecognition

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

type ReplayDocument struct {
	Schema string       `json:"schema"`
	Cases  []ReplayCase `json:"cases"`
}
type ReplayCase struct {
	ID       string   `json:"id"`
	Subject  Subject  `json:"subject"`
	Source   string   `json:"source"`
	Roots    []string `json:"roots"`
	Commands []string `json:"commands"`
	Paths    []string `json:"paths"`
}

func ReplayJSON(cases []Case) ([]byte, error) {
	document := ReplayDocument{Schema: SchemaVersion, Cases: make([]ReplayCase, 0, len(cases))}
	for _, raw := range cases {
		value := raw.normalized()
		commands := make([]string, 0, len(value.Baseline.Commands))
		for _, command := range value.Baseline.Commands {
			commands = append(commands, command.ID)
		}
		source, err := canonicalRootRelativePath(value.Baseline.WorkspaceRoot, value.Baseline.SourcePath)
		if err != nil {
			return nil, fmt.Errorf("case %s: %w", value.ID, err)
		}
		document.Cases = append(document.Cases, ReplayCase{ID: value.ID, Subject: value.Baseline.Subject, Source: source, Roots: value.Baseline.Roots, Commands: commands, Paths: value.Baseline.Path.IDs})
	}
	sort.Slice(document.Cases, func(i, j int) bool { return document.Cases[i].ID < document.Cases[j].ID })
	if err := validateReplay(document); err != nil {
		return nil, err
	}
	return json.Marshal(document)
}
func DecodeReplayJSON(data []byte) (ReplayDocument, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return ReplayDocument{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var document ReplayDocument
	if err := decoder.Decode(&document); err != nil {
		return ReplayDocument{}, err
	}
	if err := requireEOF(decoder); err != nil {
		return ReplayDocument{}, err
	}
	if err := validateReplay(document); err != nil {
		return ReplayDocument{}, err
	}
	return document, nil
}
