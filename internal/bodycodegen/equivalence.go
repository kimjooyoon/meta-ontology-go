package bodycodegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

const (
	routeEquivalenceSchema = "gooo/body-codegen-route-equivalence/v1"
	routeEquivalenceMethod = "canonical_control_flow_form/v1"
	routeEquivalenceScope  = "typechecked, terminating, side-effect-free body-codegen profile; not a proof of unstated user intent"
)

// RouteEquivalenceReceipt records a compiler-derived comparison between the
// accepted source body and the emitted body. It can prove only the explicitly
// declared source-shape rewrites inside the closed body-codegen profile.
type RouteEquivalenceReceipt struct {
	Schema                  string `json:"schema"`
	Decision                string `json:"decision"`
	Method                  string `json:"method"`
	Rule                    string `json:"rule"`
	SourceSemanticDigest    string `json:"source_semantic_digest"`
	GeneratedSemanticDigest string `json:"generated_semantic_digest"`
	Equivalent              bool   `json:"equivalent"`
	Scope                   string `json:"scope"`
}

type canonicalSemanticForm struct {
	Shape       string `json:"shape"`
	Condition   string `json:"condition,omitempty"`
	TrueResult  string `json:"true_result,omitempty"`
	FalseResult string `json:"false_result,omitempty"`
	Body        string `json:"body,omitempty"`
}

func routeEquivalence(packageName, activityName, inputType, outputType, body string, generated []byte, rule string) (RouteEquivalenceReceipt, error) {
	sourceFileSet := token.NewFileSet()
	sourceText := fmt.Sprintf("package %s\nfunc %s(input %s) %s {\n%s\n}\n", packageName, activityName, inputType, outputType, body)
	sourceFile, err := parser.ParseFile(sourceFileSet, "source-body.goo", sourceText, parser.AllErrors)
	if err != nil {
		return RouteEquivalenceReceipt{}, fmt.Errorf("parse accepted source body for equivalence receipt: %w", err)
	}
	sourceFunction, ok := findFunction(sourceFile, activityName)
	if !ok {
		return RouteEquivalenceReceipt{}, fmt.Errorf("accepted source body has no function for equivalence receipt")
	}

	generatedFileSet := token.NewFileSet()
	generatedFile, err := parser.ParseFile(generatedFileSet, "generated-body.go", generated, parser.AllErrors)
	if err != nil {
		return RouteEquivalenceReceipt{}, fmt.Errorf("parse generated body for equivalence receipt: %w", err)
	}
	generatedFunction, ok := findFunction(generatedFile, activityName)
	if !ok {
		return RouteEquivalenceReceipt{}, fmt.Errorf("generated source has no function for equivalence receipt")
	}

	sourceForm, err := canonicalizeSemanticBody(sourceFileSet, sourceFunction.Body)
	if err != nil {
		return RouteEquivalenceReceipt{}, fmt.Errorf("canonicalize accepted source body: %w", err)
	}
	generatedForm, err := canonicalizeSemanticBody(generatedFileSet, generatedFunction.Body)
	if err != nil {
		return RouteEquivalenceReceipt{}, fmt.Errorf("canonicalize generated body: %w", err)
	}
	equivalent := bytes.Equal(sourceForm, generatedForm)
	decision := "FAIL_CLOSED"
	if equivalent {
		decision = "PASS"
	}
	return RouteEquivalenceReceipt{
		Schema: routeEquivalenceSchema, Decision: decision, Method: routeEquivalenceMethod,
		Rule: rule, SourceSemanticDigest: digest(sourceForm),
		GeneratedSemanticDigest: digest(generatedForm), Equivalent: equivalent,
		Scope: routeEquivalenceScope,
	}, nil
}

func canonicalizeSemanticBody(fileSet *token.FileSet, body *ast.BlockStmt) ([]byte, error) {
	if form, ok := conditionalReturnForm(fileSet, body); ok {
		return json.Marshal(form)
	}
	var formatted bytes.Buffer
	if err := format.Node(&formatted, fileSet, body); err != nil {
		return nil, fmt.Errorf("format body AST: %w", err)
	}
	return json.Marshal(canonicalSemanticForm{Shape: "go_ast_body/v1", Body: formatted.String()})
}

func conditionalReturnForm(fileSet *token.FileSet, body *ast.BlockStmt) (canonicalSemanticForm, bool) {
	if isGuardReturnShape(body) {
		conditional := body.List[0].(*ast.IfStmt)
		otherwise := conditional.Else.(*ast.BlockStmt)
		thenReturn := conditional.Body.List[0].(*ast.ReturnStmt)
		elseReturn := otherwise.List[0].(*ast.ReturnStmt)
		if len(thenReturn.Results) != 1 || len(elseReturn.Results) != 1 {
			return canonicalSemanticForm{}, false
		}
		condition, conditionOK := formatNode(fileSet, conditional.Cond)
		trueResult, trueOK := formatNode(fileSet, thenReturn.Results[0])
		falseResult, falseOK := formatNode(fileSet, elseReturn.Results[0])
		if !conditionOK || !trueOK || !falseOK {
			return canonicalSemanticForm{}, false
		}
		return canonicalSemanticForm{
			Shape: "if_return_else_return/v1", Condition: condition,
			TrueResult: trueResult, FalseResult: falseResult,
		}, true
	}
	if form, ok := guardReturnForm(fileSet, body); ok {
		return form, true
	}
	return mergeResultForm(fileSet, body)
}

func guardReturnForm(fileSet *token.FileSet, body *ast.BlockStmt) (canonicalSemanticForm, bool) {
	if body == nil || len(body.List) != 2 {
		return canonicalSemanticForm{}, false
	}
	conditional, ok := body.List[0].(*ast.IfStmt)
	if !ok || conditional.Init != nil || conditional.Else != nil || len(conditional.Body.List) != 1 {
		return canonicalSemanticForm{}, false
	}
	thenReturn, ok := conditional.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(thenReturn.Results) != 1 {
		return canonicalSemanticForm{}, false
	}
	elseReturn, ok := body.List[1].(*ast.ReturnStmt)
	if !ok || len(elseReturn.Results) != 1 {
		return canonicalSemanticForm{}, false
	}
	condition, conditionOK := formatNode(fileSet, conditional.Cond)
	trueResult, trueOK := formatNode(fileSet, thenReturn.Results[0])
	falseResult, falseOK := formatNode(fileSet, elseReturn.Results[0])
	if !conditionOK || !trueOK || !falseOK {
		return canonicalSemanticForm{}, false
	}
	return canonicalSemanticForm{
		Shape: "if_return_else_return/v1", Condition: condition,
		TrueResult: trueResult, FalseResult: falseResult,
	}, true
}

func mergeResultForm(fileSet *token.FileSet, body *ast.BlockStmt) (canonicalSemanticForm, bool) {
	if body == nil || len(body.List) != 3 {
		return canonicalSemanticForm{}, false
	}
	declaration, ok := body.List[0].(*ast.DeclStmt)
	if !ok {
		return canonicalSemanticForm{}, false
	}
	general, ok := declaration.Decl.(*ast.GenDecl)
	if !ok || general.Tok != token.VAR || len(general.Specs) != 1 {
		return canonicalSemanticForm{}, false
	}
	valueSpec, ok := general.Specs[0].(*ast.ValueSpec)
	if !ok || len(valueSpec.Names) != 1 || valueSpec.Names[0].Name != "_goooResult" || len(valueSpec.Values) != 0 {
		return canonicalSemanticForm{}, false
	}
	typeName, ok := valueSpec.Type.(*ast.Ident)
	if !ok || (typeName.Name != "int64" && typeName.Name != "bool" && typeName.Name != "string") {
		return canonicalSemanticForm{}, false
	}
	conditional, ok := body.List[1].(*ast.IfStmt)
	if !ok || conditional.Init != nil {
		return canonicalSemanticForm{}, false
	}
	otherwise, ok := conditional.Else.(*ast.BlockStmt)
	if !ok || len(conditional.Body.List) != 1 || len(otherwise.List) != 1 {
		return canonicalSemanticForm{}, false
	}
	trueResult, ok := mergeAssignmentValue(conditional.Body.List[0])
	if !ok {
		return canonicalSemanticForm{}, false
	}
	falseResult, ok := mergeAssignmentValue(otherwise.List[0])
	if !ok {
		return canonicalSemanticForm{}, false
	}
	result, ok := body.List[2].(*ast.ReturnStmt)
	if !ok || len(result.Results) != 1 {
		return canonicalSemanticForm{}, false
	}
	resultName, ok := result.Results[0].(*ast.Ident)
	if !ok || resultName.Name != "_goooResult" {
		return canonicalSemanticForm{}, false
	}
	condition, conditionOK := formatNode(fileSet, conditional.Cond)
	trueText, trueOK := formatNode(fileSet, trueResult)
	falseText, falseOK := formatNode(fileSet, falseResult)
	if !conditionOK || !trueOK || !falseOK {
		return canonicalSemanticForm{}, false
	}
	return canonicalSemanticForm{
		Shape: "if_return_else_return/v1", Condition: condition,
		TrueResult: trueText, FalseResult: falseText,
	}, true
}

func mergeAssignmentValue(statement ast.Stmt) (ast.Expr, bool) {
	assignment, ok := statement.(*ast.AssignStmt)
	if !ok || assignment.Tok != token.ASSIGN || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
		return nil, false
	}
	target, ok := assignment.Lhs[0].(*ast.Ident)
	if !ok || target.Name != "_goooResult" {
		return nil, false
	}
	return assignment.Rhs[0], true
}

func formatNode(fileSet *token.FileSet, node ast.Node) (string, bool) {
	var formatted bytes.Buffer
	if err := format.Node(&formatted, fileSet, node); err != nil {
		return "", false
	}
	return formatted.String(), true
}
