package shadow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func decodeFiles(files Files) (decodedInputs, error) {
	var result decodedInputs
	if err := decodeStrict(files.AnalyzerBase, []string{"schema", "status", "full_suite_fallback", "registry_digest", "files", "digest"}, &result.base); err != nil {
		return decodedInputs{}, fmt.Errorf("analyzer base: %w", err)
	}
	if err := decodeStrict(files.AnalyzerHead, []string{"schema", "status", "full_suite_fallback", "registry_digest", "files", "digest"}, &result.head); err != nil {
		return decodedInputs{}, fmt.Errorf("analyzer head: %w", err)
	}
	if err := decodeStrict(files.Planner, []string{"schema", "status", "registry_digest", "base_manifest", "head_manifest", "plan_digest", "changed_root_ids", "selected_command_ids", "selected_guard_command_ids", "selected_work_ids", "commands", "guard_commands"}, &result.planner); err != nil {
		return decodedInputs{}, fmt.Errorf("planner: %w", err)
	}
	if err := decodeStrict(files.Proof, []string{"schema", "status", "fallback", "registry_digest", "plan_digest", "snapshots", "changed_root_ids", "selected_command_ids", "verified_command_ids", "proof_digest"}, &result.proof); err != nil {
		return decodedInputs{}, fmt.Errorf("proof: %w", err)
	}
	if err := decodeStrict(files.Lane, []string{"schema", "decision", "reason", "registry_digest", "base_sha", "lane_head_sha", "lane_id", "registered_branch", "owned_path_prefixes", "changed_paths", "ahead_count", "behind_count", "open_pr_count", "active_lease_count", "canonical_digest"}, &result.lane); err != nil {
		return decodedInputs{}, fmt.Errorf("lane: %w", err)
	}
	return result, nil
}

func decodeStrict(data string, required []string, target any) error {
	trimmed := bytes.TrimSpace([]byte(data))
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return fmt.Errorf("top-level object required")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	if err := scanJSON(decoder); err != nil {
		return fmt.Errorf("strict JSON: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil || fields == nil {
		return fmt.Errorf("top-level object required")
	}
	for _, name := range required {
		if _, ok := fields[name]; !ok {
			return fmt.Errorf("missing required field %q", name)
		}
	}
	valueDecoder := json.NewDecoder(bytes.NewReader(trimmed))
	valueDecoder.DisallowUnknownFields()
	if err := valueDecoder.Decode(target); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return requireEOF(valueDecoder)
}

func scanJSON(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if _, exists := seen[name]; exists {
				return fmt.Errorf("duplicate object field %q", name)
			}
			seen[name] = struct{}{}
			if err := scanJSON(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := scanJSON(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return fmt.Errorf("trailing JSON data: %w", err)
	}
	return nil
}
