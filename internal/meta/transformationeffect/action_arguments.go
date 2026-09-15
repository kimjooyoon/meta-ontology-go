package transformationeffect

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func actionArguments(executor, root string, opts Options, plan generation.Plan, action generation.Action, check bool) []string {
	args := []string{"run", "./" + executor, "-root", root, "-metrics", opts.MetricsPath,
		"-sha", opts.ExpectedSHA, "-subject", action.Subject}
	if action.Operation == sourcepolicy.OperationCollapseAssign {
		args = append(args, "-base-sha", plan.BaseSHA)
	}
	if check {
		return append(args, "-check")
	}
	if action.Operation == sourcepolicy.OperationSplitGo {
		return append(args, "-evidence-json")
	}
	return args
}
