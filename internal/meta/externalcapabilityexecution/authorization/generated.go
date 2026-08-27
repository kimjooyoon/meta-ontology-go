package authorization

import (
	"encoding/json"
	"strings"
)

func normalizeGenerated(path string, raw []byte) ([]byte, error) {
	if !strings.HasSuffix(path, ".manifest.jsonl") {
		return raw, nil
	}
	var manifest map[string]any
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, err
	}
	delete(manifest, "generated_file")
	return json.Marshal(manifest)
}
