package bodycodegen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"unicode/utf8"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

const SourcePathRecipeSchema = "gooo/source-typed-path-recipe/v1"

type sourcePathRecipe struct {
	Schema         string                   `json:"schema"`
	Choices        []sourceRecipeChoice     `json:"choices"`
	TestCases      []pathplan.TestCase      `json:"test_cases"`
	ConditionCases []pathplan.ConditionCase `json:"condition_cases,omitempty"`
	MaxAttempts    int                      `json:"max_attempts"`
	Seed           string                   `json:"seed,omitempty"`
}

type sourceRecipeChoice struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Occurrence  *int   `json:"occurrence"`
	Intent      string `json:"intent"`
	Alternative string `json:"alternative_name,omitempty"`
}

// DecodeSourcePathDocument accepts an existing full document or expands a short
// recipe from the authoritative activity. Expansion does no inference or search.
// Its result is an ordinary owned path document, bound again during generation.
func DecodeSourcePathDocument(ctx context.Context, filename string, source []byte, activity string, raw []byte) (pathplan.Document, error) {
	assembly, err := SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return pathplan.Document{}, err
	}
	if assembly != nil {
		if len(raw) != 0 {
			return pathplan.Document{}, fmt.Errorf("source assembling contract already owns the plan; omit the external document")
		}
		return expandSourceRecipe(ctx, filename, source, activity, sourceAssemblyRecipe(assembly))
	}
	if len(raw) == 0 {
		return pathplan.Document{}, fmt.Errorf("activity has no assembling contract; supply a path plan or recipe")
	}
	var header struct {
		Schema string `json:"schema"`
	}
	if len(raw) == 0 || len(raw) > 128<<10 || !utf8.Valid(raw) || decision.RejectDuplicateJSONKeys(raw) != nil {
		return pathplan.Document{}, fmt.Errorf("path document requires bounded strict UTF-8 JSON")
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return pathplan.Document{}, err
	}
	if header.Schema != SourcePathRecipeSchema {
		return pathplan.DecodeDocument(raw)
	}
	if err := recipeExactKeys(raw); err != nil {
		return pathplan.Document{}, err
	}
	var recipe sourcePathRecipe
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&recipe); err != nil {
		return pathplan.Document{}, err
	}
	if d.Decode(&struct{}{}) != io.EOF || len(recipe.Choices) < 1 || len(recipe.Choices) > 16 {
		return pathplan.Document{}, fmt.Errorf("recipe requires 1..16 choices and one JSON object")
	}
	return expandSourceRecipe(ctx, filename, source, activity, recipe)
}

func expandSourceRecipe(ctx context.Context, filename string, source []byte, activity string, recipe sourcePathRecipe) (pathplan.Document, error) {
	return expandSourceRecipeProjection(ctx, filename, source, activity, recipe, nil, true)
}

func expandSourceRecipeProjection(ctx context.Context, filename string, source []byte, activity string,
	recipe sourcePathRecipe, projection *Result, bind bool) (pathplan.Document, error) {
	if ctx == nil || len(source) == 0 || len(source) > 128<<10 || !utf8.Valid(source) {
		return pathplan.Document{}, fmt.Errorf("recipe requires context and bounded UTF-8 Gooo source")
	}
	if err := ctx.Err(); err != nil {
		return pathplan.Document{}, err
	}
	planning, err := sourceAssemblyPlanningSource(ctx, filename, source, activity)
	if err != nil {
		return pathplan.Document{}, err
	}
	planningProjection := projection
	if !bytes.Equal(planning, source) {
		planningProjection = nil
	}
	base, err := typedPathBindingProjection(ctx, filename, planning, activity, planningProjection)
	if err != nil {
		return pathplan.Document{}, err
	}
	if base.Report.InputType != "int64" || base.Report.OutputType != "int64" {
		return pathplan.Document{}, fmt.Errorf("source recipe requires Integer -> Integer")
	}
	document, err := sourceRecipeDocument(ctx, base, activity, recipe)
	if err != nil {
		return pathplan.Document{}, err
	}
	prepared, err := document.Prepare()
	if err != nil {
		return pathplan.Document{}, err
	}
	if bind {
		currentProjection := &base
		if !bytes.Equal(planning, source) {
			currentProjection = projection
		}
		if _, err = bindTypedPathSourceProjection(ctx, filename, source, activity, prepared, &BodyPathReceipt{}, currentProjection); err != nil {
			return pathplan.Document{}, fmt.Errorf("recipe source binding: %w", err)
		}
	}
	return document, nil
}

func sourceRecipeDocument(ctx context.Context, base Result, activity string, recipe sourcePathRecipe) (pathplan.Document, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "recipe.go", base.Source, parser.AllErrors)
	if err != nil {
		return pathplan.Document{}, err
	}
	removeLocalReadMarkers(file, fset)
	function, ok := findFunction(file, activity)
	if !ok {
		return pathplan.Document{}, fmt.Errorf("recipe activity projection missing")
	}
	b := recipeArena{ctx: ctx}
	root, err := b.sequence(function.Body.List, 0)
	if err != nil {
		return pathplan.Document{}, err
	}
	if err := b.includeDeclaredInput(); err != nil {
		return pathplan.Document{}, err
	}
	document := pathplan.Document{Schema: pathplan.DocumentSchema, TestCases: recipe.TestCases,
		MaxAttempts: recipe.MaxAttempts, Seed: recipe.Seed, Plan: pathplan.Plan{Schema: pathplan.Schema,
			ConditionCases: append([]pathplan.ConditionCase(nil), recipe.ConditionCases...),
			Base: bodyplan.Plan{Schema: bodyplan.Schema, ID: base.Report.ActivityID, Name: activity,
				ResultType: decision.TypeInt, Expressions: b.expressions[:b.expressionCount],
				Statements: b.statements[:b.statementCount], Root: root}}}
	for _, spec := range recipe.Choices {
		choice, err := b.choice(spec, root)
		if err != nil {
			return pathplan.Document{}, err
		}
		document.Plan.Decisions = append(document.Plan.Decisions, choice)
	}
	return document, nil
}

// Keep extraction storage request-owned and bounded before SDK preparation.
type recipeArena struct {
	ctx                             context.Context
	expressions                     [128]bodyplan.Expr
	statements                      [128]bodyplan.Stmt
	expressionPositions             [128]token.Pos
	statementPositions              [128]token.Pos
	expressionCount, statementCount int
	inputIndex                      int
	inputSeen                       bool
	normalizedCondition             bool
	normalizedArithmetic            bool
}

// The source signature declares input even when a constant body never reads it.
// Append only when absent so existing recipe indices and documents stay stable.
func (b *recipeArena) includeDeclaredInput() error {
	if b.inputSeen {
		return nil
	}
	_, err := b.expression(&ast.Ident{Name: "input"}, 0)
	return err
}

func (b *recipeArena) bounded(depth int) error {
	if err := b.ctx.Err(); err != nil {
		return err
	}
	if depth > 16 {
		return fmt.Errorf("recipe source nesting exceeds 16")
	}
	return nil
}

func (b *recipeArena) sequence(nodes []ast.Stmt, depth int) ([]int, error) {
	if err := b.bounded(depth); err != nil {
		return nil, err
	}
	var result []int
	for _, node := range nodes {
		index, err := b.statement(node, depth+1)
		if err != nil {
			return nil, err
		}
		result = append(result, index)
	}
	return result, nil
}

func recipeExactKeys(raw []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	for key := range fields {
		switch key {
		case "schema", "choices", "test_cases", "max_attempts", "seed":
		default:
			return fmt.Errorf("unknown recipe field %q", key)
		}
	}
	var choices []map[string]json.RawMessage
	if err := json.Unmarshal(fields["choices"], &choices); err != nil {
		return err
	}
	for _, choice := range choices {
		for key := range choice {
			switch key {
			case "id", "kind", "occurrence", "intent", "alternative_name":
			default:
				return fmt.Errorf("unknown recipe choice field %q", key)
			}
		}
	}
	return nil
}
