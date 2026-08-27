package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func buildProjection(root string, loaded []LoadedManifest) (Projection, error) {
	if diagnostic := validateManifests(loaded, nil); diagnostic != nil {
		return Projection{}, diagnosticError(diagnostic)
	}
	projection := Projection{Schema: projectionSchema}
	for _, item := range sortedLoaded(loaded) {
		manifest := item.Manifest
		projection.Catalog = append(projection.Catalog, CatalogEntry{
			StableID: item.Manifest.StableID, SourceManifest: item.SourcePath,
			Problem: manifest.Concept.Problem, PositiveEffect: manifest.Concept.PositiveEffect,
			MetaOperation: manifest.Concept.MetaOperation, Rarity: manifest.Concept.Rarity,
			Stage: manifest.Concept.Stage, NoveltyClaim: manifest.Concept.NoveltyClaim,
			CodeBindings: sortedStrings(manifest.CodeBindings), MetricBindings: sortedStrings(manifest.MetricBindings),
			UseCases: sortedUseCases(manifest.UseCases), VerificationStrategies: sortedStrings(manifest.VerificationStrategies),
		})
		corpus, err := resourceSnapshots(root, manifest.StableID, manifest.Corpus)
		if err != nil {
			return Projection{}, err
		}
		registry, err := resourceSnapshots(root, manifest.StableID, manifest.Registry)
		if err != nil {
			return Projection{}, err
		}
		documentation, err := resourceSnapshots(root, manifest.StableID, manifest.Documentation)
		if err != nil {
			return Projection{}, err
		}
		projection.Corpus = append(projection.Corpus, corpus...)
		projection.Registry = append(projection.Registry, registry...)
		projection.Documentation = append(projection.Documentation, documentation...)
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

func resourceLess(left, right ResourceSnapshot) bool {
	if left.Path == right.Path {
		if left.StableID == right.StableID {
			return left.Role < right.Role
		}
		return left.StableID < right.StableID
	}
	return left.Path < right.Path
}

func projectionBytes(projection Projection) ([]byte, error) {
	return renderJSON(projection)
}

func sectionDigest(value any) string {
	data, _ := canonicalJSON(value)
	return digestBytes(data)
}

type catalogFile struct {
	Schema        string         `json:"schema"`
	SectionDigest string         `json:"section_digest"`
	Entries       []CatalogEntry `json:"entries"`
}

type resourceFile struct {
	Schema        string             `json:"schema"`
	SectionDigest string             `json:"section_digest"`
	Entries       []ResourceSnapshot `json:"entries"`
}

type denominatorFile struct {
	Schema        string             `json:"schema"`
	SectionDigest string             `json:"section_digest"`
	Entries       []DenominatorEntry `json:"entries"`
}

type manifestDigestFile struct {
	Schema    string           `json:"schema"`
	Manifests []ManifestDigest `json:"manifests"`
}

func renderOutputs(root, outputDir string, loaded []LoadedManifest) (map[string][]byte, DigestFile, error) {
	projection, err := buildProjection(root, loaded)
	if err != nil {
		return nil, DigestFile{}, err
	}
	projectionRaw, err := projectionBytes(projection)
	if err != nil {
		return nil, DigestFile{}, err
	}
	catalogRaw, err := renderJSON(catalogFile{Schema: "gooo/conflict-free-registry-catalog/v1", SectionDigest: sectionDigest(projection.Catalog), Entries: projection.Catalog})
	if err != nil {
		return nil, DigestFile{}, err
	}
	corpusRaw, err := renderJSON(resourceFile{Schema: "gooo/conflict-free-registry-corpus/v1", SectionDigest: sectionDigest(projection.Corpus), Entries: projection.Corpus})
	if err != nil {
		return nil, DigestFile{}, err
	}
	registryRaw, err := renderJSON(resourceFile{Schema: "gooo/conflict-free-registry-registry/v1", SectionDigest: sectionDigest(projection.Registry), Entries: projection.Registry})
	if err != nil {
		return nil, DigestFile{}, err
	}
	denominatorRaw, err := renderJSON(denominatorFile{Schema: "gooo/conflict-free-registry-denominator/v1", SectionDigest: sectionDigest(projection.Denominator), Entries: projection.Denominator})
	if err != nil {
		return nil, DigestFile{}, err
	}
	readmeRaw := renderREADME(projection, loaded)
	manifestDigests := manifestDigestFile{Schema: "gooo/conflict-free-registry-manifests/v1", Manifests: make([]ManifestDigest, 0, len(loaded))}
	semanticViews := make([]semanticManifest, 0, len(loaded))
	for _, item := range sortedLoaded(loaded) {
		manifestDigests.Manifests = append(manifestDigests.Manifests, ManifestDigest{StableID: item.Manifest.StableID, SourceManifest: item.SourcePath, RawDigest: item.RawDigest, SemanticDigest: semanticDigest(item.Manifest)})
		semanticViews = append(semanticViews, semanticView(item.Manifest))
	}
	manifestDigestRaw, err := renderJSON(manifestDigests)
	if err != nil {
		return nil, DigestFile{}, err
	}
	rawDigest := sectionDigest(manifestDigests.Manifests)
	semanticManifestDigest := sectionDigest(semanticViews)
	projectionDigest := digestBytes(projectionRaw)
	combinedDigest := sectionDigest(map[string]string{"raw_manifest_digest": rawDigest, "semantic_manifest_digest": semanticManifestDigest, "projection_digest": projectionDigest})
	outputs := map[string][]byte{
		"projection.json":       projectionRaw,
		"catalog.json":          catalogRaw,
		"corpus.json":           corpusRaw,
		"registry.json":         registryRaw,
		"denominator.json":      denominatorRaw,
		"README.md":             readmeRaw,
		"manifest-digests.json": manifestDigestRaw,
	}
	digest := DigestFile{Schema: digestSchema, RawManifestDigest: rawDigest, SemanticManifestDigest: semanticManifestDigest, ProjectionDigest: projectionDigest, CombinedDigest: combinedDigest}
	digest.Outputs = outputMetadata(root, outputDir, outputs)
	digestRaw, err := renderJSON(digest)
	if err != nil {
		return nil, DigestFile{}, err
	}
	outputs["digest.json"] = digestRaw
	return outputs, digest, nil
}

func outputMetadata(root, outputDir string, outputs map[string][]byte) []OutputMetadata {
	keys := make([]string, 0, len(outputs))
	for key := range outputs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	metadata := make([]OutputMetadata, 0, len(keys))
	for _, key := range keys {
		data := outputs[key]
		logical := filepath.ToSlash(filepath.Join(defaultOutput, key))
		if outputDir != filepath.Join(root, defaultOutput) {
			logical = filepath.ToSlash(filepath.Join(defaultOutput, key))
		}
		metadata = append(metadata, OutputMetadata{Path: logical, Digest: digestBytes(data), Bytes: len(data)})
	}
	return metadata
}

func renderREADME(projection Projection, loaded []LoadedManifest) []byte {
	var builder strings.Builder
	builder.WriteString("# Conflict-free registry projection\n\n")
	builder.WriteString("This bounded projection is generated from local `concept.manifest.json` inputs. Existing concepts own their manifests; catalog, corpus, registry, fixed denominator, and this document are projections.\n\n")
	builder.WriteString("The root topology and root README remain outside this slice. The legacy global catalog remains readable while migration is intentionally bounded to these three concepts.\n\n")
	builder.WriteString("| Stable ID | Stage | Meta operation | Denominator entries | Verification strategies |\n")
	builder.WriteString("| --- | --- | --- | ---: | --- |\n")
	for _, entry := range projection.Catalog {
		denominators := 0
		for _, denominator := range projection.Denominator {
			if denominator.StableID == entry.StableID {
				denominators++
			}
		}
		builder.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %d | `%s` |\n", entry.StableID, entry.Stage, entry.MetaOperation, denominators, strings.Join(entry.VerificationStrategies, ", ")))
	}
	for _, entry := range projection.Catalog {
		builder.WriteString(fmt.Sprintf("\n`%s` — %s → %s\n", entry.StableID, entry.Problem, entry.PositiveEffect))
	}
	builder.WriteString("\n## Local ownership\n\n")
	builder.WriteString(fmt.Sprintf("The projection currently contains %d local manifests. Manifest order is not semantic; generation sorts by stable ID.\n", len(loaded)))
	builder.WriteString("\nGenerated by `scripts/conflict-free-registry-projection` under Go 1.27.0.\n")
	return []byte(builder.String())
}

func writeOutputs(outputDir string, outputs map[string][]byte) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	keys := make([]string, 0, len(outputs))
	for key := range outputs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := os.WriteFile(filepath.Join(outputDir, key), outputs[key], 0o644); err != nil {
			return fmt.Errorf("write generated output %s: %w", key, err)
		}
	}
	return nil
}

func readGenerated(outputDir string) (map[string][]byte, error) {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, err
	}
	outputs := make(map[string][]byte, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			return nil, fmt.Errorf("generated output contains directory %s", entry.Name())
		}
		data, err := os.ReadFile(filepath.Join(outputDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		outputs[entry.Name()] = data
	}
	return outputs, nil
}

func checkRendered(expected, observed map[string][]byte) *Diagnostic {
	for path := range expected {
		if _, ok := observed[path]; !ok {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "REGRESSION", Step: "GENERATED_OUTPUT", Reason: "MISSING_GENERATED_PROJECTION"}
		}
	}
	for path := range observed {
		if _, ok := expected[path]; !ok {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "REGRESSION", Step: "GENERATED_OUTPUT", Reason: "UNEXPECTED_GENERATED_PROJECTION"}
		}
	}
	for path, expectedBytes := range expected {
		if string(expectedBytes) != string(observed[path]) {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "REGRESSION", Step: "GENERATED_OUTPUT", Reason: "STALE_GENERATED_PROJECTION"}
		}
	}
	return nil
}
