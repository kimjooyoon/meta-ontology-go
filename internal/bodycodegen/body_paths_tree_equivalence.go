package bodycodegen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
)

// Called only after both native projections have typechecked. The typed arena
// retains every operator, child edge, local binding and ordered statement while
// omitting redundant parentheses and normalizing int64 literal spelling.
func typedBodyTreeEquivalence(ctx context.Context, name, original string, generated []byte) (RouteEquivalenceReceipt, error) {
	left, leftNormalized, err := normalizedTypedBodyTree(ctx, name,
		[]byte("package source\nfunc "+name+"(input int64) int64 {\n"+original+"\n}"))
	if err != nil {
		return RouteEquivalenceReceipt{}, err
	}
	right, rightNormalized, err := normalizedTypedBodyTree(ctx, name, generated)
	if err != nil {
		return RouteEquivalenceReceipt{}, err
	}
	equal, status := bytes.Equal(left, right), "FAIL_CLOSED"
	if equal {
		status = "PASS"
	}
	method := "canonical_typed_body_tree/v1"
	scope := "typechecked Integer -> Integer typed-body arena; exact operator, child-edge, local and statement structure"
	if leftNormalized || rightNormalized {
		method = "normalized_condition_typed_body_tree/v1"
		scope = "typechecked Integer -> Integer arena; exact local and statement structure with >, >=, != and boolean ! reduced to <, <= and equality with false"
	}
	return RouteEquivalenceReceipt{Schema: routeEquivalenceSchema, Decision: status,
		Method: method, Rule: "typed_path_fallback_matches_authoritative_source",
		SourceSemanticDigest: digest(left), GeneratedSemanticDigest: digest(right), Equivalent: equal,
		Scope: scope}, nil
}

func typedBodyTree(ctx context.Context, name string, source []byte) ([]byte, error) {
	tree, _, err := normalizedTypedBodyTree(ctx, name, source)
	return tree, err
}

func normalizedTypedBodyTree(ctx context.Context, name string, source []byte) ([]byte, bool, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "typed-tree.go", source, parser.AllErrors)
	if err != nil {
		return nil, false, err
	}
	function, ok := findFunction(file, name)
	if !ok {
		return nil, false, fmt.Errorf("typed tree function missing")
	}
	b := recipeArena{ctx: ctx}
	root, err := b.sequence(function.Body.List, 0)
	if err != nil {
		return nil, false, err
	}
	if err := b.includeDeclaredInput(); err != nil {
		return nil, false, err
	}
	plan := bodyplan.Plan{Schema: bodyplan.Schema, ID: "typed-source-tree", Name: name,
		ResultType: decision.TypeInt, Expressions: b.expressions[:b.expressionCount],
		Statements: b.statements[:b.statementCount], Root: root}
	if _, err := bodyplan.Compile(plan, nil); err != nil {
		return nil, false, err
	}
	tree, err := json.Marshal(plan)
	return tree, b.normalizedCondition, err
}
