package lanefrontier

import (
	"encoding/json"
	"fmt"
)

func requireInputFields(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("lane frontier input must be an object")
	}
	for _, name := range []string{
		"schema_version", "registry_digest", "base_sha", "lane_head_sha",
		"lane_id", "registered_branch", "owned_path_prefixes", "changed_paths",
		"ahead_count", "behind_count", "open_pr_count", "active_lease_count",
	} {
		if _, ok := fields[name]; !ok {
			return fmt.Errorf("lane frontier input missing %q", name)
		}
	}
	return nil
}
func schemaFromJSON(data []byte) (string, bool) {
	if rejectDuplicateKeys(data) != nil {
		return "", false
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return "", false
	}
	var schema string
	if err := json.Unmarshal(fields["schema_version"], &schema); err != nil {
		return "", false
	}
	return schema, true
}
