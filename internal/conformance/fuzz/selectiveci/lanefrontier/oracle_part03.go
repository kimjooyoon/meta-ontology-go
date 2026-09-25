package lanefrontier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func canonicalPath(value string, prefix bool) (string, bool) {
	if prefix {
		value = strings.TrimSuffix(value, "/")
	}
	if value == "" || strings.HasPrefix(value, "/") || strings.ContainsAny(value, "\\\x00") {
		return "", false
	}
	parts := strings.Split(value, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", false
		}
	}
	return strings.Join(parts, "/"), true
}
func ambiguousPrefixes(prefixes []string) bool {
	for i := range prefixes {
		for j := i + 1; j < len(prefixes); j++ {
			if pathContains(prefixes[i], prefixes[j]) || pathContains(prefixes[j], prefixes[i]) {
				return true
			}
		}
	}
	return false
}
func pathsInScope(paths, prefixes []string) bool {
	for _, path := range paths {
		canonical, ok := canonicalPath(path, false)
		if !ok || !ownedPath(canonical, prefixes) {
			return false
		}
	}
	return true
}
func validChangedPaths(paths []string) bool {
	for _, path := range paths {
		if _, ok := canonicalPath(path, false); !ok {
			return false
		}
	}
	return true
}
func ownedPath(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if pathContains(prefix, path) {
			return true
		}
	}
	return false
}
func pathContains(prefix, path string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}
func digestResult(input Input, decision Decision, reason Reason) string {
	canonical := struct {
		Input    Input    `json:"input"`
		Decision Decision `json:"decision"`
		Reason   Reason   `json:"reason"`
	}{normalizedInput(input), decision, reason}
	body, _ := json.Marshal(canonical)
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
