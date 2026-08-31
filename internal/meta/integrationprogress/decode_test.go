package integrationprogress

import "testing"

func TestObservationRejectsUnknownFields(t *testing.T) {
	raw := []byte(`{"schema":"gooo/integration-progress-observation/v1","repository":"kimjooyoon/meta-ontology-go","observer_head_sha":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","observed_at":"2026-08-28T00:00:00Z","cohort_id":"gooo.portfolio.pr-541-570.v1","pull_requests":[],"future":true}`)
	if _, err := DecodeObservation(raw); err == nil {
		t.Fatal("unknown observation field was accepted")
	}
}
