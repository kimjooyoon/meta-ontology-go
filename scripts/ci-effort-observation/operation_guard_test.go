package main

import "testing"

func guardedOperationFixture() ([]byte, OperationSpec, APIJob) {
	source := []byte("jobs:\n  check:\n    name: check\n    steps:\n      - name: Evidence\n        run: gofmt -l .\n      - name: Guard\n        run: test -s receipt.json\n")
	spec := OperationSpec{ID: "check", JobName: "check", StepName: "Guard", EvidenceStepName: "Evidence",
		GuardStepName: "Guard", Kind: "VERIFICATION", Command: []string{"gofmt", "-l", "."}, ProofObligationID: "ci-effort/check"}
	job := APIJob{ID: 1, Name: "check", Conclusion: "failure", Steps: []APIStep{
		{Name: "Evidence", Status: "completed", Conclusion: "success", StartedAt: "2026-08-30T00:00:00Z", CompletedAt: "2026-08-30T00:00:01Z"},
		{Name: "Guard", Status: "completed", Conclusion: "failure"},
	}}
	return source, spec, job
}

func TestRejectedGuardObservationRemainsValidFailure(t *testing.T) {
	for _, conclusion := range []string{"failure", "cancelled", "skipped", "timed_out"} {
		t.Run(conclusion, func(t *testing.T) {
			source, spec, job := guardedOperationFixture()
			job.Steps[1].Conclusion = conclusion
			operations, accounting := observeOperations([]OperationSpec{spec}, []APIJob{job}, ".github/workflows/ci.yml", source, nil, "")
			op := operations[0]
			if op.State != "REJECTED" || op.RejectionReason != "GUARD_STEP_NOT_SUCCESSFUL" || op.GuardBound ||
				op.GuardStepConclusion != conclusion || op.WallMS != 0 || accounting.Rejected != 1 || accounting.Executed != 0 {
				t.Fatal("failed guard became a successful operation", op, accounting)
			}
			if err := validateOperations(operations, []OperationSpec{spec}, ".github/workflows/ci.yml", source, ""); err != nil {
				t.Fatal("original failure observation was lost", err)
			}
		})
	}
}

func TestDuplicateObservationsBeforeGuardKeepTheirRejection(t *testing.T) {
	for _, kind := range []string{"JOB", "STEP", "GUARD_STEP"} {
		t.Run(kind, func(t *testing.T) {
			source, spec, job := guardedOperationFixture()
			jobs := []APIJob{job}
			switch kind {
			case "JOB":
				jobs = append(jobs, job)
			case "STEP":
				jobs[0].Steps = append(jobs[0].Steps, job.Steps[0])
			case "GUARD_STEP":
				jobs[0].Steps = append(jobs[0].Steps, job.Steps[1])
			}
			op := observeOperation(spec, jobs, ".github/workflows/ci.yml", source, nil)
			if op.State != "REJECTED" || op.RejectionReason != "DUPLICATE_"+kind+"_OBSERVATION" {
				t.Fatal("duplicate lost its cause", op)
			}
			if err := validateOperations([]OperationObservation{op}, []OperationSpec{spec}, ".github/workflows/ci.yml", source, ""); err != nil {
				t.Fatal("duplicate observation was lost", err)
			}
		})
	}
}

func TestGuardFailureCannotAuthorizeExecutionOrContradictItsCause(t *testing.T) {
	mutations := map[string]func(*OperationObservation){
		"executed":   func(op *OperationObservation) { op.State, op.RejectionReason = "EXECUTED", "" },
		"skipped":    func(op *OperationObservation) { op.State, op.RejectionReason, op.Skipped = "SKIPPED", "", true },
		"success":    func(op *OperationObservation) { op.GuardStepConclusion = "success" },
		"bound":      func(op *OperationObservation) { op.GuardBound = true },
		"unobserved": func(op *OperationObservation) { op.GuardStepStatus, op.GuardStepConclusion = "", "" },
		"duplicate":  func(op *OperationObservation) { op.RejectionReason = "DUPLICATE_GUARD_STEP_OBSERVATION" },
		"duration":   func(op *OperationObservation) { op.WallMS = 1 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			source, spec, job := guardedOperationFixture()
			op := observeOperation(spec, []APIJob{job}, ".github/workflows/ci.yml", source, nil)
			mutate(&op)
			op.EvidenceDigest = operationDigest(op)
			if err := validateOperations([]OperationObservation{op}, []OperationSpec{spec}, ".github/workflows/ci.yml", source, ""); err == nil {
				t.Fatal("contradictory guard observation accepted", op)
			}
		})
	}
}

func TestOperationGuardKeepsSuccessMissingAndLaterFailureDistinct(t *testing.T) {
	for _, mode := range []string{"success", "missing", "later-failure"} {
		t.Run(mode, func(t *testing.T) {
			source, spec, job := guardedOperationFixture()
			job.Steps[1].Conclusion = "success"
			want := "EXECUTED"
			if mode == "missing" {
				job.Steps = job.Steps[:1]
				want = "UNKNOWN"
			} else if mode == "later-failure" {
				job.Steps[0].CompletedAt = "2026-08-29T00:00:00Z"
				want = "REJECTED"
			}
			op := observeOperation(spec, []APIJob{job}, ".github/workflows/ci.yml", source, nil)
			if op.State != want || op.GuardBound != (mode != "missing") {
				t.Fatal("guard outcomes were conflated", mode, op)
			}
			if err := validateOperations([]OperationObservation{op}, []OperationSpec{spec}, ".github/workflows/ci.yml", source, ""); err != nil {
				t.Fatal("valid operation observation rejected", mode, err)
			}
		})
	}
}
