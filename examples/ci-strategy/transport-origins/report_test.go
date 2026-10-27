package transportorigins

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestProjectionKeepsItsCompiledOriginApartFromReportedMetadata(t *testing.T) {
	report, _, wire := runSavedReport(t, false, 201, false, false)
	requests := wire.requests
	own := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(GoooGenerationOrigin)))
	for _, reported := range []string{report.GenerationOrigin, "different reported metadata"} {
		report.GenerationOrigin = reported
		raw, err := json.Marshal(report)
		if err != nil {
			t.Fatal(err)
		}
		result, err := FromSavedReport(raw)
		if err != nil || result.F_status != "RECORDED_REPORT_ORIGINS" {
			t.Fatal("compiled projection lost its source context", err, result)
		}
		reportID := fmt.Sprintf("sha256:%x", sha256.Sum256(raw))
		projection := "<urn:gooo:transport-report:" + reportID + ":view:" + own + ">"
		text := strings.Join(result.F_lines, "\n")
		for _, required := range []string{
			projection + " <http://www.w3.org/ns/prov#specializationOf> <urn:gooo:transport-report:" + reportID + ":view> .",
			projection + " <gooo://transport-origins/projectionGenerationContext> <urn:gooo:projection-generation-context:" + own + "> .",
			fmt.Sprintf("<gooo://transport-origins/reportedGenerationContext> <urn:gooo:generation-context:sha256:%x> .", sha256.Sum256([]byte(reported))),
		} {
			if !strings.Contains(text, required) {
				t.Fatal("compiled and reported origins were mixed", required, text)
			}
		}
	}
	if wire.requests != requests {
		t.Fatal("origin projection dispatched work")
	}
}

func TestSavedReportUsesActualRuntimeRecordsAndNoFurtherRequests(t *testing.T) {
	for _, row := range []struct {
		name, state   string
		apply         bool
		status        int
		store, cancel bool
	}{
		{"read-only", "NO_WRITE_ATTEMPT", false, 201, false, false},
		{"stored", "EFFECT_OBSERVED", true, 201, true, false},
		{"failed-response-stored", "EFFECT_OBSERVED_WITH_WRITE_FAILURE", true, 500, true, false},
		{"failed-no-effect", "EFFECT_NOT_OBSERVED", true, 500, false, false},
		{"cancelled-readback", "UNRESOLVED_EFFECT", true, 201, true, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			report, raw, wire := runSavedReport(t, row.apply, row.status, row.store, row.cancel)
			if report.Outcome.F_state != row.state {
				t.Fatal("runtime fixture", report.Outcome)
			}
			original := bytes.Clone(raw)
			requests, puts := wire.requests, wire.puts
			view, err := FromSavedReport(raw)
			if err != nil || view.F_status != "RECORDED_REPORT_ORIGINS" {
				t.Fatal(err, view)
			}
			if wire.requests != requests || wire.puts != puts || !bytes.Equal(raw, original) {
				t.Fatal("view changed the retained source or dispatched work")
			}
			if view.F_report_sha256 != fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) {
				t.Fatal("raw bytes lost")
			}
			text := strings.Join(view.F_lines, "\n")
			if !strings.Contains(text, `<gooo://transport-origins/state> "`+row.state+`"`) {
				t.Fatal(text)
			}
			if strings.Contains(text, "#Activity>") || strings.Contains(text, "#wasGeneratedBy>") {
				t.Fatal("saved content invented a physical execution", text)
			}
			if !strings.Contains(text, `<gooo://transport-origins/executionIdentity> "UNOBSERVED"`) {
				t.Fatal(text)
			}
			if strings.Contains(text, `<gooo://transport-origins/recordPath> "/after"`) != (report.After != nil) {
				t.Fatal("readback record presence changed", text)
			}
			if !strings.Contains(text, `<gooo://transport-origins/status> "PROPOSED"`) {
				t.Fatal("proposal lost its status", text)
			}
		})
	}
}
