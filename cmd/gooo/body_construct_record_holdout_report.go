package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func constructionRecordHoldoutRows(c bodyexecution.JointConstruction) [][2]string {
	var rows [][2]string
	for _, step := range c.Initial.ConstructionSteps() {
		r := step.Generation.Report.RecordAssembly
		if r != nil && r.Holdout != nil {
			label := "Initial record body holdout"
			if r.Holdout.Schema == "gooo/scalar-criteria-holdout/v1" {
				label = "Initial scalar body holdout"
			}
			rows = append(rows, [2]string{label, reportedValue(step.Generation.Report.Activity) + ": " + constructionRecordHoldoutShare(r.Holdout)})
		}
	}
	for _, held := range c.FinalRecordHoldouts {
		name := "missing activity"
		if held.Evaluation != nil {
			name = reportedValue(held.Evaluation.Activity)
		}
		label := "Final selected record holdout"
		if held.Evaluation != nil && held.Evaluation.Schema == "gooo/scalar-criteria-holdout/v1" {
			label = "Final selected scalar holdout"
		}
		rows = append(rows, [2]string{label, name + ": " + constructionRecordHoldoutShare(held.Evaluation)})
		e := held.CallerExposure
		rows = append(rows, [2]string{"Declared holdout inputs across caller attempts", fmt.Sprintf("%d observed; %d other; %d unknown", len(e.ObservedInputs), len(e.OtherInputs), len(e.UnknownInputs))})
	}
	if len(rows) > 0 {
		rows = append(rows, [2]string{"Declared holdout selection effect", "measured after the corresponding selection; cases and fields have separate denominators; caller overlap is recorded separately"})
	}
	return rows
}

func constructionRecordHoldoutShare(r *bodycodegen.RecordHoldoutReceipt) string {
	if r == nil {
		return "not recorded"
	}
	if r.AccuracyPercent == nil {
		return "unavailable: " + reportedValue(r.Error)
	}
	if r.Schema == "gooo/scalar-criteria-holdout/v1" {
		return "Integer cases " + constructionShare(r.Passed, r.Total)
	}
	return fmt.Sprintf("cases %s; fields %s", constructionShare(r.Passed, r.Total), constructionShare(r.FieldsPassed, r.FieldsTotal))
}
