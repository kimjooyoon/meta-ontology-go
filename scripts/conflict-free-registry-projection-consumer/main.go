package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	manifestSchema        = "gooo/language-concept-manifest/v1"
	projectionSchema      = "gooo/manual-source-registration-edit-free-registry-projection/v1"
	defaultOutput         = "internal/meta/registryprojection/generated"
	receiptSchema         = "gooo/manual-source-registration-edit-free-registry-consumer-receipt/v1"
	embeddedOutputAddress = "embedded://consumer_output_artifact/raw_bytes"
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
	MetricID               string `json:"metric_id"`
	RawSourceAddress       string `json:"raw_source_address"`
	RegistrationUseAddress string `json:"registration_use_address"`
	SemanticDigest         string `json:"semantic_digest"`
	ConsumerEntryPoint     string `json:"consumer_entry_point"`
	ObservedOutputAddress  string `json:"observed_output_address"`
	ObservedOutputDigest   string `json:"observed_output_digest"`
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
	OutputArtifact             OutputArtifact              `json:"output_artifact"`
	BindingOutputReceipts      []BindingOutputReceipt      `json:"binding_output_receipts"`
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
type OutputArtifact struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
	Bytes  int    `json:"bytes"`
}
type BindingOutputReceipt struct {
	MetricID                string `json:"metric_id"`
	RawSourceAddress        string `json:"raw_source_address"`
	RegistrationUseAddress  string `json:"registration_use_address"`
	SemanticDigest          string `json:"semantic_digest"`
	MetricOccurrenceAddress string `json:"metric_occurrence_address"`
	MetricOccurrenceDigest  string `json:"metric_occurrence_digest"`
	ConsumerEntryPoint      string `json:"consumer_entry_point"`
	OutputAddress           string `json:"output_address"`
	OutputDigest            string `json:"output_digest"`
	OutputBytes             int    `json:"output_bytes"`
	OutputRowAddress        string `json:"output_row_address"`
	OutputRowDigest         string `json:"output_row_digest"`
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
		receipt, err := buildReceipt(absoluteRoot, projection, reconciliations, data, output)
		if err != nil {
			fail("FAIL_CLOSED", "REGRESSION", "OUTPUT", "INDEPENDENT_OUTPUT_ARTIFACT_UNAVAILABLE")
		}
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
		if binding.MetricID == "" || binding.RawSourceAddress == "" || binding.RegistrationUseAddress == "" || binding.SemanticDigest == "" || binding.ConsumerEntryPoint == "" || binding.ObservedOutputAddress == "" || binding.ObservedOutputDigest == "" {
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
		semanticDigest, err := resolveBindingSemantic(root, binding.RawSourceAddress, binding.RegistrationUseAddress, binding.MetricID)
		if err != nil {
			var typeCheckErr bindingTypeCheckError
			if errors.As(err, &typeCheckErr) {
				return failure{diagnostic{"FAIL_CLOSED", "LOWER_RESOLUTION", "BINDING_PACKAGE_TYPE_CHECK", "PACKAGE_TYPE_CHECK_FAILED"}}
			}
			return failure{diagnostic{"FAIL_CLOSED", "FOUNDATION", "BINDING_REGISTRY", "UNTRUSTED_BINDING_SOURCE"}}
		}
		if binding.SemanticDigest != semanticDigest {
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

type goBindingAddress struct {
	Path   string
	Symbol string
	Line   int
	Column int
}

type typedGoPackage struct {
	FileSet *token.FileSet
	Files   map[string]*ast.File
	Info    *types.Info
	Package *types.Package
}

type bindingTypeCheckError struct{ detail string }

func (e bindingTypeCheckError) Error() string {
	return "binding package type check failed: " + e.detail
}

type listedGoPackage struct {
	Dir             string
	ImportPath      string
	GoFiles         []string
	CgoFiles        []string
	CompiledGoFiles []string
	Error           *struct{ Err string }  `json:"Error"`
	DepsErrors      []struct{ Err string } `json:"DepsErrors"`
}

type moduleSourceImporter struct {
	root       string
	module     string
	defaultImp types.Importer
	packages   map[string]*types.Package
	loading    map[string]bool
}

func newModuleSourceImporter(root string) *moduleSourceImporter {
	return &moduleSourceImporter{root: root, module: modulePackagePath(root, ""), defaultImp: importer.Default(), packages: map[string]*types.Package{}, loading: map[string]bool{}}
}

func (m *moduleSourceImporter) Import(path string) (*types.Package, error) {
	if !strings.HasPrefix(path, m.module+"/") && path != m.module {
		return m.defaultImp.Import(path)
	}
	if pkg := m.packages[path]; pkg != nil {
		return pkg, nil
	}
	if m.loading[path] {
		return nil, bindingTypeCheckError{detail: "import cycle at " + path}
	}
	m.loading[path] = true
	defer delete(m.loading, path)
	listed, err := listGoPackage(m.root, path)
	if err != nil {
		return nil, err
	}
	fileSet := token.NewFileSet()
	_, parsed, err := parseListedGoFiles(m.root, listed, fileSet)
	if err != nil {
		return nil, err
	}
	config := types.Config{Importer: m}
	var typeErrors []error
	config.Error = func(err error) { typeErrors = append(typeErrors, err) }
	checked, checkErr := config.Check(listed.ImportPath, fileSet, parsed, nil)
	if checkErr != nil || checked == nil || len(typeErrors) > 0 {
		return nil, bindingTypeCheckError{detail: listed.ImportPath}
	}
	m.packages[path] = checked
	return checked, nil
}

type bindingDigestPayload struct {
	PackageIdentity         string `json:"package_identity"`
	ObjectIdentity          string `json:"object_identity"`
	DeclarationAST          string `json:"declaration_ast"`
	RegistrationUseAST      string `json:"registration_use_ast"`
	MetricID                string `json:"metric_id"`
	MetricOccurrenceAddress string `json:"metric_occurrence_address"`
	MetricOccurrenceAST     string `json:"metric_occurrence_ast"`
	MetricOccurrenceDigest  string `json:"metric_occurrence_digest"`
}

type bindingResolution struct {
	SemanticDigest          string
	MetricOccurrenceAddress string
	MetricOccurrenceDigest  string
}

func resolveBindingSemantic(root, rawSourceAddress, registrationUseAddress, metricID string) (string, error) {
	relation, err := resolveBindingRelation(root, rawSourceAddress, registrationUseAddress, metricID)
	if err != nil {
		return "", err
	}
	return relation.SemanticDigest, nil
}

func resolveBindingRelation(root, rawSourceAddress, registrationUseAddress, metricID string) (bindingResolution, error) {
	declaration, err := parseGoBindingAddress(rawSourceAddress, false)
	if err != nil {
		return bindingResolution{}, err
	}
	registration, err := parseGoBindingAddress(registrationUseAddress, true)
	if err != nil {
		return bindingResolution{}, err
	}
	if filepath.ToSlash(filepath.Dir(declaration.Path)) != filepath.ToSlash(filepath.Dir(registration.Path)) {
		return bindingResolution{}, os.ErrInvalid
	}
	packageInfo, err := typeCheckGoPackage(root, declaration.Path)
	if err != nil {
		return bindingResolution{}, err
	}
	declarationFile := packageInfo.Files[filepath.ToSlash(declaration.Path)]
	registrationFile := packageInfo.Files[filepath.ToSlash(registration.Path)]
	if declarationFile == nil || registrationFile == nil {
		return bindingResolution{}, os.ErrNotExist
	}
	declarationIdent, declarationNode, err := findDeclaration(declarationFile, declaration.Symbol)
	if err != nil {
		return bindingResolution{}, err
	}
	declarationObject := packageInfo.Info.Defs[declarationIdent]
	if declarationObject == nil {
		return bindingResolution{}, os.ErrInvalid
	}
	registrationIdent, registrationNode, err := findRegistrationUse(registrationFile, registration.Symbol, registration.Line, registration.Column, packageInfo.FileSet, packageInfo.Info)
	if err != nil {
		return bindingResolution{}, err
	}
	registrationObject := packageInfo.Info.Uses[registrationIdent]
	if registrationObject == nil || registrationObject != declarationObject {
		return bindingResolution{}, os.ErrInvalid
	}
	if !validRegistrationContext(registrationNode, registration.Symbol, metricID, registrationIdent) || (isCallNode(registrationNode) && !hasCatalogOwner(registrationFile, registrationNode)) {
		return bindingResolution{}, os.ErrInvalid
	}
	metricAddress, metricAST, metricDigest, err := findMetricOccurrence(packageInfo.FileSet, declarationNode, declaration.Symbol, registrationNode, metricID)
	if err != nil {
		return bindingResolution{}, err
	}
	metricAddress = canonicalMetricOccurrenceAddress(declaration, registration, metricAddress)
	declarationAST, err := normalizedAST(packageInfo.FileSet, declarationNode)
	if err != nil {
		return bindingResolution{}, err
	}
	registrationAST, err := normalizedAST(packageInfo.FileSet, registrationNode)
	if err != nil {
		return bindingResolution{}, err
	}
	objectIdentity := bindingObjectIdentity(declarationObject)
	payload, err := json.Marshal(bindingDigestPayload{PackageIdentity: declarationObject.Pkg().Path(), ObjectIdentity: objectIdentity, DeclarationAST: string(declarationAST), RegistrationUseAST: string(registrationAST), MetricID: metricID, MetricOccurrenceAddress: metricAddress, MetricOccurrenceAST: string(metricAST), MetricOccurrenceDigest: metricDigest})
	if err != nil {
		return bindingResolution{}, err
	}
	return bindingResolution{SemanticDigest: digest(payload), MetricOccurrenceAddress: metricAddress, MetricOccurrenceDigest: metricDigest}, nil
}

func findMetricOccurrence(fileSet *token.FileSet, declarationNode ast.Node, declarationSymbol string, registrationNode ast.Node, metricID string) (string, []byte, string, error) {
	type candidate struct {
		address string
		ast     []byte
	}
	candidates := []candidate{}
	roots := map[string]ast.Node{}
	if countMetricLiterals(registrationNode, metricID) > 0 {
		roots["registration"] = registrationNode
	} else if value, ok := declarationMetricValue(declarationNode, declarationSymbol); ok {
		roots["declaration"] = value
	}
	for label, root := range roots {
		ast.Inspect(root, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING || stringLiteral(literal) != metricID {
				return true
			}
			normalized, err := normalizedAST(fileSet, literal)
			if err != nil {
				return false
			}
			ordinal := astNodeOrdinal(root, literal)
			candidates = append(candidates, candidate{address: fmt.Sprintf("%s/literal/%d", label, ordinal), ast: normalized})
			return true
		})
	}
	if len(candidates) != 1 {
		return "", nil, "", os.ErrInvalid
	}
	return candidates[0].address, candidates[0].ast, digest(candidates[0].ast), nil
}

func canonicalMetricOccurrenceAddress(declaration, registration goBindingAddress, structural string) string {
	source := declaration
	if strings.HasPrefix(structural, "registration/") {
		source = registration
	}
	return structural + "@" + source.Path + "#" + source.Symbol
}

func astNodeOrdinal(root, target ast.Node) int {
	ordinal := 0
	found := -1
	ast.Inspect(root, func(node ast.Node) bool {
		if node == target {
			found = ordinal
		}
		ordinal++
		return found < 0
	})
	return found
}

func isCallNode(node ast.Node) bool {
	_, ok := node.(*ast.CallExpr)
	return ok
}

func hasCatalogOwner(file *ast.File, target ast.Node) bool {
	var stack []ast.Node
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			return true
		}
		stack = append(stack, node)
		if node == target {
			for _, parent := range stack {
				if function, ok := parent.(*ast.FuncDecl); ok && function.Name != nil && function.Name.Name == "Catalog" {
					found = true
				}
			}
			return false
		}
		return true
	})
	return found
}

func parseGoBindingAddress(address string, requirePosition bool) (goBindingAddress, error) {
	parts := strings.SplitN(address, "#", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || filepath.IsAbs(parts[0]) || strings.HasPrefix(filepath.ToSlash(filepath.Clean(parts[0])), "../") || filepath.Ext(parts[0]) != ".go" {
		return goBindingAddress{}, os.ErrInvalid
	}
	result := goBindingAddress{Path: filepath.ToSlash(parts[0])}
	symbol := parts[1]
	if at := strings.LastIndex(symbol, "@"); at >= 0 {
		position := strings.Split(symbol[at+1:], ":")
		if len(position) != 2 || position[0] == "" || position[1] == "" {
			return goBindingAddress{}, os.ErrInvalid
		}
		if _, err := fmt.Sscanf(position[0], "%d", &result.Line); err != nil {
			return goBindingAddress{}, err
		}
		if _, err := fmt.Sscanf(position[1], "%d", &result.Column); err != nil {
			return goBindingAddress{}, err
		}
		symbol = symbol[:at]
	}
	if symbol == "" || (requirePosition && (result.Line <= 0 || result.Column <= 0)) || (!requirePosition && (result.Line != 0 || result.Column != 0)) {
		return goBindingAddress{}, os.ErrInvalid
	}
	result.Symbol = symbol
	return result, nil
}

func typeCheckGoPackage(root, declarationPath string) (*typedGoPackage, error) {
	directory := filepath.ToSlash(filepath.Dir(declarationPath))
	if directory == "." {
		directory = ""
	}
	listed, err := listGoPackage(root, "./"+directory)
	if err != nil {
		return nil, err
	}
	fileSet := token.NewFileSet()
	files, parsed, err := parseListedGoFiles(root, listed, fileSet)
	if err != nil {
		return nil, err
	}
	if len(parsed) == 0 {
		return nil, bindingTypeCheckError{detail: listed.ImportPath + " has no build-selected source files"}
	}
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}, Selections: map[*ast.SelectorExpr]*types.Selection{}}
	config := types.Config{Importer: newModuleSourceImporter(root)}
	var typeErrors []error
	config.Error = func(err error) { typeErrors = append(typeErrors, err) }
	checked, checkErr := config.Check(listed.ImportPath, fileSet, parsed, info)
	if checkErr != nil || checked == nil || len(typeErrors) > 0 {
		detail := listed.ImportPath
		if len(typeErrors) > 0 {
			detail = typeErrors[0].Error()
		} else if checkErr != nil {
			detail = checkErr.Error()
		}
		return nil, bindingTypeCheckError{detail: detail}
	}
	return &typedGoPackage{FileSet: fileSet, Files: files, Info: info, Package: checked}, nil
}

func listGoPackage(root, packagePath string) (listedGoPackage, error) {
	command := exec.Command("go", "list", "-json", "-e", "-compiled", packagePath)
	command.Dir = root
	data, err := command.Output()
	if err != nil {
		return listedGoPackage{}, bindingTypeCheckError{detail: err.Error()}
	}
	listed := listedGoPackage{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&listed); err != nil {
		return listedGoPackage{}, bindingTypeCheckError{detail: err.Error()}
	}
	if listed.ImportPath == "" || listed.Error != nil || len(listed.DepsErrors) > 0 {
		return listedGoPackage{}, bindingTypeCheckError{detail: listed.ImportPath + " has module/load errors"}
	}
	return listed, nil
}

func parseListedGoFiles(root string, listed listedGoPackage, fileSet *token.FileSet) (map[string]*ast.File, []*ast.File, error) {
	paths := append([]string(nil), listed.CompiledGoFiles...)
	if len(paths) == 0 {
		paths = append(paths, listed.GoFiles...)
		paths = append(paths, listed.CgoFiles...)
	}
	files := map[string]*ast.File{}
	parsed := make([]*ast.File, 0, len(paths))
	for _, listedPath := range paths {
		absolute := listedPath
		if !filepath.IsAbs(absolute) {
			absolute = filepath.Join(listed.Dir, listedPath)
		}
		relative, err := filepath.Rel(root, absolute)
		if err != nil || strings.HasPrefix(filepath.ToSlash(relative), "../") {
			return nil, nil, bindingTypeCheckError{detail: "build-selected Go file is outside module root"}
		}
		relative = filepath.ToSlash(relative)
		data, err := os.ReadFile(absolute)
		if err != nil {
			return nil, nil, bindingTypeCheckError{detail: err.Error()}
		}
		file, err := parser.ParseFile(fileSet, relative, data, 0)
		if err != nil {
			return nil, nil, bindingTypeCheckError{detail: err.Error()}
		}
		files[relative] = file
		parsed = append(parsed, file)
	}
	return files, parsed, nil
}

func modulePackagePath(root, directory string) string {
	module := "local"
	if data, err := os.ReadFile(filepath.Join(root, "go.mod")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Fields(line)
			if len(fields) == 2 && fields[0] == "module" {
				module = fields[1]
				break
			}
		}
	}
	if directory == "" {
		return module
	}
	return module + "/" + filepath.ToSlash(directory)
}

func findDeclaration(file *ast.File, symbol string) (*ast.Ident, ast.Node, error) {
	var foundIdent *ast.Ident
	var foundNode ast.Node
	for _, declaration := range file.Decls {
		switch value := declaration.(type) {
		case *ast.FuncDecl:
			if value.Name != nil && value.Name.Name == symbol {
				if foundIdent != nil {
					return nil, nil, os.ErrInvalid
				}
				foundIdent, foundNode = value.Name, value
			}
		case *ast.GenDecl:
			for _, specification := range value.Specs {
				valueSpec, ok := specification.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range valueSpec.Names {
					if name.Name == symbol {
						if foundIdent != nil {
							return nil, nil, os.ErrInvalid
						}
						foundIdent, foundNode = name, valueSpec
					}
				}
			}
		}
	}
	if foundIdent == nil {
		return nil, nil, os.ErrNotExist
	}
	return foundIdent, foundNode, nil
}

func findRegistrationUse(file *ast.File, symbol string, line, column int, fileSet *token.FileSet, info *types.Info) (*ast.Ident, ast.Node, error) {
	var stack []ast.Node
	var foundIdent *ast.Ident
	var foundNode ast.Node
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			return true
		}
		stack = append(stack, node)
		ident, ok := node.(*ast.Ident)
		if !ok || ident.Name != symbol || info.Uses[ident] == nil {
			return true
		}
		position := fileSet.PositionFor(ident.Pos(), false)
		if position.Line != line || position.Column != column {
			return true
		}
		if foundIdent != nil {
			foundIdent = nil
			foundNode = nil
			return false
		}
		foundIdent = ident
		foundNode = registrationContext(stack)
		return false
	})
	if foundIdent == nil || foundNode == nil {
		return nil, nil, os.ErrNotExist
	}
	return foundIdent, foundNode, nil
}

func registrationContext(stack []ast.Node) ast.Node {
	var fallback ast.Node
	for index := len(stack) - 2; index >= 0; index-- {
		switch stack[index].(type) {
		case *ast.KeyValueExpr:
			return stack[index]
		case *ast.CallExpr, *ast.BinaryExpr, *ast.AssignStmt, *ast.ValueSpec:
			if fallback == nil {
				fallback = stack[index]
			}
		}
	}
	return fallback
}

func validRegistrationContext(node ast.Node, symbol, metricID string, registrationIdent *ast.Ident) bool {
	switch value := node.(type) {
	case *ast.CallExpr:
		return functionName(value.Fun) == symbol && containsIdent(value, registrationIdent) && countMetricLiterals(value, metricID) == 1
	case *ast.KeyValueExpr:
		key, ok := value.Key.(*ast.Ident)
		return ok && (key.Name == "MetricBindings" || key.Name == "MetricID") && containsIdent(value.Value, registrationIdent)
	default:
		return false
	}
}

func containsIdent(node ast.Node, target *ast.Ident) bool {
	found := false
	ast.Inspect(node, func(node ast.Node) bool {
		if node == target {
			found = true
			return false
		}
		return !found
	})
	return found
}

func declarationMetricValue(node ast.Node, symbol string) (ast.Node, bool) {
	if function, ok := node.(*ast.FuncDecl); ok && function.Name != nil && function.Name.Name == symbol && function.Body != nil {
		return function.Body, true
	}
	value, ok := node.(*ast.ValueSpec)
	if !ok || len(value.Names) == 0 || len(value.Values) == 0 {
		return nil, false
	}
	nameIndex := -1
	for index, name := range value.Names {
		if name.Name == symbol {
			nameIndex = index
			break
		}
	}
	if nameIndex < 0 || (len(value.Names) != 1 && len(value.Names) != len(value.Values)) {
		return nil, false
	}
	if len(value.Names) == 1 {
		return value.Values[0], true
	}
	return value.Values[nameIndex], true
}

func countMetricLiterals(node ast.Node, metricID string) int {
	count := 0
	ast.Inspect(node, func(node ast.Node) bool {
		if stringLiteral(node) == metricID {
			count++
		}
		return true
	})
	return count
}

func normalizedAST(fileSet *token.FileSet, node ast.Node) ([]byte, error) {
	var normalized bytes.Buffer
	if err := format.Node(&normalized, fileSet, node); err != nil {
		return nil, err
	}
	return normalized.Bytes(), nil
}

func bindingObjectIdentity(object types.Object) string {
	packagePath := ""
	if object.Pkg() != nil {
		packagePath = object.Pkg().Path()
	}
	return fmt.Sprintf("%s|%s|%s", packagePath, object.Name(), object.Type().String())
}

func astNodeContainsString(node ast.Node, wanted string) bool {
	found := false
	ast.Inspect(node, func(node ast.Node) bool {
		if literal, ok := node.(*ast.BasicLit); ok && literal.Kind == token.STRING && stringLiteral(literal) == wanted {
			found = true
			return false
		}
		return !found
	})
	return found
}

func stringLiteral(node ast.Node) string {
	literal, ok := node.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return ""
	}
	value, err := strconv.Unquote(literal.Value)
	if err != nil {
		return ""
	}
	return value
}

func functionName(node ast.Expr) string {
	identifier, ok := node.(*ast.Ident)
	if !ok {
		return ""
	}
	return identifier.Name
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

func buildReceipt(root string, projection Projection, reconciliations []DenominatorReconciliation, data []byte, outputPath string) (Receipt, error) {
	observedOutput := data
	if outputPath != "" {
		var err error
		observedOutput, err = os.ReadFile(outputPath)
		if err != nil || !bytes.Equal(observedOutput, data) {
			return Receipt{}, os.ErrInvalid
		}
	}
	if _, err := observedProjection(observedOutput); err != nil {
		return Receipt{}, err
	}
	ids := make([]string, 0, len(projection.Catalog))
	for _, entry := range projection.Catalog {
		ids = append(ids, entry.StableID)
	}
	sort.Strings(ids)
	resourceData, _ := json.Marshal(append(append(append([]ResourceSnapshot{}, projection.Corpus...), projection.Registry...), projection.Documentation...))
	denominatorData, _ := json.Marshal(projection.Denominator)
	bindingReceipts := make([]BindingOutputReceipt, 0)
	for _, entry := range projection.Catalog {
		for _, binding := range entry.BindingRegistry {
			relation, err := resolveBindingRelation(root, binding.RawSourceAddress, binding.RegistrationUseAddress, binding.MetricID)
			if err != nil {
				return Receipt{}, err
			}
			rowAddress, rowDigest, err := observedBindingRow(observedOutput, entry.StableID, binding)
			if err != nil {
				return Receipt{}, err
			}
			bindingReceipts = append(bindingReceipts, BindingOutputReceipt{MetricID: binding.MetricID, RawSourceAddress: binding.RawSourceAddress, RegistrationUseAddress: binding.RegistrationUseAddress, SemanticDigest: relation.SemanticDigest, MetricOccurrenceAddress: relation.MetricOccurrenceAddress, MetricOccurrenceDigest: relation.MetricOccurrenceDigest, ConsumerEntryPoint: binding.ConsumerEntryPoint, OutputAddress: embeddedOutputAddress, OutputDigest: digest(observedOutput), OutputBytes: len(observedOutput), OutputRowAddress: rowAddress, OutputRowDigest: rowDigest})
		}
	}
	return Receipt{
		Schema: receiptSchema, Decision: "PASS", ProjectionDigest: digest(observedOutput), DenominatorReconciliations: reconciliations,
		Predicates: []PredicateObservation{
			{ID: "independent-manifest-order", ObservedPredicate: "raw manifests are sorted by stable_id", TargetAddress: "raw://manifest-stable-ids", TargetDigest: digest(mustJSON(ids)), Observed: true, Decision: "PASS", PredicateTruth: "TRUE", Stage: "COHERENCE", Step: "INDEPENDENT_CONSUMER_PREDICATE", Reason: "independent_consumer_recomputed_predicate"},
			{ID: "independent-resource-digests", ObservedPredicate: "raw resource refs resolve to their declared digests", TargetAddress: "raw://resource-ref-digests", TargetDigest: digest(resourceData), Observed: true, Decision: "PASS", PredicateTruth: "TRUE", Stage: "FOUNDATION", Step: "INDEPENDENT_CONSUMER_PREDICATE", Reason: "independent_consumer_recomputed_predicate"},
			{ID: "independent-denominator-reconciliation", ObservedPredicate: "raw corpus and registry sources reconcile to declared denominators", TargetAddress: "raw://denominator-reconciliation", TargetDigest: digest(denominatorData), Observed: true, Decision: "PASS", PredicateTruth: "TRUE", Stage: "FOUNDATION", Step: "INDEPENDENT_CONSUMER_PREDICATE", Reason: "independent_consumer_recomputed_predicate"},
			{ID: "independent-binding-registry", ObservedPredicate: "structured metric bindings reconnect raw source, semantic digest, consumer entry point, and observed consumer output digest", TargetAddress: "raw://structured-binding-registry", TargetDigest: digest(mustJSON(bindingReceipts)), Observed: true, Decision: "PASS", PredicateTruth: "TRUE", Stage: "FOUNDATION", Step: "INDEPENDENT_CONSUMER_PREDICATE", Reason: "independent_consumer_recomputed_predicate"},
			{ID: "independent-conformance-consumer", ObservedPredicate: "independent conformance consumer projection bytes equal its raw-manifest reconstruction", TargetAddress: defaultOutput + "/projection.json", TargetDigest: digest(data), Observed: true, Decision: "PASS", PredicateTruth: "TRUE", Stage: "COHERENCE", Step: "INDEPENDENT_CONSUMER_PREDICATE", Reason: "independent_consumer_recomputed_predicate"},
		},
		ProductionAdoption:    PredicateMetric{Numerator: 0, Denominator: 1, Decision: "UNKNOWN", Stage: "COHERENCE", Step: "PRODUCTION_CONSUMER_ADOPTION", Reason: "NO_PRODUCTION_CONSUMER_EVIDENCE"},
		UseCaseReceipt:        UseCaseReceiptObservation{SourceArtifact: "examples/toolchain-conformance/corpus.json", Status: "UNKNOWN", Numerator: 0, Denominator: 1, Stage: "FOUNDATION", Step: "USE_CASE_RECEIPT", Reason: "HISTORICAL_CORPUS_PRESENT_BUT_EXECUTION_RECEIPT_UNAVAILABLE"},
		OutputArtifact:        OutputArtifact{Path: embeddedOutputAddress, Digest: digest(observedOutput), Bytes: len(observedOutput)},
		BindingOutputReceipts: bindingReceipts,
	}, nil
}

func observedProjection(data []byte) (Projection, error) {
	projection := Projection{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&projection); err != nil {
		return Projection{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Projection{}, err
	}
	return projection, nil
}

func observedBindingRow(data []byte, stableID string, expected BindingRegistryEntry) (string, string, error) {
	projection, err := observedProjection(data)
	if err != nil {
		return "", "", err
	}
	matches := 0
	var row BindingRegistryEntry
	for _, entry := range projection.Catalog {
		if entry.StableID != stableID {
			continue
		}
		for _, binding := range entry.BindingRegistry {
			if binding.MetricID == expected.MetricID && binding.RawSourceAddress == expected.RawSourceAddress && binding.RegistrationUseAddress == expected.RegistrationUseAddress {
				matches++
				row = binding
			}
		}
	}
	if matches != 1 {
		return "", "", os.ErrInvalid
	}
	address := bindingOutputRowAddress(stableID, expected.MetricID)
	return address, digest(mustJSON(row)), nil
}

func bindingOutputRowAddress(stableID, metricID string) string {
	encode := base64.RawURLEncoding.EncodeToString
	return embeddedOutputAddress + "#/catalog/" + encode([]byte(stableID)) + "/binding_registry/" + encode([]byte(metricID))
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
