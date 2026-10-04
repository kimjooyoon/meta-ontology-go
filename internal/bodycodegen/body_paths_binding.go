package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type typedPathSource struct {
	base     Result
	activity *syntax.ActivityDecl
}

// Shared source authority check for generation and explicit training export.
// It generates validation projections, never a selected candidate or model call.
func bindTypedPathSource(ctx context.Context, filename string, source []byte, activityName string,
	prepared *pathplan.PreparedPlan, receipt *BodyPathReceipt) (typedPathSource, error) {
	return bindTypedPathSourceProjection(ctx, filename, source, activityName, prepared, receipt, nil)
}

// A recipe has already generated its original projection in this request.
// Reuse only that exact source/activity result; fallback equivalence is still
// checked here. Ordinary generation/export/replay always generate a fresh base.
func bindTypedPathSourceProjection(ctx context.Context, filename string, source []byte, activityName string,
	prepared *pathplan.PreparedPlan, receipt *BodyPathReceipt, projection *Result) (typedPathSource, error) {
	bindingStarted := time.Now()
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() || file == nil || file.Package == nil {
		return typedPathSource{}, fmt.Errorf("typed path source must be a valid Gooo package")
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if candidate, ok := declaration.(*syntax.ActivityDecl); ok && candidate.Name == activityName {
			if activity != nil {
				return typedPathSource{}, fmt.Errorf("typed path source has duplicate activities")
			}
			activity = candidate
		}
	}
	if activity == nil || !activity.ValueProgramPresent || len(activity.Inputs) != 1 ||
		activity.Inputs[0].Name != "Integer" || activity.Output != "Integer" || prepared.ActivityName() != activityName {
		return typedPathSource{}, fmt.Errorf("typed path plan must match one source Integer -> Integer activity")
	}
	base, err := typedPathBindingProjection(ctx, filename, source, activityName, projection)
	if err != nil {
		return typedPathSource{}, err
	}
	fallback := prepared.Fallback()
	fallbackBody, err := rewriteLetDeclarations(fallback.GoooBody())
	if err != nil {
		return typedPathSource{}, err
	}
	fallbackRoute, err := generateRoute(file.Package.Name, activityName, base.Report.ActivityID,
		"int64", "int64", fallbackBody, preserveRoute)
	if err != nil {
		return typedPathSource{}, err
	}
	originalBody, err := rewriteLetDeclarations(activity.ValueProgram)
	if err != nil {
		return typedPathSource{}, err
	}
	receipt.SourceBinding, err = routeEquivalence(file.Package.Name, activityName, "int64", "int64",
		originalBody, fallbackRoute.source, "typed_path_fallback_matches_authoritative_source")
	if err == nil && !receipt.SourceBinding.Equivalent {
		receipt.SourceBinding, err = typedBodyTreeEquivalence(ctx, activityName, originalBody, fallbackRoute.source)
	}
	if err != nil || !receipt.SourceBinding.Equivalent {
		return typedPathSource{}, fmt.Errorf("typed path fallback does not match the authoritative source body")
	}
	receipt.SourceBaseMatched = true
	if err := verifySourceAssembly(ctx, filename, source, activity, base, receipt); err != nil {
		return typedPathSource{}, err
	}
	receipt.Timing.SourceBindingMS = elapsedMS(bindingStarted)
	if err := ctx.Err(); err != nil {
		return typedPathSource{}, err
	}
	return typedPathSource{base: base, activity: activity}, nil
}

func typedPathBindingProjection(ctx context.Context, filename string, source []byte, activityName string,
	projection *Result) (Result, error) {
	if projection == nil {
		return GenerateWithPlanner(ctx, filename, source, activityName, "", "")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	report := projection.Report
	if report.Schema != schema || report.Decision != "PASS" || report.Activity != activityName ||
		report.SourceDigest != digest(source) || report.GeneratedDigest != digest([]byte(projection.Source)) ||
		report.InputType != "int64" || report.OutputType != "int64" ||
		!report.TypecheckPassed || !report.DeterministicReplay || report.RepositoryWrites != 0 {
		return Result{}, fmt.Errorf("recipe projection must match the current checked source and activity")
	}
	return *projection, nil
}

func bindTypedPathDocument(document pathplan.Document, receipt *BodyPathReceipt) error {
	documentBytes, err := json.Marshal(document)
	if err != nil || len(documentBytes) > 128<<10 {
		return fmt.Errorf("typed path document exceeds its byte budget")
	}
	receipt.DocumentSHA256 = digest(documentBytes)
	testBytes, _ := json.Marshal(document.TestCases)
	receipt.TestSuiteSHA256 = digest(testBytes)
	receipt.DeclaredTestCases = len(document.TestCases)
	return nil
}
