package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type denominatorSource struct {
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

func validateManifestInputs(root string, loaded []LoadedManifest, requiredIDs []string) *Diagnostic {
	if diagnostic := validateManifests(loaded, requiredIDs); diagnostic != nil {
		return diagnostic
	}
	for _, item := range loaded {
		if item.SourcePath != expectedManifestPath(item.Manifest.StableID) {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "MANIFEST_OWNERSHIP", Reason: "CROSS_DIRECTORY_MANIFEST"}
		}
		for _, binding := range item.Manifest.CodeBindings {
			if !pathExists(root, binding) {
				return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "CODE_BINDING", Reason: "MISSING_CODE_BINDING"}
			}
		}
		if diagnostic := validateBindingRegistry(root, item.Manifest); diagnostic != nil {
			return diagnostic
		}
		for _, ref := range allRefs(item.Manifest) {
			data, err := readSource(root, ref.Path)
			if err != nil {
				return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "RESOURCE_AVAILABILITY", Reason: "MISSING_RESOURCE"}
			}
			if digestBytes(data) != ref.Digest {
				return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "RESOURCE_DIGEST", Reason: "RESOURCE_DIGEST_MISMATCH"}
			}
		}
	}
	_, diagnostic := reconcileDenominators(root, loaded)
	return diagnostic
}

func expectedManifestPath(stableID string) string {
	return filepath.ToSlash(filepath.Join("examples", stableID, "concept.manifest.json"))
}

func allRefs(manifest Manifest) []ResourceRef {
	refs := make([]ResourceRef, 0, len(manifest.Corpus)+len(manifest.Registry)+len(manifest.Documentation))
	refs = append(refs, manifest.Corpus...)
	refs = append(refs, manifest.Registry...)
	refs = append(refs, manifest.Documentation...)
	return refs
}

func pathExists(root, path string) bool {
	if path == "" || filepath.IsAbs(path) || strings.HasPrefix(filepath.ToSlash(filepath.Clean(path)), "../") {
		return false
	}
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(path)))
	return err == nil
}

func validateBindingRegistry(root string, manifest Manifest) *Diagnostic {
	if len(manifest.BindingRegistry) == 0 {
		return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "BINDING_REGISTRY", Reason: "MISSING_STRUCTURED_BINDING"}
	}
	metricIDs := make(map[string]struct{}, len(manifest.MetricBindings))
	for _, metricID := range manifest.MetricBindings {
		metricIDs[metricID] = struct{}{}
	}
	seen := map[string]struct{}{}
	for _, binding := range manifest.BindingRegistry {
		if binding.MetricID == "" || binding.RawSourceAddress == "" || binding.SemanticDigest == "" || binding.ConsumerEntryPoint == "" || binding.ObservedOutputAddress == "" || binding.ObservedOutputDigest == "" {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "BINDING_REGISTRY", Reason: "INCOMPLETE_STRUCTURED_BINDING"}
		}
		if _, ok := metricIDs[binding.MetricID]; !ok {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "BINDING_REGISTRY", Reason: "UNREGISTERED_STRUCTURED_BINDING"}
		}
		if _, ok := seen[binding.MetricID]; ok {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "BINDING_REGISTRY", Reason: "DUPLICATE_STRUCTURED_BINDING"}
		}
		seen[binding.MetricID] = struct{}{}
		parts := strings.SplitN(binding.RawSourceAddress, "#", 2)
		if len(parts) != 2 || !pathExists(root, parts[0]) || filepath.Ext(parts[0]) != ".go" {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "BINDING_REGISTRY", Reason: "UNTRUSTED_BINDING_SOURCE"}
		}
		semanticDigest, err := resolveBindingSemantic(root, binding.RawSourceAddress, binding.MetricID)
		if err != nil {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "BINDING_REGISTRY", Reason: "UNTRUSTED_BINDING_SOURCE"}
		}
		if binding.SemanticDigest != semanticDigest {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "BINDING_REGISTRY", Reason: "BINDING_SEMANTIC_DIGEST_MISMATCH"}
		}
		if binding.ConsumerEntryPoint != "scripts/conflict-free-registry-projection-consumer/main.go" || !pathExists(root, binding.ConsumerEntryPoint) {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "BINDING_REGISTRY", Reason: "MISSING_CONSUMER_ENTRY_POINT"}
		}
		matched := false
		for _, ref := range allRefs(manifest) {
			if ref.Path == binding.ObservedOutputAddress {
				matched = true
				if ref.Digest != binding.ObservedOutputDigest {
					return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "BINDING_REGISTRY", Reason: "BINDING_OUTPUT_DIGEST_MISMATCH"}
				}
			}
		}
		if !matched {
			return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "BINDING_REGISTRY", Reason: "UNBOUND_OBSERVED_OUTPUT"}
		}
	}
	if len(seen) != len(metricIDs) {
		return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "BINDING_REGISTRY", Reason: "MISSING_STRUCTURED_BINDING"}
	}
	return nil
}

func resolveBindingSemantic(root, rawSourceAddress, metricID string) (string, error) {
	parts := strings.SplitN(rawSourceAddress, "#", 2)
	if len(parts) != 2 || parts[1] == "" {
		return "", fmt.Errorf("invalid binding address")
	}
	data, err := readSource(root, parts[0])
	if err != nil {
		return "", err
	}
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, parts[0], data, 0)
	if err != nil {
		return "", err
	}
	var matched ast.Node
	ast.Inspect(file, func(node ast.Node) bool {
		if matched != nil || node == nil {
			return matched == nil
		}
		switch value := node.(type) {
		case *ast.ValueSpec:
			for _, name := range value.Names {
				if name.Name == parts[1] && astNodeContainsString(value, metricID) {
					matched = value
					return false
				}
			}
		case *ast.FuncDecl:
			if value.Name != nil && value.Name.Name == parts[1] && astNodeContainsString(value, metricID) {
				matched = value
				return false
			}
		case *ast.CallExpr:
			if functionName(value.Fun) == "concept" && len(value.Args) > 0 && stringLiteral(value.Args[0]) == parts[1] && astNodeContainsString(value, metricID) {
				matched = value
				return false
			}
		}
		return true
	})
	if matched == nil {
		return "", fmt.Errorf("binding symbol not resolved")
	}
	if isGoIdentifier(parts[1]) && symbolUseCount(root, parts[1]) < 2 {
		return "", fmt.Errorf("binding symbol is not registered or used")
	}
	var normalized bytes.Buffer
	if err := format.Node(&normalized, fileSet, matched); err != nil {
		return "", err
	}
	return digestBytes(normalized.Bytes()), nil
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

func isGoIdentifier(value string) bool {
	if value == "" || !(value[0] >= 'a' && value[0] <= 'z') && !(value[0] >= 'A' && value[0] <= 'Z') && value[0] != '_' {
		return false
	}
	for _, character := range value[1:] {
		if !(character >= 'a' && character <= 'z') && !(character >= 'A' && character <= 'Z') && !(character >= '0' && character <= '9') && character != '_' {
			return false
		}
	}
	return true
}

func symbolUseCount(root, symbol string) int {
	count := 0
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
		if filepath.Ext(entry.Name()) != ".go" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, data, 0)
		if err != nil {
			return nil
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if identifier, ok := node.(*ast.Ident); ok && identifier.Name == symbol {
				count++
			}
			return true
		})
		return nil
	})
	return count
}

func reconcileDenominators(root string, loaded []LoadedManifest) ([]DenominatorReconciliation, *Diagnostic) {
	receipts := make([]DenominatorReconciliation, 0)
	for _, item := range sortedLoaded(loaded) {
		for _, denominator := range sortedDenominators(item.Manifest.Denominators) {
			calculated, err := calculateDenominator(root, item, denominator)
			if err != nil {
				return receipts, &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "DENOMINATOR_SOURCE", Reason: "DENOMINATOR_SOURCE_UNAVAILABLE"}
			}
			declared := copyInts(denominator.Values)
			receipt := DenominatorReconciliation{StableID: item.Manifest.StableID, DenominatorID: denominator.ID, Declared: declared, Calculated: copyInts(calculated), Decision: "PASS", Reason: "DENOMINATOR_SOURCE_RECONCILED"}
			if !sameInts(declared, calculated) {
				receipt.Decision = "FAIL_CLOSED"
				receipt.Reason = "DENOMINATOR_SOURCE_MISMATCH"
				receipts = append(receipts, receipt)
				return receipts, &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "DENOMINATOR_RECONCILIATION", Reason: "DENOMINATOR_SOURCE_MISMATCH"}
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
	corpusRaw, err := readSource(root, item.Manifest.Corpus[0].Path)
	if err != nil {
		return nil, err
	}
	source := denominatorSource{}
	if err := json.Unmarshal(corpusRaw, &source); err != nil {
		return nil, err
	}
	calculated := make(map[string]int)
	switch item.Manifest.StableID {
	case "language-syntax-roundtrip":
		calculated["cases"] = len(source.Cases)
		for _, item := range source.Cases {
			switch item.Kind {
			case "VALID":
				calculated["valid_sources"]++
			case "INVALID":
				calculated["invalid_cases"]++
			}
		}
		calculated["gooo_lines"] = countGoooLines(root)
		if len(item.Manifest.Registry) > 0 {
			registry, err := readSource(root, item.Manifest.Registry[0].Path)
			if err != nil {
				return nil, err
			}
			valid := strings.Count(string(registry), `valid("`) - strings.Count(string(registry), `invalid("`)
			invalid := strings.Count(string(registry), `invalid("`)
			calculated["registry_cases"] = valid + invalid
			calculated["registry_valid_sources"] = valid
			calculated["registry_invalid_cases"] = invalid
		}
	case "language-semantic-model":
		if denominator.ID == "gooo/language-semantic-binding/v1" {
			raw, err := os.ReadFile(filepath.Join(root, "examples/language-syntax-roundtrip/corpus.json"))
			if err != nil {
				return nil, err
			}
			syntax := denominatorSource{}
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
		calculated["cases"] = len(source.Cases)
		for _, item := range source.Cases {
			switch item.Kind {
			case "SOURCE":
				calculated["source_units"]++
			case "LAW":
				calculated["authority_laws"]++
			case "UPSTREAM_REJECTION":
				calculated["upstream_rejections"]++
			}
		}
		if len(item.Manifest.Registry) > 0 {
			registry, err := readSource(root, item.Manifest.Registry[0].Path)
			if err != nil {
				return nil, err
			}
			for _, key := range []string{"expectedSources", "expectedLaws", "expectedRejections"} {
				value, ok := constValue(string(registry), key)
				if ok {
					calculated["registry_"+strings.TrimPrefix(strings.ToLower(key), "expected")] = value
				}
			}
		}
	case "toolchain-conformance":
		calculated["surfaces"] = len(source.Surfaces)
		for _, surface := range source.Surfaces {
			calculated["cases"] += surface.Cases
			calculated["indicators"] += surface.Indicators
			calculated["proofs"] += surface.Proofs
		}
		calculated["tamper_cases"] = len(source.TamperCases)
	default:
		return values, nil
	}
	return calculated, nil
}

func observeUseCaseReceipt(root string) UseCaseReceiptObservation {
	artifact := "examples/toolchain-conformance/corpus.json"
	raw, err := readSource(root, artifact)
	if err != nil {
		return UseCaseReceiptObservation{SourceArtifact: artifact, Status: "UNKNOWN", Numerator: 0, Denominator: 1, Stage: "FOUNDATION", Step: "USE_CASE_RECEIPT", Reason: "HISTORICAL_CORPUS_PRESENT_BUT_EXECUTION_RECEIPT_UNAVAILABLE"}
	}
	var machineArtifact struct {
		Cases []json.RawMessage `json:"cases"`
	}
	if json.Unmarshal(raw, &machineArtifact) != nil {
		return UseCaseReceiptObservation{SourceArtifact: artifact, Status: "UNKNOWN", Numerator: 0, Denominator: 1, Stage: "FOUNDATION", Step: "USE_CASE_RECEIPT", Reason: "HISTORICAL_CORPUS_PRESENT_BUT_EXECUTION_RECEIPT_UNAVAILABLE"}
	}
	return UseCaseReceiptObservation{SourceArtifact: artifact, Status: "UNKNOWN", Numerator: 0, Denominator: 1, Stage: "FOUNDATION", Step: "USE_CASE_RECEIPT", Reason: "HISTORICAL_CORPUS_PRESENT_BUT_EXECUTION_RECEIPT_UNAVAILABLE"}
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
	lines := 0
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
			lines += strings.Count(string(data), "\n")
			if len(data) > 0 && data[len(data)-1] != '\n' {
				lines++
			}
		}
		return nil
	})
	return lines
}

func copyInts(values map[string]int) map[string]int {
	copyOf := make(map[string]int, len(values))
	for key, value := range values {
		copyOf[key] = value
	}
	return copyOf
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

func formatDenominatorMismatch(item DenominatorReconciliation) string {
	return fmt.Sprintf("%s declared=%v calculated=%v", item.StableID, item.Declared, item.Calculated)
}
