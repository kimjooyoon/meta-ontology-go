package lanefrontier

import (
	"path"
	"sort"
	"strings"
)

func normalizePaths(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		clean, err := normalizePath(value, false)
		if err != nil {
			return nil, err
		}
		normalized = append(normalized, clean)
	}
	sort.Strings(normalized)
	return unique(normalized), nil
}
func normalizePath(value string, owner bool) (string, error) {
	if value == "" || strings.Contains(value, "\\") || strings.ContainsRune(value, '\x00') ||
		strings.HasPrefix(value, "/") || strings.TrimSpace(value) != value {
		return "", errInvalidPath{}
	}
	if owner && value != "." {
		value = strings.TrimSuffix(value, "/")
	}
	if value == "" || path.Clean(value) != value || value == ".." || strings.HasPrefix(value, "../") || (!owner && value == ".") {
		return "", errInvalidPath{}
	}
	return value, nil
}
func pathsInScope(paths, owners []string) bool {
	for _, changed := range paths {
		inScope := false
		for _, owner := range owners {
			if pathContains(owner, changed) {
				inScope = true
				break
			}
		}
		if !inScope {
			return false
		}
	}
	return true
}
func pathContains(owner, value string) bool {
	return owner == "." || owner == value || strings.HasPrefix(value, owner+"/")
}
func unique(values []string) []string {
	if len(values) < 2 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

type errInvalidPath struct{}

func (errInvalidPath) Error() string { return "invalid repository-relative path" }

type errAmbiguousOwner struct{}

func (errAmbiguousOwner) Error() string { return "ambiguous owned path prefixes" }
