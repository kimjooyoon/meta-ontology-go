package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// SourceAssembly returns the selected activity's detached source-owned contract.
// Absence is distinct from malformed source; no model is loaded or called.
func SourceAssembly(ctx context.Context, filename string, source []byte, activity string) (*assemblyspec.Spec, error) {
	if ctx == nil || len(source) == 0 || len(source) > 128<<10 || !utf8.Valid(source) {
		return nil, fmt.Errorf("assembly requires a context and bounded UTF-8 Gooo source")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, diagnostics := ParseBodyFile(filename, source)
	if diagnostics.HasErrors() || file == nil {
		return nil, fmt.Errorf("assembly source: %v", diagnostics)
	}
	var selected *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if d, ok := declaration.(*syntax.ActivityDecl); ok && d.Name == activity {
			if selected != nil {
				return nil, fmt.Errorf("duplicate assembly activity %q", activity)
			}
			selected = d
		}
	}
	if selected == nil || selected.Assembly == nil {
		return nil, nil
	}
	return selected.Assembly.Spec.Clone(), nil
}

func sourceAssemblyRecipe(spec *assemblyspec.Spec) sourcePathRecipe {
	recipe := sourcePathRecipe{Schema: SourcePathRecipeSchema, MaxAttempts: spec.MaxAttempts, Seed: spec.Seed}
	for _, c := range spec.Choices {
		occurrence := c.Occurrence
		recipe.Choices = append(recipe.Choices, sourceRecipeChoice{ID: c.ID, Kind: c.Kind,
			Occurrence: &occurrence, Intent: c.Intent, Alternative: c.Alternative})
	}
	for _, c := range spec.Cases {
		recipe.TestCases = append(recipe.TestCases, pathplan.TestCase{Input: c.Input, Expected: c.Expected})
	}
	for _, c := range spec.ConditionCases {
		recipe.ConditionCases = append(recipe.ConditionCases, pathplan.ConditionCase{
			ChoiceID: c.ChoiceID, Input: c.Input, Expected: c.Expected})
	}
	return recipe
}

// All entry points bind the complete contract, including expectations and search
// budget, before predictions. Reuse the already checked original projection.
func verifySourceAssembly(ctx context.Context, filename string, source []byte, activity *syntax.ActivityDecl,
	projection Result, receipt *BodyPathReceipt) error {
	if activity.Assembly == nil || receipt.DocumentSHA256 == "" {
		return nil
	}
	document, err := expandSourceRecipeProjection(ctx, filename, source, activity.Name,
		sourceAssemblyRecipe(&activity.Assembly.Spec), &projection, false)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(document)
	if err != nil || digest(raw) != receipt.DocumentSHA256 {
		return fmt.Errorf("typed path document differs from source assembling contract")
	}
	return nil
}
