package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Preparation binds the permitted mask to the original source contract before
// observation answers and executable projections are checked.
func prepareSourcePathCandidate(ctx context.Context, filename string, source []byte, activity string,
	prior *BodyPathReceipt, mask uint16) (PathCandidate, *pathplan.PreparedPlan, pathplan.Document, error) {
	var result PathCandidate
	var document pathplan.Document
	set, err := PlanSourcePathCandidates(ctx, filename, source, activity, prior)
	if err != nil {
		return result, nil, document, err
	}
	if !slices.Contains(set.Masks, mask) {
		return result, nil, document, fmt.Errorf("typed path mask exceeds the source attempt cap")
	}
	document, err = DecodeSourcePathDocument(ctx, filename, source, activity, nil)
	if err != nil {
		return result, nil, document, err
	}
	prepared, err := document.Prepare()
	if err != nil || int(mask) >= 1<<len(document.Plan.Decisions) {
		return result, nil, document, fmt.Errorf("typed path candidate exceeds its source palette: %v", err)
	}
	choices := make(map[string]string, len(document.Plan.Decisions))
	for i, choice := range document.Plan.Decisions {
		choices[choice.ID] = choice.Options[(mask>>i)&1].Label
	}
	raw, _ := json.Marshal(document)
	result = PathCandidate{Schema: "gooo/path-candidate/v1", Activity: activity, ActivityID: document.Plan.Base.ID,
		InputSourceSHA256: digest(source), DocumentSHA256: digest(raw), PlanSHA256: prepared.PlanSHA256(),
		Mask: mask, Choices: choices, CandidateSet: set, Stage: "TYPE_CHECK"}
	return result, prepared, document, nil
}
