package bodycodegen

import (
	"context"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// VerifyIRBodySearchProjection binds a prior selected expression to its
// original source and deterministically reconstructs its emitted Go before
// native execution. It does not repeat model selection or trust the receipt's
// generated source by itself.
func VerifyIRBodySearchProjection(ctx context.Context, filename string, source []byte, prior Result) error {
	search := prior.Report.BodySearch
	if search == nil || search.OriginalSourceDigest != digest(source) || search.SelectedExpression == "" {
		return fmt.Errorf("IR body-search source binding is missing or changed")
	}
	holeID := findSearchHoleID(source, prior.Report.Activity)
	if holeID == "" {
		return fmt.Errorf("IR body-search hole is missing or ambiguous")
	}
	file, activity, activityID, body, err := prepareBodySearch(filename, source, prior.Report.Activity, holeID)
	if err != nil {
		return fmt.Errorf("IR body-search source cannot be replayed: %w", err)
	}
	if activityID != prior.Report.ActivityID {
		return fmt.Errorf("IR body-search activity identity differs")
	}
	hole := bodyFillHoleToken(holeID)
	completedBody, err := replaceIdentifier(body, hole, search.SelectedExpression)
	if err != nil {
		return fmt.Errorf("IR body-search selected expression cannot be restored")
	}
	completedSource, err := replaceActivityProgram(source, activity.ValueProgramSpan, completedBody)
	if err != nil {
		return fmt.Errorf("IR body-search completed source cannot be reconstructed")
	}
	if file == nil || digest(completedSource) != prior.Report.SourceDigest {
		return fmt.Errorf("IR body-search completed source digest differs")
	}
	replayed, err := GenerateWithPlanner(ctx, filename, completedSource, prior.Report.Activity, "", "")
	if err != nil || replayed.Source != prior.Source || replayed.Report.GeneratedDigest != prior.Report.GeneratedDigest ||
		replayed.Report.ReplayDigest != prior.Report.ReplayDigest || replayed.Report.ActivityID != prior.Report.ActivityID {
		return fmt.Errorf("IR body-search Go projection does not replay")
	}
	return nil
}

func findSearchHoleID(source []byte, activityName string) string {
	file, diagnostics := syntax.ParseFile("<body-source>", string(source))
	if diagnostics.HasErrors() || file == nil {
		return ""
	}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != activityName || !activity.ValueProgramPresent {
			continue
		}
		const prefix = "__GOOO_BODY_HOLE_"
		if strings.Count(activity.ValueProgram, prefix) != 1 {
			return ""
		}
		start := strings.Index(activity.ValueProgram, prefix) + len(prefix)
		end := strings.Index(activity.ValueProgram[start:], "__")
		if end <= 0 {
			return ""
		}
		return activity.ValueProgram[start : start+end]
	}
	return ""
}
