package main

import (
	"fmt"
	"testing"
)

func TestCountercaseExecutionAccountingFresh1687(t *testing.T) {
	for _, test := range []struct {
		name             string
		row              caseResult
		err              error
		executed, passed int
	}{
		{"harness-error", caseResult{HarnessError: "input failure"}, fmt.Errorf("input failure"), 0, 0},
		{"unexecuted-pass-claim", caseResult{Passed: true}, nil, 0, 0},
		{"reported-refusal", caseResult{ConsumerAttempted: true, Passed: true}, nil, 1, 1},
		{"reported-acceptance", caseResult{ConsumerAttempted: true, ReadinessCredit: true}, nil, 1, 0},
		{"producer-only", caseResult{ProducerAttempted: true, ProducerError: "refused"}, nil, 0, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := report{CasesDeclared: 1}
			recordCase(&result, test.row, test.err)
			if result.CasesAttempted != 1 || result.CasesExecuted != test.executed || result.CasesPassed != test.passed {
				t.Fatalf("case accounting confused a claim with execution: %+v", result)
			}
		})
	}
}
