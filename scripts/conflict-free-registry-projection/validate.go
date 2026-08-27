package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"os/exec"
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
			var typeCheckErr bindingTypeCheckError
			if errors.As(err, &typeCheckErr) {
				return &Diagnostic{Decision: "FAIL_CLOSED", Stage: "LOWER_RESOLUTION", Step: "BINDING_PACKAGE_TYPE_CHECK", Reason: "PACKAGE_TYPE_CHECK_FAILED"}
			}
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
		return bindingResolution{}, fmt.Errorf("binding declaration and use are in different packages")
	}
	packageInfo, err := typeCheckGoPackage(root, declaration.Path)
	if err != nil {
		return bindingResolution{}, err
	}
	declarationFile := packageInfo.Files[filepath.ToSlash(declaration.Path)]
	registrationFile := packageInfo.Files[filepath.ToSlash(registration.Path)]
	if declarationFile == nil || registrationFile == nil {
		return bindingResolution{}, fmt.Errorf("binding source file is not in typed package")
	}
	declarationIdent, declarationNode, err := findDeclaration(declarationFile, declaration.Symbol)
	if err != nil {
		return bindingResolution{}, err
	}
	declarationObject := packageInfo.Info.Defs[declarationIdent]
	if declarationObject == nil {
		return bindingResolution{}, fmt.Errorf("binding declaration has no types object")
	}
	registrationIdent, registrationNode, err := findRegistrationUse(registrationFile, registration.Symbol, registration.Line, registration.Column, packageInfo.FileSet, packageInfo.Info)
	if err != nil {
		return bindingResolution{}, err
	}
	registrationObject := packageInfo.Info.Uses[registrationIdent]
	if registrationObject == nil || registrationObject != declarationObject {
		return bindingResolution{}, fmt.Errorf("binding registration use resolves to a different object")
	}
	if !validRegistrationContext(registrationNode, registration.Symbol, metricID, registrationIdent) || (isCallNode(registrationNode) && !hasCatalogOwner(registrationFile, registrationNode)) {
		return bindingResolution{}, fmt.Errorf("binding registration use is not a canonical registration relation")
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
	return bindingResolution{SemanticDigest: digestBytes(payload), MetricOccurrenceAddress: metricAddress, MetricOccurrenceDigest: metricDigest}, nil
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
		return "", nil, "", fmt.Errorf("metric occurrence is not unique in binding relation")
	}
	return candidates[0].address, candidates[0].ast, digestBytes(candidates[0].ast), nil
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
