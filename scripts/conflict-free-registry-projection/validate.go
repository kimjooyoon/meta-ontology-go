package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
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
		if binding.MetricID == "" || binding.RawSourceAddress == "" || binding.RegistrationUseAddress == "" || binding.SemanticDigest == "" || binding.ConsumerEntryPoint == "" || binding.ObservedOutputAddress == "" || binding.ObservedOutputDigest == "" {
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
		semanticDigest, err := resolveBindingSemantic(root, binding.RawSourceAddress, binding.RegistrationUseAddress, binding.MetricID)
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

type bindingDigestPayload struct {
	PackageIdentity    string `json:"package_identity"`
	ObjectIdentity     string `json:"object_identity"`
	DeclarationAST     string `json:"declaration_ast"`
	RegistrationUseAST string `json:"registration_use_ast"`
}

func resolveBindingSemantic(root, rawSourceAddress, registrationUseAddress, metricID string) (string, error) {
	declaration, err := parseGoBindingAddress(rawSourceAddress, false)
	if err != nil {
		return "", err
	}
	registration, err := parseGoBindingAddress(registrationUseAddress, true)
	if err != nil {
		return "", err
	}
	if filepath.ToSlash(filepath.Dir(declaration.Path)) != filepath.ToSlash(filepath.Dir(registration.Path)) {
		return "", fmt.Errorf("binding declaration and use are in different packages")
	}
	packageInfo, err := typeCheckGoPackage(root, declaration.Path)
	if err != nil {
		return "", err
	}
	declarationFile := packageInfo.Files[filepath.ToSlash(declaration.Path)]
	registrationFile := packageInfo.Files[filepath.ToSlash(registration.Path)]
	if declarationFile == nil || registrationFile == nil {
		return "", fmt.Errorf("binding source file is not in typed package")
	}
	declarationIdent, declarationNode, err := findDeclaration(declarationFile, declaration.Symbol)
	if err != nil {
		return "", err
	}
	declarationObject := packageInfo.Info.Defs[declarationIdent]
	if declarationObject == nil {
		return "", fmt.Errorf("binding declaration has no types object")
	}
	registrationIdent, registrationNode, err := findRegistrationUse(registrationFile, registration.Symbol, registration.Line, registration.Column, packageInfo.FileSet, packageInfo.Info)
	if err != nil {
		return "", err
	}
	registrationObject := packageInfo.Info.Uses[registrationIdent]
	if registrationObject == nil || registrationObject != declarationObject {
		return "", fmt.Errorf("binding registration use resolves to a different object")
	}
	if !validRegistrationContext(registrationNode, registration.Symbol, metricID) || (!astNodeContainsString(declarationNode, metricID) && !astNodeContainsString(registrationNode, metricID)) {
		return "", fmt.Errorf("binding registration use is not metric-bearing")
	}
	declarationAST, err := normalizedAST(packageInfo.FileSet, declarationNode)
	if err != nil {
		return "", err
	}
	registrationAST, err := normalizedAST(packageInfo.FileSet, registrationNode)
	if err != nil {
		return "", err
	}
	objectIdentity := bindingObjectIdentity(packageInfo.FileSet, declarationObject)
	payload, err := json.Marshal(bindingDigestPayload{PackageIdentity: declarationObject.Pkg().Path(), ObjectIdentity: objectIdentity, DeclarationAST: string(declarationAST), RegistrationUseAST: string(registrationAST)})
	if err != nil {
		return "", err
	}
	return digestBytes(payload), nil
}

func parseGoBindingAddress(address string, requirePosition bool) (goBindingAddress, error) {
	parts := strings.SplitN(address, "#", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || filepath.IsAbs(parts[0]) || strings.HasPrefix(filepath.ToSlash(filepath.Clean(parts[0])), "../") || filepath.Ext(parts[0]) != ".go" {
		return goBindingAddress{}, fmt.Errorf("invalid Go binding address")
	}
	result := goBindingAddress{Path: filepath.ToSlash(parts[0])}
	symbol := parts[1]
	if at := strings.LastIndex(symbol, "@"); at >= 0 {
		position := strings.Split(symbol[at+1:], ":")
		if len(position) != 2 || position[0] == "" || position[1] == "" {
			return goBindingAddress{}, fmt.Errorf("invalid Go binding use position")
		}
		if _, err := fmt.Sscanf(position[0], "%d", &result.Line); err != nil {
			return goBindingAddress{}, fmt.Errorf("invalid Go binding use line")
		}
		if _, err := fmt.Sscanf(position[1], "%d", &result.Column); err != nil {
			return goBindingAddress{}, fmt.Errorf("invalid Go binding use column")
		}
		symbol = symbol[:at]
	}
	if symbol == "" || (requirePosition && (result.Line <= 0 || result.Column <= 0)) || (!requirePosition && (result.Line != 0 || result.Column != 0)) {
		return goBindingAddress{}, fmt.Errorf("invalid Go binding symbol address")
	}
	result.Symbol = symbol
	return result, nil
}

func typeCheckGoPackage(root, declarationPath string) (*typedGoPackage, error) {
	directory := filepath.ToSlash(filepath.Dir(declarationPath))
	if directory == "." {
		directory = ""
	}
	absoluteDirectory := filepath.Join(root, filepath.FromSlash(directory))
	entries, err := os.ReadDir(absoluteDirectory)
	if err != nil {
		return nil, err
	}
	fileSet := token.NewFileSet()
	files := map[string]*ast.File{}
	parsed := make([]*ast.File, 0)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		relative := filepath.ToSlash(filepath.Join(directory, entry.Name()))
		data, err := os.ReadFile(filepath.Join(absoluteDirectory, entry.Name()))
		if err != nil {
			return nil, err
		}
		file, err := parser.ParseFile(fileSet, relative, data, 0)
		if err != nil {
			return nil, err
		}
		files[relative] = file
		parsed = append(parsed, file)
	}
	if len(parsed) == 0 {
		return nil, fmt.Errorf("typed Go package has no source files")
	}
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}, Selections: map[*ast.SelectorExpr]*types.Selection{}}
	config := types.Config{Importer: importer.Default(), Error: func(error) {}}
	packagePath := modulePackagePath(root, directory)
	checked, _ := config.Check(packagePath, fileSet, parsed, info)
	if checked == nil {
		return nil, fmt.Errorf("Go package type checking produced no package")
	}
	return &typedGoPackage{FileSet: fileSet, Files: files, Info: info, Package: checked}, nil
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
					return nil, nil, fmt.Errorf("binding declaration is ambiguous")
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
							return nil, nil, fmt.Errorf("binding declaration is ambiguous")
						}
						foundIdent, foundNode = name, valueSpec
					}
				}
			}
		}
	}
	if foundIdent == nil {
		return nil, nil, fmt.Errorf("binding declaration was not resolved")
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
		return nil, nil, fmt.Errorf("binding registration use was not resolved")
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

func validRegistrationContext(node ast.Node, symbol, metricID string) bool {
	switch value := node.(type) {
	case *ast.CallExpr:
		return functionName(value.Fun) == symbol
	case *ast.KeyValueExpr:
		key, ok := value.Key.(*ast.Ident)
		return ok && (key.Name == "MetricBindings" || key.Name == "MetricID")
	default:
		return false
	}
}

func normalizedAST(fileSet *token.FileSet, node ast.Node) ([]byte, error) {
	var normalized bytes.Buffer
	if err := format.Node(&normalized, fileSet, node); err != nil {
		return nil, err
	}
	return normalized.Bytes(), nil
}

func bindingObjectIdentity(fileSet *token.FileSet, object types.Object) string {
	position := fileSet.PositionFor(object.Pos(), false)
	packagePath := ""
	if object.Pkg() != nil {
		packagePath = object.Pkg().Path()
	}
	return fmt.Sprintf("%s|%s|%s|%d:%d|%s", packagePath, object.Name(), position.Filename, position.Line, position.Column, object.Type().String())
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
