package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var digestFields = []string{"semantic_digest", "source_digest", "generated_digest", "source_map_digest", "response_digest"}

// Canonicalize validates generated-manifest evidence and removes only the
// workspace-specific identity from generated_file before it is hashed.
func Canonicalize(root string, data []byte) ([]byte, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) == 0 || (len(lines) == 1 && strings.TrimSpace(lines[0]) == "") {
		return nil, fmt.Errorf("generated manifest is empty")
	}
	var output strings.Builder
	for index, line := range lines {
		canonical, err := canonicalLine(root, line, index+1)
		if err != nil {
			return nil, err
		}
		output.Write(canonical)
		output.WriteByte('\n')
	}
	return []byte(output.String()), nil
}
func canonicalLine(root, line string, lineNumber int) ([]byte, error) {
	if strings.TrimSpace(line) == "" {
		return nil, fmt.Errorf("generated manifest line %d is empty", lineNumber)
	}
	var record map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		return nil, fmt.Errorf("generated manifest line %d is invalid: %w", lineNumber, err)
	}
	relative, err := generatedPath(root, record, lineNumber)
	if err != nil {
		return nil, err
	}
	for _, field := range digestFields {
		var value string
		if raw, ok := record[field]; !ok || json.Unmarshal(raw, &value) != nil || !validDigest(value) {
			return nil, fmt.Errorf("generated manifest line %d has invalid %s", lineNumber, field)
		}
	}
	var generatedDigest string
	if err := json.Unmarshal(record["generated_digest"], &generatedDigest); err != nil {
		return nil, fmt.Errorf("generated manifest line %d has invalid generated_digest", lineNumber)
	}
	generatedData, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
	if err != nil || digest(generatedData) != generatedDigest {
		return nil, fmt.Errorf("generated manifest line %d generated_digest does not match generated_file", lineNumber)
	}
	canonicalPath, err := json.Marshal(filepath.ToSlash(relative))
	if err != nil {
		return nil, err
	}
	record["generated_file"] = canonicalPath
	canonical, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("generated manifest line %d cannot be canonicalized: %w", lineNumber, err)
	}
	return canonical, nil
}
