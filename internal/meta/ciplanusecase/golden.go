package ciplanusecase

import "github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"

func ProjectGolden(report metainvocation.Report) GoldenPlan {
	golden := GoldenPlan{Schema: "gooo/ci-plan-golden/v1", CaseID: report.CaseID, Checks: make([]GoldenCheck, 0, len(report.Plan.Checks))}
	for _, check := range report.Plan.Checks {
		projected := GoldenCheck{ID: check.ID, Command: check.Command, Files: append([]string(nil), check.Files...), Reasons: make([]GoldenReason, 0, len(check.Reasons))}
		for _, reason := range check.Reasons {
			projected.Reasons = append(projected.Reasons, GoldenReason{Operation: reason.Operation, File: reason.File, SourcePath: reason.Source.Path, SourceLine: reason.Source.StartLine})
		}
		golden.Checks = append(golden.Checks, projected)
	}
	return golden
}
