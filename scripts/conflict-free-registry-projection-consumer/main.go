package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

const manifestSchema = "gooo/language-concept-manifest/v1"

type UseCase struct {
	ID              string `json:"id"`
	Trigger         string `json:"trigger"`
	ExpectedOutcome string `json:"expected_outcome"`
}

type Concept struct {
	Problem        string `json:"problem"`
	PositiveEffect string `json:"positive_effect"`
	MetaOperation  string `json:"meta_operation"`
	Rarity         string `json:"rarity"`
	Stage          string `json:"stage"`
	NoveltyClaim   bool   `json:"novelty_claim"`
}

type ResourceRef struct {
	Path string `json:"path"`
	Role string `json:"role"`
}

type Denominator struct {
	ID     string         `json:"id"`
	Values map[string]int `json:"values"`
}

type Manifest struct {
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
	Comments               []string      `json:"comments"`
}

type LoadedManifest struct {
	Manifest   Manifest
	SourcePath string
}

type CatalogEntry struct {
	StableID               string    `json:"stable_id"`
	SourceManifest         string    `json:"source_manifest"`
	Problem                string    `json:"problem"`
	PositiveEffect         string    `json:"positive_effect"`
	MetaOperation          string    `json:"meta_operation"`
	Rarity                 string    `json:"rarity"`
	Stage                  string    `json:"stage"`
	NoveltyClaim           bool      `json:"novelty_claim"`
	CodeBindings           []string  `json:"code_bindings"`
	MetricBindings         []string  `json:"metric_bindings"`
	UseCases               []UseCase `json:"use_cases"`
	VerificationStrategies []string  `json:"verification_strategies"`
}

type ResourceSnapshot struct {
	StableID string `json:"stable_id"`
	Path     string `json:"path"`
	Role     string `json:"role"`
	Bytes    int    `json:"bytes"`
	Digest   string `json:"digest"`
}

type DenominatorEntry struct {
	StableID       string         `json:"stable_id"`
	ID             string         `json:"id"`
	SemanticDigest string         `json:"semantic_digest"`
	Values         map[string]int `json:"values"`
}

type Projection struct {
	Schema        string             `json:"schema"`
	Catalog       []CatalogEntry     `json:"catalog"`
	Corpus        []ResourceSnapshot `json:"corpus"`
	Registry      []ResourceSnapshot `json:"registry"`
	Denominator   []DenominatorEntry `json:"denominator"`
	Documentation []ResourceSnapshot `json:"documentation"`
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

func main() {
	root := "."
	output := ""
	for index := 1; index < len(os.Args); index++ {
		switch os.Args[index] {
		case "-root":
			index++
			if index < len(os.Args) {
				root = os.Args[index]
			}
		case "-output":
			index++
			if index < len(os.Args) {
				output = os.Args[index]
			}
		default:
			fail("FAIL_CLOSED", "FOUNDATION", "COMMAND", "INVALID_COMMAND_FLAGS")
		}
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		fail("FAIL_CLOSED", "FOUNDATION", "ROOT", "INVALID_REPOSITORY_ROOT")
	}
	loaded, err := loadManifests(absoluteRoot)
	if err != nil {
		fail("FAIL_CLOSED", "FOUNDATION", "MANIFESTS", "INDEPENDENT_MANIFEST_RECONSTRUCTION_FAILED")
	}
	projection, err := buildProjection(absoluteRoot, loaded)
	if err != nil {
		fail("FAIL_CLOSED", "FOUNDATION", "PROJECTION", "INDEPENDENT_PROJECTION_FAILED")
	}
	data, err := json.MarshalIndent(projection, "", "  ")
	if err != nil {
		fail("FAIL_CLOSED", "COHERENCE", "SERIALIZE", "INDEPENDENT_PROJECTION_SERIALIZE_FAILED")
	}
	data = append(data, '\n')
	if output == "" {
		_, _ = os.Stdout.Write(data)
		return
	}
	if err := os.WriteFile(output, data, 0o644); err != nil {
		fail("FAIL_CLOSED", "REGRESSION", "OUTPUT", "INDEPENDENT_OUTPUT_WRITE_FAILED")
	}
}

func loadManifests(root string) ([]LoadedManifest, error) {
	paths := []string{}
	err := filepath.WalkDir(filepath.Join(root, "examples"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && entry.Name() == "concept.manifest.json" {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	loaded := make([]LoadedManifest, 0, len(paths))
	seen := map[string]struct{}{}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		manifest := Manifest{}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&manifest); err != nil {
			return nil, err
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			return nil, fmt.Errorf("trailing manifest content")
		}
		if manifest.Schema != manifestSchema || manifest.StableID == "" || manifest.Concept.Problem == "" || manifest.Concept.PositiveEffect == "" || manifest.Concept.MetaOperation == "" {
			return nil, fmt.Errorf("invalid manifest identity")
		}
		if _, ok := seen[manifest.StableID]; ok {
			return nil, fmt.Errorf("duplicate stable id")
		}
		seen[manifest.StableID] = struct{}{}
		loaded = append(loaded, LoadedManifest{Manifest: manifest, SourcePath: relativePath(root, path)})
	}
	if len(loaded) == 0 {
		return nil, fmt.Errorf("missing manifest")
	}
	sort.Slice(loaded, func(i, j int) bool { return loaded[i].Manifest.StableID < loaded[j].Manifest.StableID })
	return loaded, nil
}

func buildProjection(root string, loaded []LoadedManifest) (Projection, error) {
	projection := Projection{Schema: "gooo/conflict-free-registry-projection/v1"}
	for _, item := range loaded {
		manifest := item.Manifest
		projection.Catalog = append(projection.Catalog, CatalogEntry{
			StableID: manifest.StableID, SourceManifest: item.SourcePath, Problem: manifest.Concept.Problem,
			PositiveEffect: manifest.Concept.PositiveEffect, MetaOperation: manifest.Concept.MetaOperation,
			Rarity: manifest.Concept.Rarity, Stage: manifest.Concept.Stage, NoveltyClaim: manifest.Concept.NoveltyClaim,
			CodeBindings: sortedStrings(manifest.CodeBindings), MetricBindings: sortedStrings(manifest.MetricBindings),
			UseCases: sortedUseCases(manifest.UseCases), VerificationStrategies: sortedStrings(manifest.VerificationStrategies),
		})
		for _, ref := range sortedRefs(manifest.Corpus) {
			snapshot, err := snapshot(root, manifest.StableID, ref)
			if err != nil {
				return Projection{}, err
			}
			projection.Corpus = append(projection.Corpus, snapshot)
		}
		for _, ref := range sortedRefs(manifest.Registry) {
			snapshot, err := snapshot(root, manifest.StableID, ref)
			if err != nil {
				return Projection{}, err
			}
			projection.Registry = append(projection.Registry, snapshot)
		}
		for _, ref := range sortedRefs(manifest.Documentation) {
			snapshot, err := snapshot(root, manifest.StableID, ref)
			if err != nil {
				return Projection{}, err
			}
			projection.Documentation = append(projection.Documentation, snapshot)
		}
		for _, denominator := range sortedDenominators(manifest.Denominators) {
			projection.Denominator = append(projection.Denominator, DenominatorEntry{StableID: manifest.StableID, ID: denominator.ID, SemanticDigest: semanticDigest(manifest), Values: denominator.Values})
		}
	}
	sort.Slice(projection.Catalog, func(i, j int) bool { return projection.Catalog[i].StableID < projection.Catalog[j].StableID })
	sort.Slice(projection.Corpus, func(i, j int) bool { return resourceLess(projection.Corpus[i], projection.Corpus[j]) })
	sort.Slice(projection.Registry, func(i, j int) bool { return resourceLess(projection.Registry[i], projection.Registry[j]) })
	sort.Slice(projection.Documentation, func(i, j int) bool { return resourceLess(projection.Documentation[i], projection.Documentation[j]) })
	sort.Slice(projection.Denominator, func(i, j int) bool {
		if projection.Denominator[i].StableID == projection.Denominator[j].StableID {
			return projection.Denominator[i].ID < projection.Denominator[j].ID
		}
		return projection.Denominator[i].StableID < projection.Denominator[j].StableID
	})
	return projection, nil
}

func snapshot(root, stableID string, ref ResourceRef) (ResourceSnapshot, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(ref.Path)))
	if err != nil {
		return ResourceSnapshot{}, err
	}
	return ResourceSnapshot{StableID: stableID, Path: ref.Path, Role: ref.Role, Bytes: len(data), Digest: digest(data)}, nil
}

func semanticDigest(manifest Manifest) string {
	data, _ := json.Marshal(semanticManifest{Schema: manifest.Schema, StableID: manifest.StableID, Concept: manifest.Concept, CodeBindings: sortedStrings(manifest.CodeBindings), MetricBindings: sortedStrings(manifest.MetricBindings), UseCases: sortedUseCases(manifest.UseCases), VerificationStrategies: sortedStrings(manifest.VerificationStrategies), Corpus: sortedRefs(manifest.Corpus), Registry: sortedRefs(manifest.Registry), Denominators: sortedDenominators(manifest.Denominators), Documentation: sortedRefs(manifest.Documentation)})
	return digest(data)
}

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func sortedUseCases(values []UseCase) []UseCase {
	result := append([]UseCase(nil), values...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func sortedRefs(values []ResourceRef) []ResourceRef {
	result := append([]ResourceRef(nil), values...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Path == result[j].Path {
			return result[i].Role < result[j].Role
		}
		return result[i].Path < result[j].Path
	})
	return result
}

func sortedDenominators(values []Denominator) []Denominator {
	result := append([]Denominator(nil), values...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func resourceLess(left, right ResourceSnapshot) bool {
	if left.Path == right.Path {
		if left.StableID == right.StableID {
			return left.Role < right.Role
		}
		return left.StableID < right.StableID
	}
	return left.Path < right.Path
}

func relativePath(root, path string) string {
	value, _ := filepath.Rel(root, path)
	return filepath.ToSlash(value)
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func fail(decision, stage, step, reason string) {
	data, _ := json.MarshalIndent(map[string]string{"decision": decision, "stage": stage, "step": step, "reason": reason}, "", "  ")
	_, _ = os.Stderr.Write(append(data, '\n'))
	os.Exit(1)
}
