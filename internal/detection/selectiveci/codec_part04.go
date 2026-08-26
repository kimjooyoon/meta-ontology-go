package selectiveci

import (
	"encoding/json"
	"fmt"
)

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delim == '{' {
		seen := map[string]struct{}{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name := key.(string)
			if _, exists := seen[name]; exists {
				return fmt.Errorf("duplicate object field %q", name)
			}
			seen[name] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	}
	if delim == '[' {
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	}
	_, err = decoder.Token()
	return err
}
func requireInputFields(data []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil || root == nil {
		return failure(ReasonInvalidInput, "input must be an object")
	}
	if err := requireFields(root, "schema_version", "base", "head", "registry", "cpu_capacity", "resource_receipts", "provenance_paths"); err != nil {
		return err
	}
	for _, name := range []string{"base", "head"} {
		fields, err := objectFields(root[name])
		if err != nil {
			return err
		}
		if err := requireFields(fields, "schema_version", "snapshot_digest", "files"); err != nil {
			return err
		}
	}
	registry, err := objectFields(root["registry"])
	if err != nil {
		return err
	}
	if err := requireFields(registry, "schema_version", "registry_digest", "policy_digest", "nodes", "dependency_edges", "obligations", "commands", "global_guard_commands"); err != nil {
		return err
	}
	if err := requireArrayObjectFields(root["resource_receipts"], "command_id", "snapshot_digest", "envelope"); err != nil {
		return err
	}
	return requireArrayObjectFields(root["provenance_paths"], "command_id", "path", "requirement")
}
