package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func canonicalJSON(value any) ([]byte, error) {
	return json.Marshal(value)
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func renderJSON(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func discoverManifestPaths(root string) ([]string, error) {
	base := filepath.Join(root, "examples")
	paths := make([]string, 0, 4)
	err := filepath.WalkDir(base, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && entry.Name() == "concept.manifest.json" {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("discover local manifests: %w", err)
	}
	sort.Strings(paths)
	return paths, nil
}

func loadManifests(root string) ([]LoadedManifest, error) {
	paths, err := discoverManifestPaths(root)
	if err != nil {
		return nil, err
	}
	loaded := make([]LoadedManifest, 0, len(paths))
	for _, path := range paths {
		item, err := loadManifest(root, path)
		if err != nil {
			return nil, err
		}
		loaded = append(loaded, item)
	}
	if diagnostic := validateManifests(loaded, nil); diagnostic != nil {
		return nil, diagnosticError(diagnostic)
	}
	return sortedLoaded(loaded), nil
}

func loadManifest(root, path string) (LoadedManifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return LoadedManifest{}, fmt.Errorf("read local manifest %s: %w", relativePath(root, path), err)
	}
	manifest := Manifest{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return LoadedManifest{}, diagnosticError(&Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "DECODE_MANIFEST", Reason: "MALFORMED_MANIFEST"})
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return LoadedManifest{}, diagnosticError(&Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "DECODE_MANIFEST", Reason: "TRAILING_MANIFEST_CONTENT"})
	}
	return LoadedManifest{Manifest: manifest, SourcePath: relativePath(root, path), RawDigest: digestBytes(raw)}, nil
}

func diagnosticError(diagnostic *Diagnostic) error {
	return diagnosticFailure{Diagnostic: *diagnostic}
}

func validateManifests(loaded []LoadedManifest, requiredIDs []string) *Diagnostic {
	if len(loaded) == 0 {
		return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "REQUIRED_MANIFEST", Reason: "MISSING_MANIFEST"}
	}
	seen := make(map[string]struct{}, len(loaded))
	for _, item := range loaded {
		manifest := item.Manifest
		if manifest.Schema != manifestSchema || manifest.StableID == "" || manifest.Concept.MetaOperation == "" || manifest.Concept.Problem == "" || manifest.Concept.PositiveEffect == "" {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "MANIFEST_IDENTITY", Reason: "INVALID_MANIFEST_IDENTITY"}
		}
		if _, ok := seen[manifest.StableID]; ok {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "UNIQUE_STABLE_ID", Reason: "DUPLICATE_STABLE_ID"}
		}
		seen[manifest.StableID] = struct{}{}
		if manifest.Concept.Stage != "OPERATING" && manifest.Concept.Stage != "CONFORMED" {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "CONCEPT_STAGE", Reason: "UNKNOWN_CONCEPT_STAGE"}
		}
		if len(manifest.VerificationStrategies) == 0 {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "VERIFICATION_STRATEGY", Reason: "MISSING_VERIFICATION_STRATEGY"}
		}
		for _, strategy := range manifest.VerificationStrategies {
			if strategy != "FOUNDATION" && strategy != "COHERENCE" && strategy != "REGRESSION" {
				return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "VERIFICATION_STRATEGY", Reason: "UNKNOWN_VERIFICATION_STRATEGY"}
			}
		}
		if diagnostic := validateRefs(manifest); diagnostic != nil {
			return diagnostic
		}
	}
	if len(requiredIDs) == 0 {
		return nil
	}
	requiredSeen := make(map[string]struct{}, len(loaded))
	for _, item := range loaded {
		requiredSeen[item.Manifest.StableID] = struct{}{}
	}
	for _, requiredID := range requiredIDs {
		if _, ok := requiredSeen[requiredID]; !ok {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "REQUIRED_MANIFEST", Reason: "MISSING_MANIFEST"}
		}
	}
	return nil
}

func validateRefs(manifest Manifest) *Diagnostic {
	for _, ref := range append(append(append([]ResourceRef{}, manifest.Corpus...), manifest.Registry...), manifest.Documentation...) {
		if ref.Path == "" || filepath.IsAbs(ref.Path) || filepath.Clean(ref.Path) == "." || strings.HasPrefix(filepath.ToSlash(filepath.Clean(ref.Path)), "../") {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "RESOURCE_PATH", Reason: "UNSAFE_RESOURCE_PATH"}
		}
		if ref.Role == "" {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "RESOURCE_ROLE", Reason: "MISSING_RESOURCE_ROLE"}
		}
	}
	for _, denominator := range manifest.Denominators {
		if denominator.ID == "" || len(denominator.Values) == 0 {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "DENOMINATOR_IDENTITY", Reason: "INVALID_FIXED_DENOMINATOR"}
		}
	}
	return nil
}

func sortedLoaded(input []LoadedManifest) []LoadedManifest {
	output := append([]LoadedManifest(nil), input...)
	sort.Slice(output, func(i, j int) bool {
		return output[i].Manifest.StableID < output[j].Manifest.StableID
	})
	return output
}

func sortedStrings(input []string) []string {
	output := append([]string(nil), input...)
	sort.Strings(output)
	return output
}

func sortedUseCases(input []UseCase) []UseCase {
	output := append([]UseCase(nil), input...)
	sort.Slice(output, func(i, j int) bool { return output[i].ID < output[j].ID })
	return output
}

func sortedRefs(input []ResourceRef) []ResourceRef {
	output := append([]ResourceRef(nil), input...)
	sort.Slice(output, func(i, j int) bool {
		if output[i].Path == output[j].Path {
			return output[i].Role < output[j].Role
		}
		return output[i].Path < output[j].Path
	})
	return output
}

func sortedDenominators(input []Denominator) []Denominator {
	output := append([]Denominator(nil), input...)
	sort.Slice(output, func(i, j int) bool { return output[i].ID < output[j].ID })
	return output
}

func relativePath(root, path string) string {
	value, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(value)
}

func readSource(root, path string) ([]byte, error) {
	if path == "" || filepath.IsAbs(path) || strings.HasPrefix(filepath.ToSlash(filepath.Clean(path)), "../") {
		return nil, errors.New("unsafe source path")
	}
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
}

func resourceSnapshots(root string, stableID string, refs []ResourceRef) ([]ResourceSnapshot, error) {
	refs = sortedRefs(refs)
	output := make([]ResourceSnapshot, 0, len(refs))
	for _, ref := range refs {
		data, err := readSource(root, ref.Path)
		if err != nil {
			return nil, fmt.Errorf("%s resource %s: %w", stableID, ref.Path, err)
		}
		output = append(output, ResourceSnapshot{StableID: stableID, Path: ref.Path, Role: ref.Role, Bytes: len(data), Digest: digestBytes(data)})
	}
	return output, nil
}

type semanticManifest struct {
	Schema                 string        `json:"schema"`
	StableID               string        `json:"stable_id"`
	Concept                Concept       `json:"concept"`
	CodeBindings           []string      `json:"code_bindings"`
	MetricBindings         []string      `json:"metric_bindings"`
	UseCases               []UseCase     `json:"use_cases"`
	VerificationStrategies []string      `json:"verification_strategies"`
	Corpus                 []ResourceRef `json:"corpus"`
	Registry               []ResourceRef `json:"registry"`
	Denominators           []Denominator `json:"denominators"`
	Documentation          []ResourceRef `json:"documentation"`
}

func semanticView(manifest Manifest) semanticManifest {
	return semanticManifest{
		Schema: manifest.Schema, StableID: manifest.StableID, Concept: manifest.Concept,
		CodeBindings: sortedStrings(manifest.CodeBindings), MetricBindings: sortedStrings(manifest.MetricBindings),
		UseCases: sortedUseCases(manifest.UseCases), VerificationStrategies: sortedStrings(manifest.VerificationStrategies),
		Corpus: sortedRefs(manifest.Corpus), Registry: sortedRefs(manifest.Registry),
		Denominators: sortedDenominators(manifest.Denominators), Documentation: sortedRefs(manifest.Documentation),
	}
}

func semanticDigest(manifest Manifest) string {
	data, _ := canonicalJSON(semanticView(manifest))
	return digestBytes(data)
}
