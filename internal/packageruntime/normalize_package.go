package packageruntime

import (
	"sort"
	"strings"
)

func normalizePackage(spec PackageSpec) (PackageSpec, error) {
	if strings.TrimSpace(spec.Path) == "" || strings.TrimSpace(spec.Name) == "" {
		return PackageSpec{}, reject("PACKAGE_IDENTITY_UNKNOWN", "path and name are required")
	}
	if len(spec.Sources) == 0 {
		return PackageSpec{}, reject("PACKAGE_SOURCE_EMPTY", "package %q", spec.Path)
	}
	result := spec
	result.Imports = uniqueSorted(spec.Imports)
	result.Sources = append([]Source(nil), spec.Sources...)
	sort.Slice(result.Sources, func(i, j int) bool {
		return result.Sources[i].Filename < result.Sources[j].Filename
	})
	for index, source := range result.Sources {
		if strings.TrimSpace(source.Filename) == "" || source.Content == "" {
			return PackageSpec{}, reject("PACKAGE_SOURCE_UNKNOWN", "package %q", spec.Path)
		}
		if index > 0 && result.Sources[index-1].Filename == source.Filename {
			return PackageSpec{}, reject("PACKAGE_SOURCE_DUPLICATE", "source %q", source.Filename)
		}
	}
	return result, nil
}

func uniqueSorted(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}
