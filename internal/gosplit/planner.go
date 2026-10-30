// Package gosplit contains the read-only declaration planner used by the
// source-splitter executor. It neither publishes files nor verifies a projected
// package; those remain obligations of the operation's existing consumers.
package gosplit

import (
	"errors"
	"go/ast"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricevidence"
)

var ErrBlocked = errors.New("split blocked")

type Part struct {
	Path             string
	Subject          string
	Data             []byte
	DeclarationOrder []DeclarationOrder
}

type DeclarationOrder struct {
	Ordinal int    `json:"ordinal"`
	Digest  string `json:"digest"`
}

type Plan struct {
	Directory string
	Mode      os.FileMode
	Parts     []Part
}

// ImportNormalizer is supplied by the Gooo operation-contract adapter. Keeping
// it explicit prevents the native planner from inventing another header policy.
type ImportNormalizer func(*ast.File, *ast.GenDecl, []*ast.ImportSpec) ([]*ast.GenDecl, bool, error)

type Planner struct {
	normalize ImportNormalizer
}

func New(normalize ImportNormalizer) Planner { return Planner{normalize: normalize} }

// Prepare reuses the executor's actual render and directory checks without
// writing. The caller must bind the metrics to its source revision. A refusal
// returns no installable parts, including when directory evidence is absent.
func (planner Planner) Prepare(root, subject string, report metricevidence.Report) (Plan, error) {
	plan, err := planner.PlanSource(root, subject, report.Meta.Policy.MaxFileLines)
	if err != nil {
		return Plan{}, err
	}
	if err := ValidateTopology(report, plan); err != nil {
		return Plan{}, err
	}
	return plan, nil
}
