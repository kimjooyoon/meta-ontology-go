package verify

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func (r CouplingRegistry) Normalized() (CouplingRegistry, error) {
	if r.Schema != CouplingRegistrySchemaVersion || r.Version != "1" {
		return CouplingRegistry{}, fmt.Errorf("invalid coupling registry schema or version")
	}
	if len(r.Surfaces) == 0 {
		return CouplingRegistry{}, fmt.Errorf("coupling registry has no surfaces")
	}
	out := CouplingRegistry{Schema: r.Schema, Version: r.Version, Surfaces: make([]CouplingSurface, 0, len(r.Surfaces))}
	seenSurface := make(map[string]struct{}, len(r.Surfaces))
	seenSymbol := make(map[string]struct{}, len(r.Surfaces))
	for _, raw := range r.Surfaces {
		surface, err := normalizeSurface(raw)
		if err != nil {
			return CouplingRegistry{}, err
		}
		if _, ok := seenSurface[surface.SurfaceID]; ok {
			return CouplingRegistry{}, fmt.Errorf("duplicate surface ID %q", surface.SurfaceID)
		}
		if _, ok := seenSymbol[surface.CodeSymbolID]; ok {
			return CouplingRegistry{}, fmt.Errorf("duplicate code symbol ID %q", surface.CodeSymbolID)
		}
		seenSurface[surface.SurfaceID] = struct{}{}
		seenSymbol[surface.CodeSymbolID] = struct{}{}
		out.Surfaces = append(out.Surfaces, surface)
	}
	sort.Slice(out.Surfaces, func(i, j int) bool { return out.Surfaces[i].SurfaceID < out.Surfaces[j].SurfaceID })
	return out, nil
}

func normalizeSurface(raw CouplingSurface) (CouplingSurface, error) {
	out := raw
	identityFields := []struct {
		label string
		value string
	}{
		{"surface ID", raw.SurfaceID},
		{"code symbol ID", raw.CodeSymbolID},
		{"semantic owner ID", raw.SemanticOwnerID},
		{"scope ID", raw.ScopeID},
		{"source map ID", raw.SourceMapID},
	}
	for _, field := range identityFields {
		label, value := field.label, field.value
		if _, err := parseStableID(value); err != nil {
			return CouplingSurface{}, fmt.Errorf("%s: %w", label, err)
		}
	}
	for _, value := range []string{raw.SourceMapBindingDigest, raw.ProfileDigest, raw.ToolchainDigest} {
		if !validHexDigest(value) {
			return CouplingSurface{}, fmt.Errorf("surface digest %q is malformed", value)
		}
	}
	if raw.ProfileID == "" || raw.ProfileVersion == "" || len(raw.RuleDigests) == 0 {
		return CouplingSurface{}, fmt.Errorf("surface %q is missing profile or rules", raw.SurfaceID)
	}
	if raw.Applicability != CouplingApplicable && raw.Applicability != CouplingNotApplicable {
		return CouplingSurface{}, fmt.Errorf("surface %q has unknown applicability", raw.SurfaceID)
	}
	if len(raw.CodePathPatterns) == 0 || len(raw.SemanticSourceIDs) == 0 || len(raw.SemanticSourcePaths) == 0 || len(raw.SemanticSourceIDs) != len(raw.SemanticSourcePaths) {
		return CouplingSurface{}, fmt.Errorf("surface %q has incomplete path bindings", raw.SurfaceID)
	}
	out.CodePathPatterns = normalizePatterns(raw.CodePathPatterns)
	if len(out.CodePathPatterns) != len(raw.CodePathPatterns) {
		return CouplingSurface{}, fmt.Errorf("surface %q has duplicate code patterns", raw.SurfaceID)
	}
	out.SemanticSourceIDs = normalizeIDs(raw.SemanticSourceIDs)
	for _, id := range out.SemanticSourceIDs {
		if _, err := parseStableID(id); err != nil {
			return CouplingSurface{}, fmt.Errorf("surface %q source ID: %w", raw.SurfaceID, err)
		}
	}
	out.SemanticSourcePaths = normalizePaths(raw.SemanticSourcePaths)
	for _, value := range out.SemanticSourcePaths {
		if !validRepoPath(value) {
			return CouplingSurface{}, fmt.Errorf("surface %q has invalid semantic source path", raw.SurfaceID)
		}
	}
	out.RuleDigests = normalizeDigests(raw.RuleDigests)
	for _, digest := range out.RuleDigests {
		if !validHexDigest(digest) {
			return CouplingSurface{}, fmt.Errorf("surface %q has malformed rule digest", raw.SurfaceID)
		}
	}
	for _, pattern := range out.CodePathPatterns {
		if !validPattern(pattern) {
			return CouplingSurface{}, fmt.Errorf("surface %q has invalid code pattern %q", raw.SurfaceID, pattern)
		}
	}
	return out, nil
}

func (r CouplingRegistry) Canonical() string {
	normalized, err := r.Normalized()
	if err != nil {
		normalized = r
	}
	var b strings.Builder
	b.WriteString(CouplingRegistrySchemaVersion)
	b.WriteByte('\n')
	for _, s := range normalized.Surfaces {
		writeRegistryField(&b, s.SurfaceID)
		writeRegistryField(&b, s.CodeSymbolID)
		writeRegistryField(&b, s.SemanticOwnerID)
		writeRegistryField(&b, s.ScopeID)
		writeRegistryField(&b, s.SourceMapID)
		writeRegistryField(&b, s.SourceMapBindingDigest)
		writeRegistryField(&b, s.ProfileID)
		writeRegistryField(&b, s.ProfileVersion)
		writeRegistryField(&b, s.ProfileDigest)
		writeRegistryField(&b, s.ToolchainDigest)
		writeRegistryField(&b, s.Applicability)
		for _, value := range sortedStrings(s.CodePathPatterns) {
			writeRegistryField(&b, "code-pattern:"+value)
		}
		for _, value := range sortedStrings(s.SemanticSourceIDs) {
			writeRegistryField(&b, "source-id:"+value)
		}
		for _, value := range sortedStrings(s.SemanticSourcePaths) {
			writeRegistryField(&b, "source-path:"+value)
		}
		for _, value := range sortedStrings(s.RuleDigests) {
			writeRegistryField(&b, "rule:"+value)
		}
	}
	return b.String()
}

func writeRegistryField(b *strings.Builder, value string) {
	b.WriteString(fmt.Sprintf("%d:", len(value)))
	b.WriteString(value)
	b.WriteByte('\n')
}

func (r CouplingRegistry) Digest() string { return semantic.StableHashString(r.Canonical()) }

func (r CouplingRegistry) resolve(sites []ChangedCodeSite) ([]CouplingSurface, error) {
	normalized, err := r.Normalized()
	if err != nil {
		return nil, err
	}
	if len(sites) == 0 {
		return nil, fmt.Errorf("no-changed-sites: changed site set is empty")
	}
	byID := make(map[string]CouplingSurface)
	for _, site := range sites {
		if !validRepoPath(site.Path) {
			return nil, fmt.Errorf("ambiguous-origin: invalid changed path %q", site.Path)
		}
		candidates := make([]CouplingSurface, 0)
		for _, surface := range normalized.Surfaces {
			if site.CodeSymbolID != "" && surface.CodeSymbolID != site.CodeSymbolID {
				continue
			}
			if site.SourceMapBindingDigest != "" && surface.SourceMapBindingDigest != site.SourceMapBindingDigest {
				continue
			}
			for _, pattern := range surface.CodePathPatterns {
				if globMatch(pattern, site.Path) {
					candidates = append(candidates, surface)
					break
				}
			}
		}
		if len(candidates) == 0 {
			return nil, fmt.Errorf("surface-unregistered: %s", site.Path)
		}
		if len(candidates) != 1 {
			return nil, fmt.Errorf("ambiguous-origin: %s", site.Path)
		}
		if candidates[0].Applicability == CouplingNotApplicable {
			return nil, fmt.Errorf("surface-not-applicable: %s", candidates[0].SurfaceID)
		}
		byID[candidates[0].SurfaceID] = candidates[0]
	}
	result := make([]CouplingSurface, 0, len(byID))
	for _, surface := range byID {
		result = append(result, surface)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].SurfaceID < result[j].SurfaceID })
	return result, nil
}

func parseStableID(raw string) (semantic.ID, error) { return semantic.ParseIdentity(raw) }

func validHexDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, c := range value {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

func validCommit(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, c := range value {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

func validRepoPath(value string) bool {
	clean := path.Clean(strings.TrimSpace(value))
	return value != "" && value == clean && value != "." && !strings.HasPrefix(value, "/") && !strings.Contains(value, "\\") && !strings.HasPrefix(value, "../") && value != ".."
}

func validPattern(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") || strings.HasPrefix(value, "../") || value == ".." {
		return false
	}
	return true
}

func normalizePatterns(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if _, ok := seen[value]; !ok {
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}

func normalizePaths(values []string) []string   { return normalizePatterns(values) }
func normalizeIDs(values []string) []string     { return normalizePatterns(values) }
func normalizeDigests(values []string) []string { return normalizePatterns(values) }

func globMatch(pattern, value string) bool {
	if ok, _ := path.Match(pattern, value); ok {
		return true
	}
	if !strings.Contains(pattern, "**") {
		return false
	}
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				b.WriteString(".*")
				i++
			} else {
				b.WriteString("[^/]*")
			}
		case '?':
			b.WriteString("[^/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(pattern[i])))
		}
	}
	b.WriteString("$")
	matched, err := regexp.MatchString(b.String(), value)
	return err == nil && matched
}
