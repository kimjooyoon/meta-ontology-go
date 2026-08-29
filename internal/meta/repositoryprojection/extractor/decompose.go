package extractor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
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

type typeEvidence struct {
	info *types.Info
}

func prepareOversizedFunctions(root, logical string, source []byte, fset *token.FileSet, file *ast.File) ([]byte, error) {
	current := append([]byte(nil), source...)
	currentSet, currentFile := fset, file
	for {
		function := firstOversizedFunction(currentSet, currentFile)
		if function == nil {
			return current, nil
		}
		prepared, err := decomposeFunction(root, logical, current, currentSet, currentFile, function)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(prepared, current) {
			return nil, failWithDiagnostics("derive-recipe", "select-safe-suffix", "NO_SAFE_DECLARATION_CAPACITY", "KNOWN_CONTRADICTION", "report-contradiction", []string{
				"declaration=" + functionIdentity(currentSet, function),
				fmt.Sprintf("function_lines=%d", declarationLines(currentSet, function)),
			})
		}
		current = prepared
		currentSet = token.NewFileSet()
		parsed, err := parser.ParseFile(currentSet, logical, current, parser.ParseComments)
		if err != nil {
			return nil, fail("rewrite-source", "parse-decomposed-source", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
		}
		currentFile = parsed
	}
}

func firstOversizedFunction(fset *token.FileSet, file *ast.File) *ast.FuncDecl {
	var result *ast.FuncDecl
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name == nil || function.Name.Name == "init" || function.Body == nil {
			continue
		}
		if declarationLines(fset, function) <= functionLineLimit {
			continue
		}
		if result == nil || function.Pos() < result.Pos() {
			result = function
		}
	}
	return result
}

func declarationLines(fset *token.FileSet, declaration ast.Node) int {
	start := fset.Position(declaration.Pos()).Line
	end := fset.Position(declaration.End()).Line
	return end - start + 1
}

func functionIdentity(fset *token.FileSet, function *ast.FuncDecl) string {
	if function.Recv == nil {
		return "func:" + function.Name.Name
	}
	return "method:" + fset.Position(function.Pos()).String() + ":" + function.Name.Name
}

func decomposeFunction(root, logical string, source []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl) ([]byte, error) {
	if function.Recv != nil {
		return nil, failWithDiagnostics("derive-recipe", "select-safe-suffix", "METHOD_SUFFIX_DECOMPOSITION_UNSAFE", "KNOWN_CONTRADICTION", "report-contradiction", []string{
			"declaration=" + functionIdentity(fset, function),
		})
	}
	evidence, err := checkTypes(fset, file)
	if err != nil {
		return nil, err
	}
	existing := functionNames(file)
	for index := len(function.Body.List) - 1; index >= 0; index-- {
		candidate, candidateErr := buildSuffixCandidate(source, fset, file, function, index, evidence, existing)
		if candidateErr != nil {
			if isKnownSuffixContradiction(candidateErr) {
				continue
			}
			return nil, candidateErr
		}
		if candidate != nil {
			return candidate.result, nil
		}
	}
	return nil, failWithDiagnostics("derive-recipe", "select-safe-suffix", "NO_SAFE_DECLARATION_CAPACITY", "KNOWN_CONTRADICTION", "report-contradiction", []string{
			"declaration=" + functionIdentity(fset, function),
		fmt.Sprintf("function_lines=%d", declarationLines(fset, function)),
	})
}

func checkTypes(fset *token.FileSet, file *ast.File) (typeEvidence, error) {
	info := &types.Info{
		Defs:   map[*ast.Ident]types.Object{},
		Uses:   map[*ast.Ident]types.Object{},
		Scopes: map[ast.Node]*types.Scope{},
	}
	configuration := types.Config{Importer: importer.Default(), Error: func(error) {}}
	_, err := configuration.Check("gooo/oversized-function", fset, []*ast.File{file}, info)
	if err != nil {
		return typeEvidence{}, fail("derive-recipe", "type-check-suffix", "TYPE_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-type-evidence", nil)
	}
	return typeEvidence{info: info}, nil
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
	inside := map[types.Object]bool{}
	for _, statement := range statements {
		ast.Inspect(statement, func(node ast.Node) bool {
			identifier, ok := node.(*ast.Ident)
			if !ok {
				return true
			}
			if object := evidence.info.Defs[identifier]; object != nil {
				inside[object] = true
			}
			return true
		})
	}
	objects := map[types.Object]bool{}
	for _, statement := range statements {
		ast.Inspect(statement, func(node ast.Node) bool {
			identifier, ok := node.(*ast.Ident)
			if !ok {
				return true
			}
			object := evidence.info.Uses[identifier]
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
	result := make([]suffixBinding, 0, len(objects))
	for object := range objects {
		text := types.TypeString(object.Type(), func(imported *types.Package) string {
			if imported == nil {
				return ""
			}
			for identifier, used := range evidence.info.Uses {
				packageName, ok := used.(*types.PkgName)
				if ok && packageName.Imported() == imported {
					return identifier.Name
				}
			}
			return imported.Name()
		})
		typeExpr, err := parser.ParseExpr(text)
		if err != nil {
			return nil, fail("derive-recipe", "type-check-suffix", "TYPE_EVIDENCE_MISSING", "DIRECT_MISSING", "restore-type-evidence", nil)
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
		if ok && function.Name != nil && function.Name.Name == name {
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
