package main

import "testing"

func TestReadOnlyTimingPreservesMissingRuntimeAsUnknown(t *testing.T) {
	source := sourceRunInput{ID: 42, RunStartedAt: "2026-08-30T00:00:00Z", UpdatedAt: "2026-08-30T00:00:02Z"}
	jobs, window, err := observeJobsWithSource([]APIJob{{ID: 7, RunID: 42, Name: "check", Status: "completed", Conclusion: "failure", Steps: []APIStep{{Name: "Verify", Status: "completed", Conclusion: "failure"}}}}, source)
	if err != nil || len(jobs) != 1 || jobs[0].Unknown == nil || jobs[0].WallMS != 0 || jobs[0].Steps[0].Unknown == nil || jobs[0].Steps[0].WallMS != 0 {
		t.Fatalf("missing source timing was not kept unknown: jobs=%+v window=%+v err=%v", jobs, window, err)
	}
	timing := summarizeReadOnlyTiming(window, jobs, false)
	if timing.ObservedJobIntervals != 0 || timing.ObservedStepIntervals != 0 || timing.MissingJobIntervals != 1 || timing.MissingStepIntervals != 1 || timing.WindowWallMS != 2000 {
		t.Fatalf("read-only timing summary lost missing intervals: %+v", timing)
	}
}

func TestReadOnlyOperationCountsKeepUnknownOperationsSeparate(t *testing.T) {
	specs := []OperationSpec{{ID: "check", JobName: "check", StepName: "Verify", Kind: "VERIFICATION", Command: []string{"go", "test"}, ProofObligationID: "ci-effort/check"}}
	operations, accounting := observeOperations(specs, nil, ".github/workflows/ci.yml", nil, nil, "push")
	if len(operations) != 1 || operations[0].State != "UNKNOWN" || operations[0].WallMS != 0 || accounting.Unknown != 1 || accounting.Executed != 0 || accounting.Skipped != 0 || accounting.Rejected != 0 {
		t.Fatalf("unknown operation was not accounted separately: operations=%+v accounting=%+v", operations, accounting)
	}
	counts := readOnlyCounts(accounting)
	if counts.Manifest != 1 || counts.Missing != 1 || counts.Observed != 0 || counts.Skipped != 0 || counts.Rejected != 0 {
		t.Fatalf("read-only operation counts lost missing operation: %+v", counts)
	}
}
