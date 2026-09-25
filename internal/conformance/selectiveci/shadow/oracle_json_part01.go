package shadow

import (
	"bytes"
	"encoding/json"
	"fmt"
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
