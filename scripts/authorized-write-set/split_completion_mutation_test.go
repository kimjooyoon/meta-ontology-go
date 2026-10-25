package main

import "testing"

func TestPendingCompletionCannotAuthorizeAnUnhandledPath(t *testing.T) {
	density, extraction, _ := exactReports()
	report := reduce("abc", density, extraction, pendingSplitReceipt(t),
		[]string{"a.go", "c.go", "d.go", "d_split02.go", "pending.go"}, []string{"d_split02.go"})
	if report.Decision != "FAIL_CLOSED" || report.Reason != "WRITE_SET_NOT_EXACT" {
		t.Fatal("an unhandled source became an authorized mutation", report)
	}
}

func TestPendingCompletionRejectsChangedSourceAndCoverage(t *testing.T) {
	for name, mutate := range map[string]func(*splitReport){
		"policy":         func(r *splitReport) { r.PolicySourceDigest = "other-policy" },
		"plan":           func(r *splitReport) { r.PlanSourceDigest = "" },
		"missing policy": func(r *splitReport) { r.PolicySourceDigest = "" },
		"decision":       func(r *splitReport) { r.Decision = "PASS" },
		"resolution":     func(r *splitReport) { r.Resolution = "EXACT" },
		"reason":         func(r *splitReport) { r.Reason = "PROJECTABLE_SUBJECTS_SPLIT" },
		"next":           func(r *splitReport) { r.NextOperation = "none" },
		"input":          func(r *splitReport) { r.Coordinates.Input = 1 },
		"unhandled":      func(r *splitReport) { r.Coordinates.Unhandled = 0 },
		"deferred":       func(r *splitReport) { r.Coordinates.Deferred = 0 },
		"application":    func(r *splitReport) { r.Coordinates.Applied = 0 },
		"unknown":        func(r *splitReport) { r.Coordinates.Unknowns = 1 },
	} {
		t.Run(name, func(t *testing.T) {
			report := pendingSplitReceipt(t)
			mutate(&report)
			if _, _, exact := splitPaths(report); exact {
				t.Fatal("changed completion tuple retained write-set authority")
			}
		})
	}
}
