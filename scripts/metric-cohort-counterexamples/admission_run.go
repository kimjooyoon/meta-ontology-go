package main

import (
	"fmt"
	"time"
)

func observeAdmissionControls(cfg config, f fixture) (result admissionReport, observationErr error) {
	started := time.Now()
	defer func() { result.ElapsedNanoseconds = time.Since(started).Nanoseconds() }()
	result.Decision = "FAIL_CLOSED"
	if f.Receipt.ExpectedCount == 0 || len(f.Receipt.Expected) == 0 {
		return result, fmt.Errorf("observed admission controls require the declared source cohort")
	}
	planned := observedAdmissionPlans(f.Receipt.Expected[0].Operation)
	result.CasesDeclared = len(planned)
	baseline, receipt, err := observedAdmissionBaseline(cfg, f)
	if err != nil {
		result.SetupError = err.Error()
		return result, err
	}
	result.BaselineReceiptDigest, result.BaselineCredit = receipt.Digest, true
	for _, plan := range planned {
		row, err := observeAdmissionCase(cfg, f, baseline, plan)
		if err != nil {
			row.HarnessError = err.Error()
		}
		result.Cases = append(result.Cases, row)
		if row.Passed && err == nil {
			result.CasesPassed++
		}
	}
	if result.CasesPassed == result.CasesDeclared && len(result.Cases) == result.CasesDeclared {
		result.Decision = "PASS"
		return result, nil
	}
	return result, fmt.Errorf("observed admission controls passed %d/%d", result.CasesPassed, result.CasesDeclared)
}
