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
	left, err := typedBodyTree(ctx, name, []byte("package source\nfunc "+name+"(input int64) int64 {\n"+original+"\n}"))
	if err != nil {
		return RouteEquivalenceReceipt{}, err
	}
	right, err := typedBodyTree(ctx, name, generated)
	if err != nil {
		return RouteEquivalenceReceipt{}, err
	}
	equal, status := bytes.Equal(left, right), "FAIL_CLOSED"
	if equal {
		status = "PASS"
	}
	return RouteEquivalenceReceipt{Schema: routeEquivalenceSchema, Decision: status,
		Method: "canonical_typed_body_tree/v1", Rule: "typed_path_fallback_matches_authoritative_source",
		SourceSemanticDigest: digest(left), GeneratedSemanticDigest: digest(right), Equivalent: equal,
		Scope: "typechecked Integer -> Integer typed-body arena; exact operator, child-edge, local and statement structure"}, nil
}

func typedBodyTree(ctx context.Context, name string, source []byte) ([]byte, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "typed-tree.go", source, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	function, ok := findFunction(file, name)
	if !ok {
		return nil, fmt.Errorf("typed tree function missing")
	}
	b := recipeArena{ctx: ctx}
	root, err := b.sequence(function.Body.List, 0)
	if err != nil {
		return nil, err
	}
	if err := b.includeDeclaredInput(); err != nil {
		return nil, err
	}
	plan := bodyplan.Plan{Schema: bodyplan.Schema, ID: "typed-source-tree", Name: name,
		ResultType: decision.TypeInt, Expressions: b.expressions[:b.expressionCount],
		Statements: b.statements[:b.statementCount], Root: root}
	if _, err := bodyplan.Compile(plan, nil); err != nil {
		return nil, err
	}
	return json.Marshal(plan)
}
