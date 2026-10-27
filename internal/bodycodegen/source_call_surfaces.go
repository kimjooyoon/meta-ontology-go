package bodycodegen

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// Include every permitted alternative, not only a baseline or ranked candidate.
func activityConstructionSyntax(activity *syntax.ActivityDecl) ([]*ast.File, error) {
	surfaces := []string{activity.ValueProgram}
	if activity.Assembly != nil {
		spec := activity.Assembly.Spec
		surfaces = append(surfaces, spec.Baseline)
		for _, choice := range spec.Choices {
			if choice.Kind == "field_value" || choice.Kind == "field_update" {
				surfaces = append(surfaces, "return "+choice.Alternative)
			}
		}
		if spec.FillPlan != nil {
			for _, candidate := range spec.FillPlan.Candidates {
				for _, fill := range candidate.Fills {
					surfaces = append(surfaces, "return "+fill.Expression)
				}
			}
		}
	}
	var bodies []*ast.File
	for _, surface := range surfaces {
		if surface == "" {
			continue
		}
		body, err := rewriteLetDeclarations(surface)
		if err != nil {
			return nil, err
		}
		file, err := parser.ParseFile(token.NewFileSet(), "calls.go", "package p\nvar _ = func(){\n"+body+"\n}", 0)
		if err != nil {
			return nil, err
		}
		bodies = append(bodies, file)
	}
	return bodies, nil
}
