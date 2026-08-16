package couplingmanifest

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func sortedSurfaceIDs(values map[semantic.ID]Surface) []semantic.ID {
	result := make([]semantic.ID, 0, len(values))
	for surfaceID := range values {
		result = append(result, surfaceID)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func normalizeID(value semantic.ID) (semantic.ID, error) {
	if value == "" {
		return "", fmt.Errorf("ID is empty")
	}
	parsed, err := semantic.ParseIdentity(value.String())
	if err != nil || parsed != value {
		return "", fmt.Errorf("ID %q is not canonical", value)
	}
	return parsed, nil
}

func normalizeIDString(value string) (semantic.ID, error) { return normalizeID(semantic.ID(value)) }

func normalizeRepoPath(value string) (string, error) {
	if value == "" || value != strings.TrimSpace(value) || strings.Contains(value, "\\") || strings.HasPrefix(value, "/") || (len(value) > 1 && value[1] == ':') {
		return "", fmt.Errorf("path %q is not a repository-relative path", value)
	}
	clean := path.Clean(value)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("path %q escapes the repository", value)
	}
	return clean, nil
}

func normalizeDigest(value, label string) (string, error) {
	canonical, err := rawDigest(value)
	if err != nil {
		return "", failError(CodeMalformedBinding, "%s is malformed", label)
	}
	return canonical, nil
}

func rawDigest(value string) (string, error) {
	if strings.HasPrefix(value, "sha256:") {
		value = strings.TrimPrefix(value, "sha256:")
	}
	if len(value) != 64 || value != strings.ToLower(value) {
		return "", fmt.Errorf("digest is not lowercase SHA-256")
	}
	for _, char := range value {
		if char < '0' || char > '9' && char < 'a' || char > 'f' {
			return "", fmt.Errorf("digest is not lowercase SHA-256")
		}
	}
	return value, nil
}

func sourceMapBindingDigest(surface Surface) string {
	var builder strings.Builder
	field(&builder, surface.SurfaceID.String())
	field(&builder, surface.CodeSymbolID.String())
	field(&builder, surface.SemanticOwnerID.String())
	field(&builder, surface.Binding.SourceMapID.String())
	return stableDigest(builder.String())
}

func isZeroChange(entries []ManifestEntry) bool {
	for _, entry := range entries {
		if entry.BeforeBindingDigest != entry.AfterBindingDigest || entry.BeforeBlobDigest != entry.AfterBlobDigest {
			return false
		}
	}
	return true
}
