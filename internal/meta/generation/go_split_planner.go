package generation

import (
	"go/ast"

	"github.com/kimjooyoon/meta-ontology-go/internal/gosplit"
)

// NewGoSplitPlanner retains the same Gooo-authorized normalization used by
// execution. Native capacity planning cannot substitute a different policy.
func NewGoSplitPlanner() gosplit.Planner {
	return gosplit.New(func(file *ast.File, group *ast.GenDecl, specs []*ast.ImportSpec) ([]*ast.GenDecl, bool, error) {
		return NormalizeImportHeaderGroup(file, group, specs, ImportHeaderNormalizationNamedAlias)
	})
}
