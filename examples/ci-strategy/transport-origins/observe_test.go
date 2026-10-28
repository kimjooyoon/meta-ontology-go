package transportorigins

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	transport "github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/proof-transport"
)

func TestObserveForReassessmentCollectsOnlyTheSourceSelectedRead(t *testing.T) {
	for _, row := range []struct {
		name, state string
		status      int
		store       bool
	}{
		{"applied-unresolved", "READ_REPORT_EFFECT_OBSERVED", 201, true},
		{"failed-applied", "READ_REPORT_EFFECT_OBSERVED", 500, true},
		{"failed-absent", "READ_REPORT_EFFECT_NOT_OBSERVED", 500, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			prior, raw, wire := runSavedReport(t, true, row.status, row.store, true)
			wire.cancel = nil
			original, requests, puts := bytes.Clone(raw), wire.requests, wire.puts
			pending, err := ReassessSavedReports(raw, nil)
			if err != nil || !pending.F_observation_needed || pending.F_read_operation != prior.Scope.F_operation {
				t.Fatal("source must select the actual read operation", err, pending)
			}
			currentRaw, result, err := ObserveForReassessment(context.Background(), raw, wire.material, "")
			if err != nil || result.F_state != row.state || result.F_observation_needed || result.F_read_operation != "" {
				t.Fatal("selected read must return to the source assessment", err, result)
			}
			var current transport.Report
			if err := json.Unmarshal(currentRaw, &current); err != nil {
				t.Fatal(err)
			}
			if current.ApplyRequested || current.Write.Attempted || current.After != nil || current.Next != nil ||
				current.Requests <= 0 || int(current.Requests) != wire.requests-requests || wire.puts != puts {
				t.Fatal("one fresh read must not perform another write", current, wire.requests-requests)
			}
			if !bytes.Equal(raw, original) || result.F_prior_command_failed != prior.Outcome.F_command_failed ||
				result.F_prior_state != prior.Outcome.F_state || result.F_prior_report_sha256 != contentDigest(raw) ||
				result.F_current_report_sha256 != contentDigest(currentRaw) {
				t.Fatal("collecting another observation must preserve both report origins", result)
			}
		})
	}
}

func TestObserveForReassessmentKeepsSourceRefusalsLocal(t *testing.T) {
	prior, raw, wire := runSavedReport(t, true, 500, true, false)
	altered := prior
	altered.Scope.F_effect_id = "another-effect"
	contradictory, err := json.Marshal(altered)
	if err != nil {
		t.Fatal(err)
	}
	requests, puts := wire.requests, wire.puts
	for _, row := range []struct {
		name, state string
		prior       []byte
		wantError   bool
	}{
		{"observed-failure", "RETAIN_PRIOR_ASSESSMENT", raw, false},
		{"contradictory", "UNRESOLVED_SOURCE_PAIR", contradictory, false},
		{"malformed", "", []byte("{"), true},
	} {
		t.Run(row.name, func(t *testing.T) {
			current, result, err := ObserveForReassessment(context.Background(), row.prior, nil, "")
			if (err != nil) != row.wantError || err == nil && result.F_state != row.state || current != nil ||
				result.F_observation_needed || result.F_read_operation != "" || wire.requests != requests || wire.puts != puts {
				t.Fatal("unselected reads must not load material or dispatch HTTP", err, result)
			}
		})
	}
}

func TestObserveForReassessmentRetainsCancelledRead(t *testing.T) {
	prior, raw, wire := runSavedReport(t, true, 500, true, true)
	wire.cancel = nil
	requests, puts := wire.requests, wire.puts
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	current, result, err := ObserveForReassessment(ctx, raw, wire.material, "")
	if err != nil || len(current) == 0 || result.F_state != "READ_REPORT_UNRESOLVED" || !result.F_observation_needed ||
		result.F_read_operation != prior.Scope.F_operation || !result.F_prior_command_failed ||
		wire.requests != requests || wire.puts != puts {
		t.Fatal("cancelled collection must remain an unresolved observation", err, result)
	}
}
