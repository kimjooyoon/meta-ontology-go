package transportoutcome

import "testing"

func TestSourceOwnedWriteAndReadbackAssessment(t *testing.T) {
	for _, row := range []struct {
		name, decision, readback, responseError, want, next string
		attempted, recorded, failed, observed               bool
		status                                              int64
	}{
		{"read-only-proposal", "READY_TO_RESTORE", "", "", "NO_WRITE_ATTEMPT", "put-pinned-manifest-once", false, false, false, false, 0},
		{"already-present", "ALREADY_PRESENT", "", "", "NO_WRITE_ATTEMPT", "retain-current-observation", false, false, false, false, 0},
		{"not-sent", "READY_TO_RESTORE", "", "cancelled before send", "WRITE_NOT_SENT", "refresh-before-new-attempt", false, false, true, false, 0},
		{"observed", "READY_TO_RESTORE", "ALREADY_PRESENT", "", "EFFECT_OBSERVED", "retain-current-observation", true, true, false, true, 201},
		{"effect-with-http-failure", "READY_TO_RESTORE", "ALREADY_PRESENT", "", "EFFECT_OBSERVED_WITH_WRITE_FAILURE", "retain-current-observation", true, true, true, true, 500},
		{"effect-with-client-error", "READY_TO_RESTORE", "ALREADY_PRESENT", "response interrupted", "EFFECT_OBSERVED_WITH_WRITE_FAILURE", "retain-current-observation", true, true, true, true, 201},
		{"failed-absent", "READY_TO_RESTORE", "READY_TO_RESTORE", "", "EFFECT_NOT_OBSERVED", "revise-before-new-attempt", true, true, true, false, 500},
		{"accepted-absent", "READY_TO_RESTORE", "READY_TO_RESTORE", "", "EFFECT_NOT_OBSERVED", "revise-before-new-attempt", true, true, false, false, 201},
		{"no-readback", "READY_TO_RESTORE", "", "", "UNRESOLVED_EFFECT", "read-current-effect", true, false, false, false, 201},
		{"expired-readback", "READY_TO_RESTORE", "REFRESH_OBSERVATION", "", "UNRESOLVED_EFFECT", "read-current-effect", true, true, true, false, 500},
		{"unknown-readback", "READY_TO_RESTORE", "future-state", "", "UNRESOLVED_EFFECT", "read-current-effect", true, true, false, false, 201},
	} {
		t.Run(row.name, func(t *testing.T) {
			next := "put-pinned-manifest-once"
			if row.decision == "ALREADY_PRESENT" {
				next = "retain-current-observation"
			}
			input := R_Observation{F_effect_id: "urn:effect:original", F_decision_effect_id: "urn:effect:original",
				F_decision_state: row.decision, F_decision_next: next, F_write_attempted: row.attempted,
				F_response_status: row.status, F_response_error: row.responseError, F_readback_recorded: row.recorded,
				F_readback_state: row.readback}
			if row.recorded {
				input.F_readback_effect_id = input.F_effect_id
			}
			got := Evaluate(input)
			if got.F_state != row.want || got.F_next_operation != row.next || got.F_command_failed != row.failed || got.F_response_failed != (row.attempted && row.failed) || got.F_effect_observed != row.observed || got.F_effect_id != input.F_effect_id {
				t.Fatal(got, row)
			}
		})
	}
}

func TestOutcomeDeclinesInconsistentEffectRecords(t *testing.T) {
	for _, change := range []func(*R_Observation){
		func(i *R_Observation) { i.F_effect_id = "" },
		func(i *R_Observation) { i.F_decision_effect_id = "urn:effect:another" },
		func(i *R_Observation) { i.F_readback_effect_id = "urn:effect:another" },
		func(i *R_Observation) { i.F_write_attempted = false },
		func(i *R_Observation) { i.F_decision_state = "ALREADY_PRESENT" },
		func(i *R_Observation) { i.F_readback_recorded = false },
		func(i *R_Observation) {
			i.F_write_attempted, i.F_readback_recorded = false, false
			i.F_readback_effect_id, i.F_readback_state = "", ""
		},
	} {
		input := R_Observation{F_effect_id: "urn:effect:original", F_decision_effect_id: "urn:effect:original",
			F_decision_state: "READY_TO_RESTORE", F_write_attempted: true, F_response_status: 201,
			F_readback_recorded: true, F_readback_effect_id: "urn:effect:original", F_readback_state: "ALREADY_PRESENT"}
		change(&input)
		if got := Evaluate(input); got.F_state != "INCONSISTENT_OBSERVATION" || !got.F_command_failed || got.F_effect_observed {
			t.Fatal(got)
		}
	}
}

func TestOutcomeHTTPResponseBoundaries(t *testing.T) {
	for _, row := range []struct {
		status int64
		failed bool
	}{{199, true}, {299, false}, {300, true}} {
		got := Evaluate(R_Observation{F_effect_id: "urn:effect:original", F_decision_effect_id: "urn:effect:original",
			F_decision_state: "READY_TO_RESTORE", F_write_attempted: true, F_response_status: row.status})
		if got.F_state != "UNRESOLVED_EFFECT" || got.F_command_failed != row.failed || got.F_response_failed != row.failed || got.F_effect_observed {
			t.Fatal(got, row)
		}
	}
}
