package transportorigins

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	transport "github.com/kimjooyoon/meta-ontology-go/examples/ci-strategy/proof-transport"
)

func TestReportViewRechecksItsSourceDecisions(t *testing.T) {
	_, original, _ := runSavedReport(t, true, 500, true, false)
	for _, row := range []struct {
		name, reason string
		change       func(*transport.Report)
	}{
		{"scope", "SCOPE_DEFINITION_CONFLICT", func(r *transport.Report) { r.Scope.F_goal_id = "urn:changed-goal" }},
		{"before", "DECISION_SOURCE_CONFLICT", func(r *transport.Report) { r.Before.F_material_digest = "changed" }},
		{"decision", "DECISION_SOURCE_CONFLICT", func(r *transport.Report) { r.Decision.F_write_manifest = false }},
		{"after", "READBACK_SOURCE_CONFLICT", func(r *transport.Report) { r.After.F_read_error = "unavailable" }},
		{"next", "READBACK_SOURCE_CONFLICT", func(r *transport.Report) { r.Next.F_effect_id = "urn:other-effect" }},
		{"outcome", "RETAINED_ASSESSMENT_CONFLICT", func(r *transport.Report) { r.Outcome.F_command_failed = false }},
		{"missing-pair", "READBACK_PAIR_REQUIRED", func(r *transport.Report) { r.After = nil }},
	} {
		t.Run(row.name, func(t *testing.T) {
			var report transport.Report
			if err := json.Unmarshal(original, &report); err != nil {
				t.Fatal(err)
			}
			row.change(&report)
			raw, _ := json.Marshal(report)
			view, err := FromSavedReport(raw)
			if err != nil || view.F_status != "UNRESOLVED_ORIGINS" || view.F_reason != row.reason || len(view.F_lines) != 0 {
				t.Fatal("contradictory source records were projected", row.reason, err, view)
			}
		})
	}
}

func TestReportDecodingKeepsCanonicalFieldLocationsAndRawBytes(t *testing.T) {
	_, original, _ := runSavedReport(t, false, 201, false, false)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(original, &fields); err != nil {
		t.Fatal(err)
	}
	fields["SCOPE"] = json.RawMessage(`{"goal_id":"urn:case-folded-alias"}`)
	fields["future_extension"] = json.RawMessage(`{"value":true}`)
	raw, _ := json.Marshal(fields)
	view, err := FromSavedReport(raw)
	if err != nil || view.F_status != "RECORDED_REPORT_ORIGINS" || view.F_report_sha256 != contentDigest(raw) {
		t.Fatal("unknown aliases overwrote exact source locations", err, view)
	}
	other, err := FromSavedReport(append(bytes.Clone(raw), '\n'))
	if err != nil || other.F_report_sha256 == view.F_report_sha256 {
		t.Fatal("byte-distinct retained reports shared content identity", err)
	}
	delete(fields, "outcome")
	raw, _ = json.Marshal(fields)
	view, err = FromSavedReport(raw)
	if err != nil || view.F_reason != "REPORT_FIELDS_MISSING" || len(view.F_lines) != 0 {
		t.Fatal("missing retained assessment was invented", err, view)
	}
}

func TestReportViewRejectsUnreadableOrOversizedInput(t *testing.T) {
	for _, raw := range [][]byte{nil, []byte(`{"scope":{},"scope":{}}`), []byte(`{} {}`), {0xff}, bytes.Repeat([]byte(" "), int(ReportByteLimit()+1))} {
		if _, err := FromSavedReport(raw); err == nil {
			t.Fatal("ambiguous, invalid or out-of-profile report accepted")
		}
	}
}

func TestEarlierRetainedReportPreservesMissingLaterFields(t *testing.T) {
	raw, err := os.ReadFile("testdata/retained-readonly-report.json")
	if err != nil {
		t.Fatal(err)
	}
	before := contentDigest(raw)
	view, err := FromSavedReport(raw)
	if err != nil || view.F_status != "UNRESOLVED_ORIGINS" || view.F_reason != "REPORT_FIELDS_MISSING" || view.F_report_sha256 != before {
		t.Fatal("older bytes were silently completed", err, view)
	}
	stored, _ := os.ReadFile("testdata/retained-readonly-report.json")
	if !bytes.Equal(raw, stored) {
		t.Fatal("original report changed")
	}
}
