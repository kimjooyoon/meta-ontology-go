package metarecognition

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

type ReplayDocument struct {
	Schema string       `json:"schema"`
	Cases  []ReplayCase `json:"cases"`
}

type ReplayCase struct {
	ID       string   `json:"id"`
	Subject  Subject  `json:"subject"`
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
		document.Cases = append(document.Cases, ReplayCase{ID: value.ID, Subject: value.Baseline.Subject, Roots: value.Baseline.Roots, Commands: commands, Paths: value.Baseline.Path.IDs})
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

func validateReplay(document ReplayDocument) error {
	if document.Schema != SchemaVersion {
		return fmt.Errorf("replay schema %q is not %q", document.Schema, SchemaVersion)
	}
	caseIDs := make(map[string]struct{}, len(document.Cases))
	for _, value := range document.Cases {
		if value.ID == "" {
			return fmt.Errorf("replay case has empty id")
		}
		if _, exists := caseIDs[value.ID]; exists {
			return fmt.Errorf("duplicate replay case %q", value.ID)
		}
		caseIDs[value.ID] = struct{}{}
		if !value.Subject.Valid() {
			return fmt.Errorf("replay case %q has invalid subject %q", value.ID, value.Subject)
		}
		if err := uniqueValues(value.ID, "root", value.Roots); err != nil {
			return err
		}
		if err := uniqueValues(value.ID, "command", value.Commands); err != nil {
			return err
		}
		if err := uniqueValues(value.ID, "path", value.Paths); err != nil {
			return err
		}
	}
	return nil
}

func uniqueValues(caseID, kind string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			return fmt.Errorf("replay case %q has empty %s", caseID, kind)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("replay case %q has duplicate %s %q", caseID, kind, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	return requireEOF(decoder)
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		return expectDelimiter(decoder, ']')
	case '{':
		keys := make(map[string]struct{})
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if _, exists := keys[name]; exists {
				return fmt.Errorf("duplicate JSON field %q", name)
			}
			keys[name] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		return expectDelimiter(decoder, '}')
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
}

func expectDelimiter(decoder *json.Decoder, expected json.Delim) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token != expected {
		return fmt.Errorf("expected JSON delimiter %q", expected)
	}
	return nil
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("replay JSON has trailing value")
		}
		return err
	}
	return nil
}
