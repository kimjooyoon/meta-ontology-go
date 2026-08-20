package main

import (
	"strings"
	"testing"
)

func TestCIProofRequiresExactIndependentSchedulerVector(t *testing.T) {
	head := strings.Repeat("a", 40)
	evidence := validSchedulerInput(head, 1, 1)
	lagJobs := validProof().Jobs
	for index := range lagJobs {
		lagJobs[index].Status = stringPointer("in_progress")
		lagJobs[index].Conclusion = nil
		lagJobs[index].CompletedAt = nil
		lagJobs[index].ObservationState = observerLag
		if state, err := jobObservationState(lagJobs[index], evidence[index]); err != nil || state != observerLag {
			t.Fatalf("observer-lag fixture was not accepted as typed lag: state=%q err=%v", state, err)
		}
	}
	mutations := []struct {
		name   string
		mutate func([]schedulerInput) []schedulerInput
	}{
		{name: "forged", mutate: func(final []schedulerInput) []schedulerInput {
			final[0].Result = "failure"
			return final
		}},
		{name: "reordered", mutate: func(final []schedulerInput) []schedulerInput {
			final[0], final[1] = final[1], final[0]
			return final
		}},
		{name: "omitted", mutate: func(final []schedulerInput) []schedulerInput {
			return final[:len(final)-1]
		}},
	}
	for _, test := range mutations {
		t.Run(test.name, func(t *testing.T) {
			final := append([]schedulerInput(nil), evidence...)
			final = test.mutate(final)
			if err := validateSchedulerAgreement(evidence, final, head, 1, 1); err == nil {
				t.Fatal("proof accepted a scheduler vector that differed from evidence")
			}
		})
	}
	if err := validateSchedulerAgreement(evidence, append([]schedulerInput(nil), evidence...), head, 1, 1); err != nil {
		t.Fatalf("matching independent scheduler vector was rejected: %v", err)
	}
}
