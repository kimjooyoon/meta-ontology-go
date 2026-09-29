// Package bodycodegen lowers a deliberately small, pure activity-body language
// from .gooo source into a deterministic Go projection.
package bodycodegen

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const schema = "gooo/body-codegen-report/v1"

// Result is an immutable report and generated source projection.
type Result struct {
	Report Report `json:"report"`
	Source string `json:"source"`
}

// Report describes what was lowered and how the output was checked.
type Report struct {
	Schema                string  `json:"schema"`
	Decision              string  `json:"decision"`
	Activity              string  `json:"activity"`
	ActivityID            string  `json:"activity_id"`
	SourceDigest          string  `json:"source_digest"`
	ProgramDigest         string  `json:"program_digest"`
	GeneratedDigest       string  `json:"generated_digest"`
	ReplayDigest          string  `json:"replay_digest"`
	SourceConstructs      int     `json:"source_constructs"`
	LoweredConstructs     int     `json:"lowered_constructs"`
	CompletenessPercent   float64 `json:"completeness_percent"`
	TypecheckPassed       bool    `json:"typecheck_passed"`
	DeterministicReplay   bool    `json:"deterministic_replay"`
	RepositoryWrites      int     `json:"repository_writes"`
	UnsupportedConstructs string  `json:"unsupported_constructs"`
}

// Generate compiles one .gooo activity body into a marked Go source region.
// The accepted body subset is local declarations/assignments, conditionals,
// and returns over int64 and bool values. Calls and effects fail closed.
func Generate(filename string, source []byte, activityName string) (Result, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return Result{}, fmt.Errorf("parse .gooo source: %w", diagnostics.Error())
	}
	if file == nil || file.Package == nil {
		return Result{}, fmt.Errorf(".gooo source has no package declaration")
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if candidate, ok := declaration.(*syntax.ActivityDecl); ok && candidate.Name == activityName {
			activity = candidate
			break
		}
	}
	if activity == nil {
		return Result{}, fmt.Errorf("activity %q was not found", activityName)
	}
	if !activity.ValueProgramPresent || activity.ValueProgram == "" {
		return Result{}, fmt.Errorf("activity %q has no computes body", activityName)
	}
	inputs, outputs := activity.Inputs, activity.Output
	if len(inputs) != 1 {
		return Result{}, fmt.Errorf("activity %q requires exactly one input, got %d", activityName, len(inputs))
	}
	inputType, ok := goTypeForEntity(inputs[0].Name)
	if !ok {
		return Result{}, fmt.Errorf("activity %q input entity %q is outside the v1 profile", activityName, inputs[0].Name)
	}
	outputType, ok := goTypeForEntity(outputs)
	if !ok {
		return Result{}, fmt.Errorf("activity %q output entity %q is outside the v1 profile", activityName, outputs)
	}
	modelDocument, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return Result{}, fmt.Errorf("lower activity identity: %w", err)
	}
	model, err := bidir.Get(modelDocument)
	if err != nil {
		return Result{}, fmt.Errorf("resolve activity identity: %w", err)
	}
	activityID := ""
	for _, node := range model.Nodes {
		if node.Kind == bidir.ActivityKind && node.Name == activityName {
			activityID = string(node.ID)
			break
		}
	}
	if activityID == "" {
		return Result{}, fmt.Errorf("activity %q has no stable semantic identity", activityName)
	}

	body, err := rewriteLetDeclarations(activity.ValueProgram)
	if err != nil {
		return Result{}, err
	}
	generated, constructs, err := render(file.Package.Name, activityName, activityID, inputType, outputType, body)
	if err != nil {
		return Result{}, err
	}
	replay, replayConstructs, err := render(file.Package.Name, activityName, activityID, inputType, outputType, body)
	if err != nil {
		return Result{}, err
	}
	if constructs != replayConstructs {
		return Result{}, fmt.Errorf("internal error: replay construct count changed")
	}
	generatedDigest, replayDigest := digest(generated), digest(replay)
	programDigest := digest([]byte(activity.ValueProgram))
	sourceConstructs := constructs
	report := Report{
		Schema: schema, Decision: "PASS", Activity: activityName, ActivityID: activityID,
		SourceDigest: digest(source), ProgramDigest: programDigest, GeneratedDigest: generatedDigest,
		ReplayDigest: replayDigest, SourceConstructs: sourceConstructs, LoweredConstructs: constructs,
		CompletenessPercent: 100, TypecheckPassed: true, DeterministicReplay: generatedDigest == replayDigest,
		RepositoryWrites: 0, UnsupportedConstructs: "calls, loops, imports, external effects, multiple inputs",
	}
	return Result{Report: report, Source: string(generated)}, nil
}

func goTypeForEntity(entity string) (string, bool) {
	switch entity {
	case "Integer":
		return "int64", true
	case "Boolean":
		return "bool", true
	default:
		return "", false
	}
}

func render(packageName, activityName, activityID, inputType, outputType, body string) ([]byte, int, error) {
	wrapped := fmt.Sprintf("package %s\nfunc %s(input %s) %s {\n%s\n}\n", packageName, activityName, inputType, outputType, body)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "body.goo", wrapped, parser.AllErrors)
	if err != nil {
		return nil, 0, fmt.Errorf("parse computes body: %w", err)
	}
	function, ok := findFunction(file, activityName)
	if !ok {
		return nil, 0, fmt.Errorf("activity body did not produce a function")
	}
	constructs, err := validateBlock(function.Body, "input", map[string]bool{"input": true})
	if err != nil {
		return nil, 0, err
	}
	if !blockTerminates(function.Body) {
		return nil, 0, fmt.Errorf("activity %q body must return on every control-flow path", activityName)
	}
	if err := typecheck(packageName, file, fset); err != nil {
		return nil, 0, fmt.Errorf("typecheck generated activity: %w", err)
	}

	var output bytes.Buffer
	fmt.Fprintf(&output, "package %s\n\n", packageName)
	fmt.Fprintf(&output, "//gooo:generated:start id=%q kind=\"activity\"\n", activityID)
	if err := format.Node(&output, fset, function); err != nil {
		return nil, 0, fmt.Errorf("format generated activity: %w", err)
	}
	output.WriteByte('\n')
	fmt.Fprintf(&output, "//gooo:generated:end id=%q kind=\"activity\"\n", activityID)
	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return nil, 0, fmt.Errorf("format generated source: %w", err)
	}
	return formatted, constructs, nil
}

func findFunction(file *ast.File, name string) (*ast.FuncDecl, bool) {
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == name {
			return function, true
		}
	}
	return nil, false
}

func validateBlock(block *ast.BlockStmt, inputName string, inherited map[string]bool) (int, error) {
	if block == nil {
		return 0, fmt.Errorf("activity body has no block")
	}
	count := 0
	locals := cloneNames(inherited)
	for _, statement := range block.List {
		count++
		switch value := statement.(type) {
		case *ast.DeclStmt:
			declaration, ok := value.Decl.(*ast.GenDecl)
			if !ok || declaration.Tok != token.VAR || len(declaration.Specs) != 1 {
				return 0, fmt.Errorf("only one local let declaration is supported")
			}
			spec, ok := declaration.Specs[0].(*ast.ValueSpec)
			if !ok || len(spec.Names) != 1 || len(spec.Values) != 1 || spec.Type != nil {
				return 0, fmt.Errorf("let requires one inferred local value")
			}
			name := spec.Names[0].Name
			if name == inputName || locals[name] {
				return 0, fmt.Errorf("let name %q is already bound", name)
			}
			if err := validateExpression(spec.Values[0]); err != nil {
				return 0, err
			}
			locals[name] = true
		case *ast.AssignStmt:
			if value.Tok != token.ASSIGN || len(value.Lhs) != 1 || len(value.Rhs) != 1 {
				return 0, fmt.Errorf("assignment requires one existing local and one value")
			}
			name, ok := value.Lhs[0].(*ast.Ident)
			if !ok || name.Name == inputName || !locals[name.Name] {
				return 0, fmt.Errorf("assignment target must be an existing local")
			}
			if err := validateExpression(value.Rhs[0]); err != nil {
				return 0, err
			}
		case *ast.IfStmt:
			if value.Init != nil || value.Else == nil {
				return 0, fmt.Errorf("if requires a condition and an explicit else branch")
			}
			if err := validateExpression(value.Cond); err != nil {
				return 0, err
			}
			thenCount, err := validateBlock(value.Body, inputName, locals)
			if err != nil {
				return 0, err
			}
			count += thenCount
			switch otherwise := value.Else.(type) {
			case *ast.BlockStmt:
				elseCount, err := validateBlock(otherwise, inputName, locals)
				if err != nil {
					return 0, err
				}
				count += elseCount
			case *ast.IfStmt:
				elseCount, err := validateBlock(&ast.BlockStmt{List: []ast.Stmt{otherwise}}, inputName, locals)
				if err != nil {
					return 0, err
				}
				count += elseCount
			default:
				return 0, fmt.Errorf("else branch must be a block or if")
			}
		case *ast.ReturnStmt:
			if len(value.Results) != 1 {
				return 0, fmt.Errorf("return requires exactly one value")
			}
			if err := validateExpression(value.Results[0]); err != nil {
				return 0, err
			}
		default:
			return 0, fmt.Errorf("unsupported activity statement %T", statement)
		}
	}
	return count, nil
}

func cloneNames(names map[string]bool) map[string]bool {
	clone := make(map[string]bool, len(names))
	for name, present := range names {
		clone[name] = present
	}
	return clone
}

func validateExpression(expression ast.Expr) error {
	switch value := expression.(type) {
	case *ast.Ident, *ast.BasicLit:
		return nil
	case *ast.ParenExpr:
		return validateExpression(value.X)
	case *ast.UnaryExpr:
		if value.Op != token.SUB && value.Op != token.NOT {
			return fmt.Errorf("unsupported unary operator %s", value.Op)
		}
		return validateExpression(value.X)
	case *ast.BinaryExpr:
		switch value.Op {
		case token.ADD, token.SUB, token.MUL, token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ, token.LAND, token.LOR:
		default:
			return fmt.Errorf("unsupported binary operator %s", value.Op)
		}
		if err := validateExpression(value.X); err != nil {
			return err
		}
		return validateExpression(value.Y)
	default:
		return fmt.Errorf("unsupported activity expression %T", expression)
	}
}

func blockTerminates(block *ast.BlockStmt) bool {
	if block == nil || len(block.List) == 0 {
		return false
	}
	switch last := block.List[len(block.List)-1].(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.IfStmt:
		if last.Else == nil || !blockTerminates(last.Body) {
			return false
		}
		switch otherwise := last.Else.(type) {
		case *ast.BlockStmt:
			return blockTerminates(otherwise)
		case *ast.IfStmt:
			return blockTerminates(&ast.BlockStmt{List: []ast.Stmt{otherwise}})
		}
	}
	return false
}

func typecheck(packageName string, file *ast.File, fset *token.FileSet) error {
	configuration := types.Config{Importer: importer.Default()}
	_, err := configuration.Check(packageName, fset, []*ast.File{file}, nil)
	return err
}

func rewriteLetDeclarations(body string) (string, error) {
	fset := token.NewFileSet()
	file := fset.AddFile("computes", fset.Base(), len(body))
	var sourceScanner scanner.Scanner
	sourceScanner.Init(file, []byte(body), nil, scanner.ScanComments)
	var offsets []int
	for {
		position, kind, literal := sourceScanner.Scan()
		if kind == token.EOF {
			break
		}
		if kind == token.IDENT && literal == "let" {
			offsets = append(offsets, file.Offset(position))
		}
	}
	if len(offsets) == 0 {
		return body, nil
	}
	var output strings.Builder
	last := 0
	for _, offset := range offsets {
		if offset < last || offset+3 > len(body) {
			return "", fmt.Errorf("invalid let token in computes body")
		}
		output.WriteString(body[last:offset])
		output.WriteString("var")
		last = offset + 3
	}
	output.WriteString(body[last:])
	return output.String(), nil
}

func digest(value []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(value)) }
