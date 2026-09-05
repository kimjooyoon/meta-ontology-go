package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"sort"
	"strings"
)

const functionLineLimit = 75

type suffixBinding struct {
	object types.Object
	name   string
	type_  ast.Expr
	pos    token.Pos
}

type suffixCandidate struct {
	start      int
	end        int
	helperName string
	arguments  []suffixBinding
	helper     []byte
	result     []byte
}

type returnTailCandidate struct {
	helperName string
	arguments  []suffixBinding
	helper     []byte
	result     []byte
	evidence   StrategyEvidence
}

type renderedCapacityObservationStatus string

const (
	renderedCapacityWithinCap renderedCapacityObservationStatus = "WITHIN_CAP"
	renderedCapacityOverCap    renderedCapacityObservationStatus = "OVER_CAP"
	renderedCapacityUnmeasured renderedCapacityObservationStatus = "UNMEASURED"
)

type renderedCapacityObservation struct {
	declaration       *ast.FuncDecl
	subject           string
	receiver          string
	functionStart     string
	functionEnd       string
	declarationStart  string
	declarationEnd    string
	sourceDigest      string
	functionLines     int
	functionStatus    renderedCapacityObservationStatus
	helperLines       *int
	helperStatus      renderedCapacityObservationStatus
	helperFailure     error
}

type renderedCapacitySelection struct {
	function      *ast.FuncDecl
	observations []renderedCapacityObservation
}

type typeEvidence struct {
	info  *types.Info
	pkg   *types.Package
	files []*ast.File
	funcs map[*types.Func]*ast.FuncDecl
}

func prepareOversizedFunctions(root, logical string, source []byte, fset *token.FileSet, file *ast.File) ([]byte, []StrategyEvidence, error) {
	current := append([]byte(nil), source...)
	currentSet, currentFile := fset, file
	evidence := make([]StrategyEvidence, 0)
	for {
		selection, err := firstOversizedFunction(currentSet, currentFile, current)
		if err != nil {
			return nil, nil, err
		}
		if selection.function == nil {
			return current, evidence, nil
		}
		function := selection.function
		prepared, strategyEvidence, err := decomposeFunction(root, logical, current, currentSet, currentFile, function, selection.observations)
		if err != nil {
			return nil, nil, withRenderedCapacityDiagnostics(err, selection.observations)
		}
		if bytes.Equal(prepared, current) {
			err := failWithDiagnostics("derive-recipe", "select-safe-suffix", "NO_SAFE_DECLARATION_CAPACITY", "KNOWN_CONTRADICTION", "report-counterexample", []string{
				"declaration=" + functionIdentity(currentSet, function),
				fmt.Sprintf("function_lines=%d", declarationLines(currentSet, function)),
			})
			return nil, nil, withRenderedCapacityDiagnostics(err, selection.observations)
		}
		if strategyEvidence != nil {
			evidence = append(evidence, *strategyEvidence)
		}
		current = prepared
		currentSet = token.NewFileSet()
		parsed, err := parser.ParseFile(currentSet, logical, current, parser.ParseComments)
		if err != nil {
			return nil, nil, fail("rewrite-source", "parse-decomposed-source", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
		}
		currentFile = parsed
	}
}

func firstOversizedFunction(fset *token.FileSet, file *ast.File, source []byte) (*renderedCapacitySelection, error) {
	return firstOversizedFunctionWithRenderer(fset, file, source, renderedDeclarationHelper)
}

func firstOversizedFunctionWithRenderer(fset *token.FileSet, file *ast.File, source []byte, renderer func(*token.FileSet, *ast.File, []byte, *ast.FuncDecl) ([]byte, error)) (*renderedCapacitySelection, error) {
	selection := &renderedCapacitySelection{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name == nil || function.Name.Name == "init" || function.Body == nil {
			continue
		}
		observation := observeRenderedCapacityWithRenderer(fset, file, source, function, renderer)
		selection.observations = append(selection.observations, observation)
		if observation.functionStatus == renderedCapacityOverCap || observation.helperStatus == renderedCapacityOverCap {
			if selection.function == nil || function.Pos() < selection.function.Pos() {
				selection.function = function
			}
		}
	}
	if selection.function != nil {
		return selection, nil
	}
	for _, observation := range selection.observations {
		if observation.helperStatus == renderedCapacityUnmeasured {
			return nil, renderedCapacityObservationFailure(observation)
		}
	}
	return selection, nil
}

func observeRenderedCapacity(fset *token.FileSet, file *ast.File, source []byte, function *ast.FuncDecl) renderedCapacityObservation {
	return observeRenderedCapacityWithRenderer(fset, file, source, function, renderedDeclarationHelper)
}

func observeRenderedCapacityWithRenderer(fset *token.FileSet, file *ast.File, source []byte, function *ast.FuncDecl, renderer func(*token.FileSet, *ast.File, []byte, *ast.FuncDecl) ([]byte, error)) renderedCapacityObservation {
	declarationStart := function.Pos()
	if function.Doc != nil {
		declarationStart = function.Doc.Pos()
	}
	functionLines := declarationLines(fset, function)
	observation := renderedCapacityObservation{
		declaration:      function,
		subject:          functionIdentity(fset, function),
		functionStart:    fset.Position(function.Pos()).String(),
		functionEnd:      fset.Position(function.End()).String(),
		declarationStart: fset.Position(declarationStart).String(),
		declarationEnd:   fset.Position(function.End()).String(),
		sourceDigest:     proofDigest(source),
		functionLines:    functionLines,
		functionStatus:   renderedCapacityWithinCap,
		helperStatus:     renderedCapacityUnmeasured,
	}
	if function.Recv != nil {
		if receiver, ok := receiverBaseIdentifier(function.Recv); ok {
			observation.receiver = receiver
		}
	}
	if functionLines > functionLineLimit {
		observation.functionStatus = renderedCapacityOverCap
	}
	rendered, err := renderer(fset, file, source, function)
	if err != nil {
		observation.helperFailure = err
		return observation
	}
	helperLines := physicalLines(rendered)
	observation.helperLines = &helperLines
	observation.helperStatus = renderedCapacityWithinCap
	if helperLines > functionLineLimit {
		observation.helperStatus = renderedCapacityOverCap
	}
	return observation
}

func renderedCapacityObservationFailure(observation renderedCapacityObservation) error {
	if observation.helperFailure == nil {
		return nil
	}
	var failure Failure
	if errors.As(observation.helperFailure, &failure) {
		failure.Diagnostics = append(failure.Diagnostics, renderedCapacityObservationDiagnostics(observation)...)
		return failure
	}
	return observation.helperFailure
}

func withRenderedCapacityDiagnostics(err error, observations []renderedCapacityObservation) error {
	if err == nil {
		return nil
	}
	var failure Failure
	if !errors.As(err, &failure) {
		return err
	}
	for _, observation := range observations {
		if observation.helperStatus == renderedCapacityUnmeasured {
			failure.Diagnostics = append(failure.Diagnostics, renderedCapacityObservationDiagnostics(observation)...)
		}
	}
	return failure
}

func renderedCapacityObservationDiagnostics(observation renderedCapacityObservation) []string {
	diagnostics := []string{
		"preflight=rendered-capacity",
		"measurement=UNMEASURED",
		"subject=" + observation.subject,
		"function_start=" + observation.functionStart,
		"function_end=" + observation.functionEnd,
		"declaration_start=" + observation.declarationStart,
		"declaration_end=" + observation.declarationEnd,
		"source_digest=" + observation.sourceDigest,
		fmt.Sprintf("function_lines=%d", observation.functionLines),
	}
	if observation.helperFailure != nil {
		diagnostics = append(diagnostics, "helper_failure="+observation.helperFailure.Error())
	}
	return diagnostics
}

func declarationLines(fset *token.FileSet, declaration ast.Node) int {
	start := fset.Position(declaration.Pos()).Line
	end := fset.Position(declaration.End()).Line
	return end - start + 1
}

func renderedDeclarationHelper(fset *token.FileSet, file *ast.File, source []byte, function *ast.FuncDecl) ([]byte, error) {
	list, err := imports(file)
	if err != nil {
		return nil, err
	}
	start := function.Pos()
	if function.Doc != nil {
		start = function.Doc.Pos()
	}
	selected := []declaration{{
		node:     function,
		start:    fset.Position(start).Offset,
		end:      fset.Position(function.End()).Offset,
		identity: functionIdentity(fset, function),
	}}
	_, helperImports, renderErr := renderImports(fset, file, []ast.Decl{function}, list, false)
	if renderErr != nil {
		return nil, renderErr
	}
	helper, err := renderSelectedHelper(fset, file, source, selected, helperImports)
	if err != nil {
		return nil, fail("generate-helpers", "format-helper", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
	}
	return helper, nil
}

func functionIdentity(fset *token.FileSet, function *ast.FuncDecl) string {
	if function.Recv == nil {
		return "func:" + function.Name.Name
	}
	return "method:" + fset.Position(function.Pos()).String() + ":" + function.Name.Name
}

func decomposeFunction(root, logical string, source []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl, preflight []renderedCapacityObservation) ([]byte, *StrategyEvidence, error) {
	if function.Recv != nil {
		return nil, nil, failWithDiagnostics("derive-recipe", "select-safe-suffix", "METHOD_SUFFIX_DECOMPOSITION_UNSAFE", "KNOWN_CONTRADICTION", "report-contradiction", []string{
			"declaration=" + functionIdentity(fset, function),
		})
	}
	if function.Type != nil && function.Type.TypeParams != nil && len(function.Type.TypeParams.List) != 0 {
		return nil, nil, failWithDiagnostics("derive-recipe", "select-safe-suffix", "UNSUPPORTED_GENERIC", "KNOWN_CONTRADICTION", "report-contradiction", []string{
			"declaration=" + functionIdentity(fset, function),
		})
	}
	evidence, err := checkTypes(root, logical, fset, file, function)
	if err != nil {
		return nil, nil, err
	}
	if candidate, candidateErr := buildReturnTailCandidate(root, logical, source, fset, file, function, evidence, functionNames(file), preflight); candidateErr != nil {
		if !isKnownSuffixContradiction(candidateErr) {
			return nil, nil, candidateErr
		}
		if returnTailShapeEligible(function, evidence.info) {
			return nil, nil, failWithDiagnostics("derive-recipe", "select-safe-suffix", "NO_SAFE_DECLARATION_CAPACITY", "KNOWN_CONTRADICTION", "report-counterexample", []string{
				"declaration=" + functionIdentity(fset, function),
				"strategy=" + returnTailStrategy,
				fmt.Sprintf("function_lines=%d", declarationLines(fset, function)),
			})
		}
	} else if candidate != nil {
		return candidate.result, &candidate.evidence, nil
	}
	existing := functionNames(file)
	for index := range slices.Backward(function.Body.List) {
		candidate, candidateErr := buildSuffixCandidate(source, fset, file, function, index, evidence, existing)
		if candidateErr != nil {
			if isKnownSuffixContradiction(candidateErr) {
				continue
			}
			return nil, nil, candidateErr
		}
		if candidate != nil {
			return candidate.result, nil, nil
		}
	}
	return nil, nil, failWithDiagnostics("derive-recipe", "select-safe-suffix", "NO_SAFE_DECLARATION_CAPACITY", "KNOWN_CONTRADICTION", "report-counterexample", []string{
		"declaration=" + functionIdentity(fset, function),
		fmt.Sprintf("function_lines=%d", declarationLines(fset, function)),
	})
}

func functionNames(file *ast.File) map[string]bool {
	result := make(map[string]bool)
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name != nil {
			result[function.Name.Name] = true
		}
	}
	return result
}

func buildSuffixCandidate(source []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl, startIndex int, evidence typeEvidence, existing map[string]bool) (*suffixCandidate, error) {
	statements := function.Body.List[startIndex:]
	if len(statements) == 0 || hasUnsafeOuterScope(function.Body.List[:startIndex]) || hasUnsafeSuffix(statements, evidence.info) {
		return nil, knownSuffixContradiction("suffix control-flow or scope invariant is not preserved")
	}
	bindings, err := suffixBindings(statements, function, fset, evidence)
	if err != nil {
		return nil, err
	}
	if hasMutatedBinding(statements, bindings, evidence.info) {
		return nil, knownSuffixContradiction("outer free binding is mutated")
	}
	name := stableSuffixName(function.Name.Name, startIndex+1)
	if existing[name] {
		return nil, failWithDiagnostics("derive-recipe", "select-safe-suffix", "DECLARATION_IDENTITY_COLLISION", "KNOWN_CONTRADICTION", "report-contradiction", []string{"helper=" + name})
	}
	first := statements[0]
	last := statements[len(statements)-1]
	start := fset.Position(first.Pos()).Offset
	end := fset.Position(last.End()).Offset
	start = includeLeadingComments(fset, file, first, start)
	if start < 0 || end < start || end > len(source) {
		return nil, knownSuffixContradiction("suffix source coordinates are invalid")
	}
	helper, err := renderSuffixHelper(fset, name, bindings, source[start:end])
	if err != nil {
		return nil, err
	}
	if physicalLines(helper) > functionLineLimit {
		return nil, knownSuffixContradiction("suffix helper exceeds the physical line limit")
	}
	call := []byte(name + "(" + bindingArguments(bindings) + ")")
	modified, err := replaceSource(source, start, end, call)
	if err != nil {
		return nil, err
	}
	formatted, err := format.Source(modified)
	if err != nil {
		return nil, fail("rewrite-source", "format-decomposed-source", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
	}
	if lines, ok := namedFunctionLines(formatted, function.Name.Name); !ok || lines > functionLineLimit {
		return nil, knownSuffixContradiction("outer function remains over the physical line limit")
	}
	combined := append(bytes.TrimRight(formatted, "\n"), '\n', '\n')
	combined = append(combined, helper...)
	combined, err = format.Source(combined)
	if err != nil {
		return nil, fail("rewrite-source", "format-decomposed-source", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
	}
	return &suffixCandidate{start: start, end: end, helperName: name, arguments: bindings, helper: helper, result: combined}, nil
}

func stableSuffixName(function string, suffix int) string {
	return fmt.Sprintf("%sExtractedSuffix%02d", function, suffix)
}

func includeLeadingComments(fset *token.FileSet, file *ast.File, first ast.Stmt, start int) int {
	firstLine := fset.Position(first.Pos()).Line
	for _, group := range file.Comments {
		end := fset.Position(group.End())
		if end.Offset > start || firstLine-end.Line > 1 {
			continue
		}
		begin := fset.Position(group.Pos()).Offset
		if begin < start {
			start = begin
		}
	}
	return start
}

func suffixBindings(statements []ast.Stmt, function *ast.FuncDecl, fset *token.FileSet, evidence typeEvidence) ([]suffixBinding, error) {
	inside := suffixDefinedObjects(statements, evidence.info)
	objects := suffixFreeObjects(statements, function, inside, evidence.info)
	return renderSuffixBindings(objects, fset, evidence.info, evidence.pkg)
}

func suffixDefinedObjects(statements []ast.Stmt, info *types.Info) map[types.Object]bool {
	inside := map[types.Object]bool{}
	for _, statement := range statements {
		ast.Inspect(statement, func(node ast.Node) bool {
			identifier, ok := node.(*ast.Ident)
			if ok {
				if object := info.Defs[identifier]; object != nil {
					inside[object] = true
				}
			}
			return true
		})
	}
	return inside
}

func suffixFreeObjects(statements []ast.Stmt, function *ast.FuncDecl, inside map[types.Object]bool, info *types.Info) map[types.Object]bool {
	objects := map[types.Object]bool{}
	for _, statement := range statements {
		ast.Inspect(statement, func(node ast.Node) bool {
			identifier, ok := node.(*ast.Ident)
			if !ok {
				return true
			}
			object := info.Uses[identifier]
			if object == nil || inside[object] || object.Pos() == token.NoPos ||
				object.Pos() >= statements[0].Pos() ||
				object.Pos() < function.Pos() || object.Pos() >= function.End() {
				return true
			}
			switch object.(type) {
			case *types.Var, *types.Const:
				objects[object] = true
			}
			return true
		})
	}
	return objects
}

func renderSuffixBindings(objects map[types.Object]bool, fset *token.FileSet, info *types.Info, current *types.Package) ([]suffixBinding, error) {
	result := make([]suffixBinding, 0, len(objects))
	for object := range objects {
		text := types.TypeString(object.Type(), func(imported *types.Package) string {
			if imported == current {
				return ""
			}
			return packageAlias(info, imported)
		})
		typeExpr, err := parser.ParseExpr(text)
		if err != nil {
			return nil, failWithDiagnostics("derive-recipe", "type-check-suffix", "TYPE_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-type-evidence", []string{
				"evidence=type-string",
				"binding=" + object.Name(),
				"type-string=" + text,
				"type-string-parse-error=" + err.Error(),
			})
		}
		result = append(result, suffixBinding{object: object, name: object.Name(), type_: typeExpr, pos: object.Pos()})
	}
	sort.Slice(result, func(left, right int) bool {
		leftOffset := fset.Position(result[left].pos).Offset
		rightOffset := fset.Position(result[right].pos).Offset
		if leftOffset != rightOffset {
			return leftOffset < rightOffset
		}
		return result[left].name < result[right].name
	})
	return result, nil
}

func packageAlias(info *types.Info, imported *types.Package) string {
	if imported == nil {
		return ""
	}
	aliases := make([]string, 0)
	for identifier, used := range info.Uses {
		packageName, ok := used.(*types.PkgName)
		if ok && packageName.Imported() == imported {
			aliases = append(aliases, identifier.Name)
		}
	}
	sort.Strings(aliases)
	if len(aliases) > 0 {
		return aliases[0]
	}
	return imported.Name()
}

func bindingArguments(bindings []suffixBinding) string {
	arguments := make([]string, len(bindings))
	for index, binding := range bindings {
		arguments[index] = binding.name
	}
	return strings.Join(arguments, ", ")
}

func renderSuffixHelper(fset *token.FileSet, name string, bindings []suffixBinding, body []byte) ([]byte, error) {
	var output bytes.Buffer
	output.WriteString("func ")
	output.WriteString(name)
	output.WriteByte('(')
	for index, binding := range bindings {
		if index != 0 {
			output.WriteString(", ")
		}
		output.WriteString(binding.name)
		output.WriteByte(' ')
		if err := format.Node(&output, fset, binding.type_); err != nil {
			return nil, fail("derive-recipe", "render-suffix-helper", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
		}
	}
	output.WriteString(") {\n")
	output.Write(body)
	if len(body) == 0 || body[len(body)-1] != '\n' {
		output.WriteByte('\n')
	}
	output.WriteString("}\n")
	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return nil, fail("derive-recipe", "render-suffix-helper", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
	}
	return formatted, nil
}

func namedFunctionLines(source []byte, name string) (int, bool) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "decomposed.go", source, parser.ParseComments)
	if err != nil {
		return 0, false
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Recv == nil && function.Name != nil && function.Name.Name == name {
			return declarationLines(fset, function), true
		}
	}
	return 0, false
}

func replaceSource(source []byte, start, end int, replacement []byte) ([]byte, error) {
	if start < 0 || end < start || end > len(source) {
		return nil, knownSuffixContradiction("source edit coordinates are invalid")
	}
	result := make([]byte, 0, len(source)-end+start+len(replacement))
	result = append(result, source[:start]...)
	result = append(result, replacement...)
	result = append(result, source[end:]...)
	return result, nil
}

func hasUnsafeOuterScope(statements []ast.Stmt) bool {
	for _, statement := range statements {
		unsafe := false
		ast.Inspect(statement, func(node ast.Node) bool {
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}
			switch node.(type) {
			case *ast.DeferStmt, *ast.GoStmt, *ast.LabeledStmt, *ast.ReturnStmt:
				unsafe = true
			}
			if branch, ok := node.(*ast.BranchStmt); ok && branch.Tok != token.FALLTHROUGH {
				unsafe = true
			}
			return !unsafe
		})
		if unsafe {
			return true
		}
	}
	return false
}

func hasUnsafeSuffix(statements []ast.Stmt, info *types.Info) bool {
	for _, statement := range statements {
		hazards := &suffixHazards{}
		ast.Walk(hazardVisitor{hazards: hazards}, statement)
		if hazards.unsafe {
			return true
		}
	}
	return false
}

type suffixHazards struct{ unsafe bool }

type hazardVisitor struct {
	hazards   *suffixHazards
	loopDepth int
}

func (visitor hazardVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil || visitor.hazards.unsafe {
		return nil
	}
	if _, ok := node.(*ast.FuncLit); ok {
		return nil
	}
	switch value := node.(type) {
	case *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt:
		return hazardVisitor{hazards: visitor.hazards, loopDepth: visitor.loopDepth + 1}
	case *ast.DeferStmt, *ast.GoStmt, *ast.LabeledStmt, *ast.ReturnStmt:
		visitor.hazards.unsafe = true
	case *ast.BranchStmt:
		if visitor.loopDepth == 0 && value.Tok != token.FALLTHROUGH {
			visitor.hazards.unsafe = true
		}
	case *ast.CallExpr:
		if identifier, ok := value.Fun.(*ast.Ident); ok &&
			(identifier.Name == "panic" || identifier.Name == "recover") {
			visitor.hazards.unsafe = true
		}
	}
	return hazardVisitor{hazards: visitor.hazards, loopDepth: visitor.loopDepth}
}

func hasMutatedBinding(statements []ast.Stmt, bindings []suffixBinding, info *types.Info) bool {
	free := make(map[types.Object]bool, len(bindings))
	for _, binding := range bindings {
		free[binding.object] = true
	}
	mutated := false
	for _, statement := range statements {
		ast.Inspect(statement, func(node ast.Node) bool {
			switch value := node.(type) {
			case *ast.AssignStmt:
				for _, expression := range value.Lhs {
					if object := assignedObject(expression, info); free[object] {
						mutated = true
					}
				}
			case *ast.IncDecStmt:
				if object := assignedObject(value.X, info); free[object] {
					mutated = true
				}
			case *ast.UnaryExpr:
				if value.Op == token.AND && free[assignedObject(value.X, info)] {
					mutated = true
				}
			}
			return !mutated
		})
		if mutated {
			return true
		}
	}
	return false
}

func assignedObject(expression ast.Expr, info *types.Info) types.Object {
	if identifier, ok := expression.(*ast.Ident); ok {
		if object := info.Defs[identifier]; object != nil {
			return object
		}
		return info.Uses[identifier]
	}
	if selector, ok := expression.(*ast.SelectorExpr); ok {
		return assignedObject(selector.X, info)
	}
	if index, ok := expression.(*ast.IndexExpr); ok {
		return assignedObject(index.X, info)
	}
	if index, ok := expression.(*ast.IndexListExpr); ok {
		return assignedObject(index.X, info)
	}
	return usedObject(expression, info)
}

func usedObject(expression ast.Expr, info *types.Info) types.Object {
	identifier, ok := expression.(*ast.Ident)
	if !ok {
		return nil
	}
	return info.Uses[identifier]
}

type suffixContradiction struct{ message string }

func (e suffixContradiction) Error() string { return e.message }

func knownSuffixContradiction(message string) error { return suffixContradiction{message: message} }

func isKnownSuffixContradiction(err error) bool {
	_, ok := err.(suffixContradiction)
	return ok
}
