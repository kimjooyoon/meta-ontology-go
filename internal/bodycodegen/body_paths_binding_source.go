package bodycodegen

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func typedPathActivity(filename string, source []byte, name string, prepared *pathplan.PreparedPlan) (*syntax.File, *syntax.ActivityDecl, error) {
	file, diagnostics := ParseBodyFile(filename, source)
	if diagnostics.HasErrors() || file == nil || file.Package == nil {
		return nil, nil, fmt.Errorf("typed path source must be a valid Gooo package")
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if candidate, ok := declaration.(*syntax.ActivityDecl); ok && candidate.Name == name {
			if activity != nil {
				return nil, nil, fmt.Errorf("typed path source has duplicate activities")
			}
			activity = candidate
		}
	}
	if activity == nil || !activity.ValueProgramPresent || len(activity.Inputs) != 1 ||
		activity.Inputs[0].Name != "Integer" || activity.Output != "Integer" || prepared.ActivityName() != name {
		return nil, nil, fmt.Errorf("typed path plan must match one source Integer -> Integer activity")
	}
	return file, activity, nil
}

func bindTypedFallback(ctx context.Context, file *syntax.File, activity *syntax.ActivityDecl,
	prepared *pathplan.PreparedPlan, base Result, receipt *BodyPathReceipt) error {
	fallback := prepared.Fallback()
	fallbackBody, err := rewriteLetDeclarations(fallback.GoooBody())
	if err != nil {
		return err
	}
	fallbackRoute, err := generateRoute(file.Package.Name, activity.Name, base.Report.ActivityID,
		"int64", "int64", fallbackBody, preserveRoute)
	if err != nil {
		return err
	}
	planningBody, err := typedPathPlanningBody(ctx, activity, file.Package.Name, base.Report.ActivityID, prepared)
	if err != nil {
		return err
	}
	originalBody, err := rewriteLetDeclarations(planningBody)
	if err != nil {
		return err
	}
	receipt.SourceBinding, err = routeEquivalence(file.Package.Name, activity.Name, "int64", "int64",
		originalBody, fallbackRoute.source, "typed_path_fallback_matches_authoritative_source")
	if err == nil && !receipt.SourceBinding.Equivalent {
		receipt.SourceBinding, err = typedBodyTreeEquivalence(ctx, activity.Name, originalBody, fallbackRoute.source)
	}
	if err != nil || !receipt.SourceBinding.Equivalent {
		return fmt.Errorf("typed path fallback does not match the authoritative source body")
	}
	return nil
}

func typedPathPlanningBody(ctx context.Context, activity *syntax.ActivityDecl, packageName, activityID string,
	prepared *pathplan.PreparedPlan) (string, error) {
	planningBody := activity.ValueProgram
	if activity.Assembly == nil || activity.Assembly.Spec.Baseline == "" {
		return planningBody, nil
	}
	planningBody = activity.Assembly.Spec.Baseline
	choices := make(map[string]string, len(activity.Assembly.Spec.Picked))
	for _, pick := range activity.Assembly.Spec.Picked {
		choices[pick.ID] = pick.Label
	}
	current, err := prepared.Compile(choices)
	if err != nil {
		return "", err
	}
	if err := verifySourceCheckpoint(ctx, activity, packageName, activityID, current.GoooBody()); err != nil {
		return "", err
	}
	return planningBody, nil
}
