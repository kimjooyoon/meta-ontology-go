package transportorigins

import (
	"context"
	"encoding/json"
	"testing"

	transport "github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/proof-transport"
)

func TestSavedReassessmentRequiresAConsistentReadOnlyReport(t *testing.T) {
	prior, raw, wire := runSavedReport(t, true, 500, true, true)
	wire.cancel = nil
	current := transport.Run(context.Background(), wire.scope.F_operation, wire.material, "", false)
	encode := func(value any) []byte {
		t.Helper()
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	currentRaw := encode(current)
	var extended map[string]json.RawMessage
	if err := json.Unmarshal(currentRaw, &extended); err != nil {
		t.Fatal(err)
	}
	extended["Scope"] = json.RawMessage(`{"effect_id":"shadow"}`)
	altered := current
	altered.Scope.F_effect_id = "another-effect"
	unsupported := transport.Run(context.Background(), "unsupported", wire.material, "", false)
	requests, puts := wire.requests, wire.puts
	for _, row := range []struct {
		name, state string
		current     []byte
		wantError   bool
	}{
		{"write-report", "READ_ONLY_REPORT_REQUIRED", raw, false},
		{"altered-scope", "READ_REPORT_SOURCE_UNRESOLVED", encode(altered), false},
		{"other-operation", "READ_REPORT_SOURCE_UNRESOLVED", encode(unsupported), false},
		{"case-alias", "READ_REPORT_EFFECT_OBSERVED", encode(extended), false},
		{"empty-is-not-omitted", "", []byte{}, true},
		{"invalid-json", "", []byte("{"), true},
	} {
		t.Run(row.name, func(t *testing.T) {
			result, err := ReassessSavedReports(raw, row.current)
			if (err != nil) != row.wantError || err == nil && result.F_state != row.state {
				t.Fatal(err, result)
			}
			if wire.requests != requests || wire.puts != puts {
				t.Fatal("query dispatched an operation")
			}
		})
	}
	prior.Outcome.F_effect_id = "another-effect"
	result, err := ReassessSavedReports(encode(prior), nil)
	if err != nil || result.F_state != "UNRESOLVED_SOURCE_PAIR" {
		t.Fatal("contradictory prior result must retain unresolved source", err, result)
	}
}

func TestSavedReassessmentRetainsAnObservedFailure(t *testing.T) {
	prior, raw, wire := runSavedReport(t, true, 500, true, false)
	requests := wire.requests
	result, err := ReassessSavedReports(raw, nil)
	if err != nil || result.F_state != "RETAIN_PRIOR_ASSESSMENT" || result.F_observation_needed ||
		!result.F_prior_command_failed || result.F_prior_state != prior.Outcome.F_state || wire.requests != requests {
		t.Fatal("a later assessment must retain the original response failure", err, result)
	}
}
