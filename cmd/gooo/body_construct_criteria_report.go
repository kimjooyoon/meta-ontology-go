package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func constructionCriteriaRows(c bodyexecution.JointConstruction) [][2]string {
	var rows [][2]string
	if c.SelectedAttempt >= 0 && c.SelectedAttempt < len(c.Attempts) {
		a := c.Attempts[c.SelectedAttempt]
		if a.CriteriaTotal > 0 {
			rows = append(rows, [2]string{"Selected Gooo criteria", fmt.Sprintf("%d/%d satisfied; %d unsatisfied; %d unknown (predicate/input pairs)", a.CriteriaPassed, a.CriteriaTotal, a.CriteriaFailed, a.CriteriaUnknown)})
		}
	}
	appendEvidence := func(label string, e *bodycodegen.RecordCriteriaEvidence) {
		if e == nil {
			return
		}
		rows = append(rows, [2]string{label, fmt.Sprintf("%s: %d/%d satisfied; %d unsatisfied; %d unknown", reportedValue(e.Activity), e.Passed, e.Total, e.Failed, e.Unknown)})
	}
	for _, step := range c.Initial.ConstructionSteps() {
		if a := step.Generation.Report.RecordAssembly; a != nil {
			appendEvidence("Initial Gooo criteria", a.Criteria)
		}
	}
	for _, h := range c.FinalCriteriaHoldouts {
		appendEvidence("Final Gooo criteria on held inputs", h.Evaluation)
		rows = append(rows, [2]string{"Criteria holdout caller overlap", fmt.Sprintf("%d observed; %d other; %d unknown", len(h.CallerExposure.ObservedInputs), len(h.CallerExposure.OtherInputs), len(h.CallerExposure.UnknownInputs))})
	}
	return rows
}
