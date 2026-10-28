package transportorigins

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	transport "github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/proof-transport"
	resolution "github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/transport-resolution"
)

func TestSavedReassessmentUsesANewReadAndPreservesPriorFailure(t *testing.T) {
	for _, row := range []struct {
		name, expected string
		status         int
		store, cancel  bool
	}{
		{"applied-unresolved", "READ_REPORT_EFFECT_OBSERVED", 201, true, true},
		{"failed-applied-unresolved", "READ_REPORT_EFFECT_OBSERVED", 500, true, true},
		{"failed-not-observed", "READ_REPORT_EFFECT_NOT_OBSERVED", 500, false, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			prior, raw, wire := runSavedReport(t, true, row.status, row.store, row.cancel)
			original, requests, puts := bytes.Clone(raw), wire.requests, wire.puts
			pending, err := ReassessSavedReports(raw, nil)
			if err != nil || pending.F_state != "OBSERVATION_REQUIRED" || !pending.F_observation_needed || wire.requests != requests {
				t.Fatal("source must request another observation without dispatching it", err, pending)
			}
			wire.cancel = nil
			current := transport.Run(context.Background(), wire.scope.F_operation, wire.material, "", false)
			currentRaw, err := json.Marshal(current)
			if err != nil {
				t.Fatal(err)
			}
			requests = wire.requests
			result, err := ReassessSavedReports(raw, currentRaw)
			if err != nil || result.F_state != row.expected || result.F_observation_needed || wire.requests != requests || wire.puts != puts {
				t.Fatal("reassessment must use the read report without another write", err, result)
			}
			if result.F_prior_state != prior.Outcome.F_state || result.F_prior_command_failed != prior.Outcome.F_command_failed ||
				result.F_prior_report_sha256 != contentDigest(raw) || result.F_current_report_sha256 != contentDigest(currentRaw) || !bytes.Equal(raw, original) {
				t.Fatal("prior failure or original report correspondence changed", result)
			}
			if result.F_read_report_effect_observed != row.store || result.F_effect_id != prior.Scope.F_effect_id ||
				result.F_generation_sha256 != contextDigest(resolution.GoooGenerationOrigin) {
				t.Fatal("effect observation lost its canonical scope", result)
			}
		})
	}
}
