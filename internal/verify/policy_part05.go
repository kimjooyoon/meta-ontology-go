package verify

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

func lineCount(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	lines := bytes.Count(source, []byte{'\n'})
	if source[len(source)-1] != '\n' {
		lines++
	}
	return lines
}

// CheckPathScope rejects changed paths outside the verifier's ownership
// boundary. Paths are repository-relative and use slash separators.
func CheckPathScope(paths, allowedPrefixes []string) error {
	allowed := normalizePrefixes(allowedPrefixes)
	violations := make([]string, 0)
	for _, path := range sortedUnique(paths) {
		canonical := filepath.ToSlash(filepath.Clean(path))
		if path == "" || path != canonical || strings.Contains(path, "\\") || strings.HasPrefix(path, "/") {
			violations = append(violations, path)
			continue
		}
		if path == "." || isAllowed(path, allowed) {
			continue
		}
		violations = append(violations, path)
	}
	if len(violations) > 0 {
		return fmt.Errorf("changed paths outside CI ownership: %s", strings.Join(violations, ", "))
	}
	return nil
}
func normalizePrefixes(prefixes []string) []string {
	result := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		prefix = strings.Trim(strings.ReplaceAll(prefix, "\\", "/"), "/")
		if prefix != "" {
			result = append(result, prefix)
		}
	}
	return sortedUnique(result)
}
func isAllowed(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}
