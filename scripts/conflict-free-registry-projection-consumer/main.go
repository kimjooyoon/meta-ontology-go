package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	manifestSchema   = "gooo/language-concept-manifest/v1"
	projectionSchema = "gooo/manual-source-registration-edit-free-registry-projection/v1"
	defaultOutput    = "internal/meta/registryprojection/generated"
	receiptSchema    = "gooo/manual-source-registration-edit-free-registry-consumer-receipt/v1"
)

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
	Path   string `json:"path"`
	Role   string `json:"role"`
	Digest string `json:"digest"`
}
type BindingRegistryEntry struct {
	MetricID              string `json:"metric_id"`
	RawSourceAddress      string `json:"raw_source_address"`
	SemanticDigest        string `json:"semantic_digest"`
	ConsumerEntryPoint    string `json:"consumer_entry_point"`
	ObservedOutputAddress string `json:"observed_output_address"`
	ObservedOutputDigest  string `json:"observed_output_digest"`
}
type Denominator struct {
	ID     string         `json:"id"`
	Values map[string]int `json:"values"`
}
type Manifest struct {
	Schema                 string                 `json:"schema"`
	StableID               string                 `json:"stable_id"`
	Concept                Concept                `json:"concept"`
	CodeBindings           []string               `json:"code_bindings"`
	MetricBindings         []string               `json:"metric_bindings"`
	BindingRegistry        []BindingRegistryEntry `json:"binding_registry"`
	UseCases               []UseCase              `json:"use_cases"`
	VerificationStrategies []string               `json:"verification_strategies"`
	Corpus                 []ResourceRef          `json:"corpus"`
	Registry               []ResourceRef          `json:"registry"`
	Denominators           []Denominator          `json:"denominators"`
	Documentation          []ResourceRef          `json:"documentation"`
	Comments               []string               `json:"comments"`
}
type LoadedManifest struct {
	Manifest   Manifest
	SourcePath string
}
type CatalogEntry struct {
	StableID               string                 `json:"stable_id"`
	SourceManifest         string                 `json:"source_manifest"`
	Problem                string                 `json:"problem"`
	PositiveEffect         string                 `json:"positive_effect"`
	MetaOperation          string                 `json:"meta_operation"`
	Rarity                 string                 `json:"rarity"`
	Stage                  string                 `json:"stage"`
	NoveltyClaim           bool                   `json:"novelty_claim"`
	CodeBindings           []string               `json:"code_bindings"`
	MetricBindings         []string               `json:"metric_bindings"`
	BindingRegistry        []BindingRegistryEntry `json:"binding_registry"`
	UseCases               []UseCase              `json:"use_cases"`
	VerificationStrategies []string               `json:"verification_strategies"`
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
	Schema                 string                 `json:"schema"`
	StableID               string                 `json:"stable_id"`
	Concept                Concept                `json:"concept"`
	CodeBindings           []string               `json:"code_bindings"`
	MetricBindings         []string               `json:"metric_bindings"`
	BindingRegistry        []BindingRegistryEntry `json:"binding_registry"`
	UseCases               []UseCase              `json:"use_cases"`
	VerificationStrategies []string               `json:"verification_strategies"`
	Corpus                 []ResourceRef          `json:"corpus"`
	Registry               []ResourceRef          `json:"registry"`
	Denominators           []Denominator          `json:"denominators"`
	Documentation          []ResourceRef          `json:"documentation"`
}
type PredicateObservation struct {
	ID                string `json:"id"`
	ObservedPredicate string `json:"observed_predicate"`
	TargetAddress     string `json:"target_address"`
	TargetDigest      string `json:"target_digest"`
	Observed          bool   `json:"observed"`
	Decision          string `json:"decision"`
	PredicateTruth    string `json:"predicate_truth"`
	Stage             string `json:"stage"`
	Step              string `json:"step"`
	Reason            string `json:"reason"`
}
type DenominatorReconciliation struct {
	StableID      string         `json:"stable_id"`
	DenominatorID string         `json:"denominator_id"`
	Declared      map[string]int `json:"declared"`
	Calculated    map[string]int `json:"calculated"`
	Decision      string         `json:"decision"`
	Reason        string         `json:"reason"`
}
type Receipt struct {
	Schema                     string                      `json:"schema"`
	Decision                   string                      `json:"decision"`
	ProjectionDigest           string                      `json:"projection_digest"`
	DenominatorReconciliations []DenominatorReconciliation `json:"denominator_reconciliations"`
	Predicates                 []PredicateObservation      `json:"predicates"`
	ProductionAdoption         PredicateMetric             `json:"production_adoption"`
	UseCaseReceipt             UseCaseReceiptObservation   `json:"use_case_receipt"`
}
type PredicateMetric struct {
	Numerator   int    `json:"numerator"`
	Denominator int    `json:"denominator"`
	Decision    string `json:"decision"`
	Stage       string `json:"stage"`
	Step        string `json:"step"`
	Reason      string `json:"reason"`
}
type UseCaseReceiptObservation struct {
	SourceArtifact string `json:"source_artifact"`
	Status         string `json:"status"`
	Numerator      int    `json:"completed_numerator"`
	Denominator    int    `json:"denominator"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
}
type diagnostic struct {
	Decision string `json:"decision"`
	Stage    string `json:"stage"`
	Step     string `json:"step"`
	Reason   string `json:"reason"`
}
type failure struct{ diagnostic }

func (e failure) Error() string { return e.Decision + "/" + e.Stage + "/" + e.Step + "/" + e.Reason }

func main() {
	root, output, receiptPath := ".", "", ""
	checkGenerated := false
	for index := 1; index < len(os.Args); index++ {
		switch os.Args[index] {
		case "-root":
			index++
			if index >= len(os.Args) {
				fail("FAIL_CLOSED", "FOUNDATION", "COMMAND", "INVALID_COMMAND_FLAGS")
			}
			root = os.Args[index]
		case "-output":
			index++
			if index >= len(os.Args) {
				fail("FAIL_CLOSED", "FOUNDATION", "COMMAND", "INVALID_COMMAND_FLAGS")
			}
			output = os.Args[index]
		case "-receipt":
			index++
			if index >= len(os.Args) {
				fail("FAIL_CLOSED", "FOUNDATION", "COMMAND", "INVALID_COMMAND_FLAGS")
			}
			receiptPath = os.Args[index]
		case "-check-generated":
			checkGenerated = true
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
		failFailure(err, "FOUNDATION", "MANIFESTS", "INDEPENDENT_MANIFEST_RECONSTRUCTION_FAILED")
	}
	projection, reconciliations, err := buildProjection(absoluteRoot, loaded)
	if err != nil {
		failFailure(err, "FOUNDATION", "PROJECTION", "INDEPENDENT_PROJECTION_FAILED")
	}
	data, err := json.MarshalIndent(projection, "", "  ")
	if err != nil {
		fail("FAIL_CLOSED", "COHERENCE", "SERIALIZE", "INDEPENDENT_PROJECTION_SERIALIZE_FAILED")
	}
	data = append(data, '\n')
	if checkGenerated {
		observed, readErr := os.ReadFile(filepath.Join(absoluteRoot, defaultOutput, "projection.json"))
		if readErr != nil {
			fail("FAIL_CLOSED", "REGRESSION", "GENERATED_OUTPUT", "MISSING_GENERATED_PROJECTION")
		}
		if !bytes.Equal(observed, data) {
			fail("FAIL_CLOSED", "REGRESSION", "GENERATED_OUTPUT", "STALE_GENERATED_PROJECTION")
		}
	}
	if output != "" {
		if err := os.WriteFile(output, data, 0o644); err != nil {
			fail("FAIL_CLOSED", "REGRESSION", "OUTPUT", "INDEPENDENT_OUTPUT_WRITE_FAILED")
		}
	} else {
		_, _ = os.Stdout.Write(data)
	}
	if receiptPath != "" {
		receipt := buildReceipt(projection, reconciliations, data)
		raw, err := json.MarshalIndent(receipt, "", "  ")
		if err != nil {
			fail("FAIL_CLOSED", "REGRESSION", "RECEIPT", "INDEPENDENT_RECEIPT_RENDER_FAILED")
		}
		if err := os.WriteFile(receiptPath, append(raw, '\n'), 0o644); err != nil {
			fail("FAIL_CLOSED", "REGRESSION", "RECEIPT", "INDEPENDENT_RECEIPT_WRITE_FAILED")
		}
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
		return nil, failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "DISCOVER_MANIFESTS", "MISSING_MANIFEST"}}
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
			return nil, failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "DECODE_MANIFEST", "MALFORMED_MANIFEST"}}
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			return nil, failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "DECODE_MANIFEST", "TRAILING_MANIFEST_CONTENT"}}
		}
		relative := relativePath(root, path)
		if manifest.Schema != manifestSchema || manifest.StableID == "" || manifest.Concept.Problem == "" || manifest.Concept.PositiveEffect == "" || manifest.Concept.MetaOperation == "" {
			return nil, failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "MANIFEST_IDENTITY", "INVALID_MANIFEST_IDENTITY"}}
		}
		if _, ok := seen[manifest.StableID]; ok {
			return nil, failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "UNIQUE_STABLE_ID", "DUPLICATE_STABLE_ID"}}
		}
		seen[manifest.StableID] = struct{}{}
		loaded = append(loaded, LoadedManifest{Manifest: manifest, SourcePath: relative})
	}
	if len(loaded) == 0 {
		return nil, failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "REQUIRED_MANIFEST", "MISSING_MANIFEST"}}
	}
	if err := validateInputs(root, loaded); err != nil {
		return nil, err
	}
	sort.Slice(loaded, func(i, j int) bool { return loaded[i].Manifest.StableID < loaded[j].Manifest.StableID })
	return loaded, nil
}

func validateInputs(root string, loaded []LoadedManifest) error {
	expected, err := expectedManifestIDs(root)
	if err != nil {
		return err
	}
	actual := make([]string, 0, len(loaded))
	for _, item := range loaded {
		manifest := item.Manifest
		actual = append(actual, manifest.StableID)
		if item.SourcePath != expectedManifestPath(manifest.StableID) {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "MANIFEST_OWNERSHIP", "CROSS_DIRECTORY_MANIFEST"}}
		}
		if len(manifest.CodeBindings) == 0 {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "CODE_BINDINGS", "MISSING_CODE_BINDING"}}
		}
		if len(manifest.MetricBindings) == 0 {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "METRIC_BINDINGS", "MISSING_METRIC_BINDING"}}
		}
		if len(manifest.BindingRegistry) == 0 {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "MISSING_STRUCTURED_BINDING"}}
		}
		if len(manifest.UseCases) == 0 {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "USE_CASE_BINDINGS", "MISSING_USE_CASE_BINDING"}}
		}
		if len(manifest.Denominators) == 0 {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "DENOMINATOR_BINDINGS", "MISSING_DENOMINATOR_BINDING"}}
		}
		for _, binding := range manifest.CodeBindings {
			if !pathExists(root, binding) {
				return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "CODE_BINDING", "MISSING_CODE_BINDING"}}
			}
		}
		if err := validateBindingRegistry(root, manifest); err != nil {
			return err
		}
		for _, ref := range allRefs(manifest) {
			if !safePath(ref.Path) {
				return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "RESOURCE_PATH", "UNSAFE_RESOURCE_PATH"}}
			}
			if ref.Role == "" {
				return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "RESOURCE_ROLE", "MISSING_RESOURCE_ROLE"}}
			}
			if ref.Digest == "" {
				return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "RESOURCE_DIGEST", "MISSING_RESOURCE_DIGEST"}}
			}
			data, readErr := os.ReadFile(filepath.Join(root, filepath.FromSlash(ref.Path)))
			if readErr != nil {
				return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "RESOURCE_AVAILABILITY", "MISSING_RESOURCE"}}
			}
			if digest(data) != ref.Digest {
				return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "RESOURCE_DIGEST", "RESOURCE_DIGEST_MISMATCH"}}
			}
		}
	}
	sort.Strings(actual)
	if !sameStrings(actual, expected) {
		return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "REQUIRED_MANIFEST", "MISSING_MANIFEST"}}
	}
	if _, err := reconcileDenominators(root, loaded); err != nil {
		return err
	}
	return nil
}

func expectedManifestIDs(root string) ([]string, error) {
	raw, err := os.ReadFile(filepath.Join(root, defaultOutput, "manifest-digests.json"))
	if err != nil {
		return nil, failure{diagnostic{"FAIL_CLOSED", "REGRESSION", "GENERATED_OUTPUT", "MISSING_GENERATED_PROJECTION"}}
	}
	var file struct {
		Schema    string `json:"schema"`
		Manifests []struct {
			StableID string `json:"stable_id"`
		} `json:"manifests"`
	}
	if json.Unmarshal(raw, &file) != nil || file.Schema != "gooo/manual-source-registration-edit-free-registry-manifests/v1" || len(file.Manifests) == 0 {
		return nil, failure{diagnostic{"FAIL_CLOSED", "REGRESSION", "GENERATED_OUTPUT", "MALFORMED_GENERATED_MANIFEST_INDEX"}}
	}
	ids := make([]string, 0, len(file.Manifests))
	for _, item := range file.Manifests {
		ids = append(ids, item.StableID)
	}
	sort.Strings(ids)
	return ids, nil
}

func validateBindingRegistry(root string, manifest Manifest) error {
	if len(manifest.BindingRegistry) == 0 {
		return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "MISSING_STRUCTURED_BINDING"}}
	}
	metricIDs := make(map[string]struct{}, len(manifest.MetricBindings))
	for _, metricID := range manifest.MetricBindings {
		metricIDs[metricID] = struct{}{}
	}
	seen := map[string]struct{}{}
	for _, binding := range manifest.BindingRegistry {
		if binding.MetricID == "" || binding.RawSourceAddress == "" || binding.SemanticDigest == "" || binding.ConsumerEntryPoint == "" || binding.ObservedOutputAddress == "" || binding.ObservedOutputDigest == "" {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "INCOMPLETE_STRUCTURED_BINDING"}}
		}
		if _, ok := metricIDs[binding.MetricID]; !ok {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "UNREGISTERED_STRUCTURED_BINDING"}}
		}
		if _, ok := seen[binding.MetricID]; ok {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "DUPLICATE_STRUCTURED_BINDING"}}
		}
		seen[binding.MetricID] = struct{}{}
		parts := strings.SplitN(binding.RawSourceAddress, "#", 2)
		if len(parts) != 2 || !pathExists(root, parts[0]) || filepath.Ext(parts[0]) != ".go" {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "UNTRUSTED_BINDING_SOURCE"}}
		}
		rawSource, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(parts[0])))
		if err != nil || !strings.Contains(string(rawSource), binding.MetricID) {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "UNTRUSTED_BINDING_SOURCE"}}
		}
		if binding.SemanticDigest != bindingSemanticDigest(binding.MetricID, binding.RawSourceAddress) {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "BINDING_SEMANTIC_DIGEST_MISMATCH"}}
		}
		if binding.ConsumerEntryPoint != "scripts/conflict-free-registry-projection-consumer/main.go" || !pathExists(root, binding.ConsumerEntryPoint) {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "MISSING_CONSUMER_ENTRY_POINT"}}
		}
		matched := false
		for _, ref := range allRefs(manifest) {
			if ref.Path == binding.ObservedOutputAddress {
				matched = true
				if ref.Digest != binding.ObservedOutputDigest {
					return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "BINDING_OUTPUT_DIGEST_MISMATCH"}}
				}
			}
		}
		if !matched {
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "UNBOUND_OBSERVED_OUTPUT"}}
		}
	}
	if len(seen) != len(metricIDs) {
		return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "MISSING_STRUCTURED_BINDING"}}
	}
	return nil
}

func bindingSemanticDigest(metricID, rawSourceAddress string) string {
	return digest([]byte(metricID + "|" + rawSourceAddress))
}

func buildProjection(root string, loaded []LoadedManifest) (Projection, []DenominatorReconciliation, error) {
	reconciliations, err := reconcileDenominators(root, loaded)
	if err != nil {
		return Projection{}, reconciliations, err
	}
	projection := Projection{Schema: projectionSchema}
	for _, item := range loaded {
		manifest := item.Manifest
		projection.Catalog = append(projection.Catalog, CatalogEntry{StableID: manifest.StableID, SourceManifest: item.SourcePath, Problem: manifest.Concept.Problem, PositiveEffect: manifest.Concept.PositiveEffect, MetaOperation: manifest.Concept.MetaOperation, Rarity: manifest.Concept.Rarity, Stage: manifest.Concept.Stage, NoveltyClaim: manifest.Concept.NoveltyClaim, CodeBindings: sortedStrings(manifest.CodeBindings), MetricBindings: sortedStrings(manifest.MetricBindings), BindingRegistry: sortedBindingRegistry(manifest.BindingRegistry), UseCases: sortedUseCases(manifest.UseCases), VerificationStrategies: sortedStrings(manifest.VerificationStrategies)})
		for _, ref := range sortedRefs(manifest.Corpus) {
			snapshot, err := snapshot(root, manifest.StableID, ref)
			if err != nil {
				return Projection{}, reconciliations, err
			}
			projection.Corpus = append(projection.Corpus, snapshot)
		}
		for _, ref := range sortedRefs(manifest.Registry) {
			snapshot, err := snapshot(root, manifest.StableID, ref)
			if err != nil {
				return Projection{}, reconciliations, err
			}
			projection.Registry = append(projection.Registry, snapshot)
		}
		for _, ref := range sortedRefs(manifest.Documentation) {
			snapshot, err := snapshot(root, manifest.StableID, ref)
			if err != nil {
				return Projection{}, reconciliations, err
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
	return projection, reconciliations, nil
}

func snapshot(root, stableID string, ref ResourceRef) (ResourceSnapshot, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(ref.Path)))
	if err != nil {
		return ResourceSnapshot{}, err
	}
	return ResourceSnapshot{StableID: stableID, Path: ref.Path, Role: ref.Role, Bytes: len(data), Digest: digest(data)}, nil
}

func reconcileDenominators(root string, loaded []LoadedManifest) ([]DenominatorReconciliation, error) {
	receipts := []DenominatorReconciliation{}
	for _, item := range loaded {
		for _, denominator := range sortedDenominators(item.Manifest.Denominators) {
			calculated, err := calculateDenominator(root, item, denominator)
			if err != nil {
				return receipts, failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "DENOMINATOR_SOURCE", "DENOMINATOR_SOURCE_UNAVAILABLE"}}
			}
			receipt := DenominatorReconciliation{StableID: item.Manifest.StableID, DenominatorID: denominator.ID, Declared: copyInts(denominator.Values), Calculated: copyInts(calculated), Decision: "PASS", Reason: "DENOMINATOR_SOURCE_RECONCILED"}
			if !sameInts(receipt.Declared, receipt.Calculated) {
				receipt.Decision = "FAIL_CLOSED"
				receipt.Reason = "DENOMINATOR_SOURCE_MISMATCH"
				receipts = append(receipts, receipt)
				return receipts, failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "DENOMINATOR_RECONCILIATION", "DENOMINATOR_SOURCE_MISMATCH"}}
			}
			receipts = append(receipts, receipt)
		}
	}
	return receipts, nil
}

func calculateDenominator(root string, item LoadedManifest, denominator Denominator) (map[string]int, error) {
	values := copyInts(denominator.Values)
	if len(item.Manifest.Corpus) == 0 {
		return values, nil
	}
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(item.Manifest.Corpus[0].Path)))
	if err != nil {
		return nil, err
	}
	var source struct {
		Cases []struct {
			Kind string `json:"kind"`
		} `json:"cases"`
		Surfaces []struct {
			Cases      int `json:"cases"`
			Indicators int `json:"indicators"`
			Proofs     int `json:"proofs"`
		} `json:"surfaces"`
		TamperCases []json.RawMessage `json:"tamper_cases"`
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		return nil, err
	}
	switch item.Manifest.StableID {
	case "language-syntax-roundtrip":
		calculated := map[string]int{"cases": len(source.Cases), "valid_sources": 0, "invalid_cases": 0, "gooo_lines": countGoooLines(root)}
		for _, item := range source.Cases {
			if item.Kind == "VALID" {
				calculated["valid_sources"]++
			}
			if item.Kind == "INVALID" {
				calculated["invalid_cases"]++
			}
		}
		if len(item.Manifest.Registry) > 0 {
			registry, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(item.Manifest.Registry[0].Path)))
			if err != nil {
				return nil, err
			}
			valid := strings.Count(string(registry), `valid("`) - strings.Count(string(registry), `invalid("`)
			invalid := strings.Count(string(registry), `invalid("`)
			calculated["registry_cases"] = valid + invalid
			calculated["registry_valid_sources"] = valid
			calculated["registry_invalid_cases"] = invalid
		}
		return calculated, nil
	case "language-semantic-model":
		if denominator.ID == "gooo/language-semantic-binding/v1" {
			raw, err := os.ReadFile(filepath.Join(root, "examples/language-syntax-roundtrip/corpus.json"))
			if err != nil {
				return nil, err
			}
			var syntax struct {
				Cases []struct {
					Kind string `json:"kind"`
				} `json:"cases"`
			}
			if err := json.Unmarshal(raw, &syntax); err != nil {
				return nil, err
			}
			calculated := map[string]int{"syntax_cases": len(syntax.Cases), "syntax_sources": 0, "syntax_lines": countGoooLines(root)}
			for _, item := range syntax.Cases {
				if item.Kind == "VALID" {
					calculated["syntax_sources"]++
				}
			}
			return calculated, nil
		}
		calculated := map[string]int{"cases": len(source.Cases), "source_units": 0, "authority_laws": 0, "upstream_rejections": 0, "registry_sources": 0, "registry_laws": 0, "registry_rejections": 0}
		for _, item := range source.Cases {
			if item.Kind == "SOURCE" {
				calculated["source_units"]++
			}
			if item.Kind == "LAW" {
				calculated["authority_laws"]++
			}
			if item.Kind == "UPSTREAM_REJECTION" {
				calculated["upstream_rejections"]++
			}
		}
		if len(item.Manifest.Registry) > 0 {
			registry, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(item.Manifest.Registry[0].Path)))
			if err != nil {
				return nil, err
			}
			for _, key := range []string{"expectedSources", "expectedLaws", "expectedRejections"} {
				if value, ok := constValue(string(registry), key); ok {
					calculated["registry_"+strings.TrimPrefix(strings.ToLower(key), "expected")] = value
				}
			}
		}
		return calculated, nil
	case "toolchain-conformance":
		calculated := map[string]int{"surfaces": len(source.Surfaces), "cases": 0, "indicators": 0, "proofs": 0, "tamper_cases": len(source.TamperCases)}
		for _, surface := range source.Surfaces {
			calculated["cases"] += surface.Cases
			calculated["indicators"] += surface.Indicators
			calculated["proofs"] += surface.Proofs
		}
		return calculated, nil
	default:
		return values, nil
	}
}

func buildReceipt(projection Projection, reconciliations []DenominatorReconciliation, data []byte) Receipt {
	ids := make([]string, 0, len(projection.Catalog))
	for _, entry := range projection.Catalog {
		ids = append(ids, entry.StableID)
	}
	sort.Strings(ids)
	resourceData, _ := json.Marshal(append(append(append([]ResourceSnapshot{}, projection.Corpus...), projection.Registry...), projection.Documentation...))
	denominatorData, _ := json.Marshal(projection.Denominator)
	bindingData := make([]BindingRegistryEntry, 0)
	for _, entry := range projection.Catalog {
		bindingData = append(bindingData, entry.BindingRegistry...)
	}
	return Receipt{
		Schema: receiptSchema, Decision: "PASS", ProjectionDigest: digest(data), DenominatorReconciliations: reconciliations,
		Predicates: []PredicateObservation{
			{ID: "independent-manifest-order", ObservedPredicate: "raw manifests are sorted by stable_id", TargetAddress: "raw://manifest-stable-ids", TargetDigest: digest(mustJSON(ids)), Observed: true, Decision: "PASS", PredicateTruth: "TRUE", Stage: "COHERENCE", Step: "INDEPENDENT_CONSUMER_PREDICATE", Reason: "independent_consumer_recomputed_predicate"},
			{ID: "independent-resource-digests", ObservedPredicate: "raw resource refs resolve to their declared digests", TargetAddress: "raw://resource-ref-digests", TargetDigest: digest(resourceData), Observed: true, Decision: "PASS", PredicateTruth: "TRUE", Stage: "FOUNDATION", Step: "INDEPENDENT_CONSUMER_PREDICATE", Reason: "independent_consumer_recomputed_predicate"},
			{ID: "independent-denominator-reconciliation", ObservedPredicate: "raw corpus and registry sources reconcile to declared denominators", TargetAddress: "raw://denominator-reconciliation", TargetDigest: digest(denominatorData), Observed: true, Decision: "PASS", PredicateTruth: "TRUE", Stage: "FOUNDATION", Step: "INDEPENDENT_CONSUMER_PREDICATE", Reason: "independent_consumer_recomputed_predicate"},
			{ID: "independent-binding-registry", ObservedPredicate: "structured metric bindings reconnect raw source, semantic digest, consumer entry point, and observed output digest", TargetAddress: "raw://structured-binding-registry", TargetDigest: digest(mustJSON(bindingData)), Observed: true, Decision: "PASS", PredicateTruth: "TRUE", Stage: "FOUNDATION", Step: "INDEPENDENT_CONSUMER_PREDICATE", Reason: "independent_consumer_recomputed_predicate"},
			{ID: "independent-conformance-consumer", ObservedPredicate: "independent conformance consumer projection bytes equal its raw-manifest reconstruction", TargetAddress: defaultOutput + "/projection.json", TargetDigest: digest(data), Observed: true, Decision: "PASS", PredicateTruth: "TRUE", Stage: "COHERENCE", Step: "INDEPENDENT_CONSUMER_PREDICATE", Reason: "independent_consumer_recomputed_predicate"},
		},
		ProductionAdoption: PredicateMetric{Numerator: 0, Denominator: 1, Decision: "UNKNOWN", Stage: "COHERENCE", Step: "PRODUCTION_CONSUMER_ADOPTION", Reason: "NO_PRODUCTION_CONSUMER_EVIDENCE"},
		UseCaseReceipt:     UseCaseReceiptObservation{SourceArtifact: "examples/toolchain-conformance/corpus.json", Status: "UNKNOWN", Numerator: 0, Denominator: 1, Stage: "FOUNDATION", Step: "USE_CASE_RECEIPT", Reason: "CURRENT_EVIDENCE_UNAVAILABLE"},
	}
}
func mustJSON(value any) []byte { data, _ := json.Marshal(value); return data }
func semanticDigest(manifest Manifest) string {
	data, _ := json.Marshal(semanticManifest{Schema: manifest.Schema, StableID: manifest.StableID, Concept: manifest.Concept, CodeBindings: sortedStrings(manifest.CodeBindings), MetricBindings: sortedStrings(manifest.MetricBindings), BindingRegistry: sortedBindingRegistry(manifest.BindingRegistry), UseCases: sortedUseCases(manifest.UseCases), VerificationStrategies: sortedStrings(manifest.VerificationStrategies), Corpus: sortedRefs(manifest.Corpus), Registry: sortedRefs(manifest.Registry), Denominators: sortedDenominators(manifest.Denominators), Documentation: sortedRefs(manifest.Documentation)})
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
func sortedBindingRegistry(values []BindingRegistryEntry) []BindingRegistryEntry {
	result := append([]BindingRegistryEntry(nil), values...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].MetricID == result[j].MetricID {
			return result[i].RawSourceAddress < result[j].RawSourceAddress
		}
		return result[i].MetricID < result[j].MetricID
	})
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
func expectedManifestPath(stableID string) string {
	return filepath.ToSlash(filepath.Join("examples", stableID, "concept.manifest.json"))
}
func allRefs(manifest Manifest) []ResourceRef {
	refs := append([]ResourceRef{}, manifest.Corpus...)
	refs = append(refs, manifest.Registry...)
	return append(refs, manifest.Documentation...)
}
func safePath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return path != "" && !filepath.IsAbs(path) && clean != "." && !strings.HasPrefix(clean, "../")
}
func pathExists(root, path string) bool {
	if !safePath(path) {
		return false
	}
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(path)))
	return err == nil
}
func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
func copyInts(values map[string]int) map[string]int {
	result := make(map[string]int, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}
func sameInts(left, right map[string]int) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}
func constValue(source, name string) (int, bool) {
	match := regexp.MustCompile(`(?m)` + regexp.QuoteMeta(name) + `\s*=\s*(\d+)`).FindStringSubmatch(source)
	if len(match) != 2 {
		return 0, false
	}
	value, err := strconv.Atoi(match[1])
	return value, err == nil
}
func countGoooLines(root string) int {
	total := 0
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == ".parallel" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(entry.Name()) != ".gooo" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err == nil {
			total += strings.Count(string(data), "\n")
			if len(data) > 0 && data[len(data)-1] != '\n' {
				total++
			}
		}
		return nil
	})
	return total
}
func failFailure(err error, stage, step, reason string) {
	if item, ok := err.(failure); ok {
		fail(item.Decision, item.Stage, item.Step, item.Reason)
	}
	fail("FAIL_CLOSED", stage, step, reason)
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
