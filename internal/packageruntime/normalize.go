package packageruntime

import (
	"sort"
	"strings"
)

func normalizeManifest(manifest Manifest) (Manifest, error) {
	if manifest.Schema != ManifestSchema {
		return Manifest{}, reject("MANIFEST_SCHEMA_UNKNOWN", "schema %q", manifest.Schema)
	}
	if manifest.MutationAuthorized {
		return Manifest{}, reject("MUTATION_AUTHORITY_DENIED", "runtime images are read-only")
	}
	if strings.TrimSpace(manifest.Entry.PackagePath) == "" ||
		strings.TrimSpace(manifest.Entry.Activity) == "" {
		return Manifest{}, reject("ENTRY_UNKNOWN", "entry package and activity are required")
	}
	if len(manifest.Packages) == 0 {
		return Manifest{}, reject("PACKAGE_SET_EMPTY", "at least one package is required")
	}
	result := manifest
	result.Packages = make([]PackageSpec, len(manifest.Packages))
	seen := map[string]bool{}
	for index, spec := range manifest.Packages {
		normalized, err := normalizePackage(spec)
		if err != nil {
			return Manifest{}, err
		}
		if seen[normalized.Path] {
			return Manifest{}, reject("PACKAGE_PATH_DUPLICATE", "path %q", normalized.Path)
		}
		seen[normalized.Path] = true
		result.Packages[index] = normalized
	}
	sort.Slice(result.Packages, func(i, j int) bool {
		return result.Packages[i].Path < result.Packages[j].Path
	})
	return result, nil
}
