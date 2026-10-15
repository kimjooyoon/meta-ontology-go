package main

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func constructionCriteriaRejectionRows(c bodyexecution.JointConstruction) [][2]string {
	var rows [][2]string
	for _, step := range c.Initial.ConstructionSteps() {
		a := step.Generation.Report.RecordAssembly
		if a == nil || a.CriteriaQuestions == nil {
			continue
		}
		r := bodycodegen.ExplainCriteriaRejections(a)
		if r.Status == "NO_QUESTION_HISTORY" || (r.Status == "CANDIDATES_REMAIN" && r.Rejected == 0) {
			continue
		}
		rows = append(rows,
			[2]string{"Initial Gooo rejection explanation", fmt.Sprintf("%s: %s; %d/%d candidates examined; %d rejected, %d explained; %d unavailable, %d unobserved",
				reportedValue(step.Generation.Report.Activity), r.Status, r.Considered, r.CandidateSpace, r.Rejected, r.Explained, r.Unavailable, r.Unobserved)},
			[2]string{"Rejection history source", r.SourceSHA256})
		for _, group := range r.Groups[:min(len(r.Groups), 6)] {
			rows = append(rows, [2]string{"Recorded rejection reason", criteriaRejectionGroupText(group)})
		}
		if len(r.Groups) > 6 {
			rows = append(rows, [2]string{"Further rejection reasons", fmt.Sprintf("%d groups; full question evaluations remain in JSON", len(r.Groups)-6)})
		}
		if len(r.UnexplainedMasks) > 0 {
			rows = append(rows, [2]string{"Rejections without a linked reason", fmt.Sprint(r.UnexplainedMasks)})
		}
		rows = append(rows, [2]string{"Rejection scope", "initial recorded question history and declared candidates; other implementations may satisfy the conditions"})
	}
	return rows
}

func criteriaRejectionGroupText(group bodycodegen.CriteriaRejectionGroup) string {
	condition := "exact case expected " + group.ExpectedJSON
	if group.Criterion != nil {
		condition = group.Criterion.ID + " (" + group.Criterion.Activity + ") returned false"
	}
	var values []string
	for _, witness := range group.Witnesses {
		values = append(values, fmt.Sprintf("candidate %d produced %s", witness.Mask, witness.ActualJSON))
	}
	return fmt.Sprintf("%s for input %s: %s", condition, group.InputJSON, strings.Join(values, "; "))
}
