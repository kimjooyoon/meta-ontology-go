package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestRunBodyCodegenUsesSourceDerivedIRFillWithDeterministicFallback(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-derived.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen([]string{"--json", "--activity", "Lift", "derived.gooo"},
		mapSourceReader{"derived.gooo": source}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("source-derived IR fill failed: code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var report struct {
		GoooSource string `json:"gooo_source"`
		Report     struct {
			BodyFill struct {
				CandidateGeneration struct {
					AssignmentsRetained int     `json:"assignments_retained"`
					Coverage            float64 `json:"assignment_coverage_percent"`
				} `json:"candidate_generation"`
			} `json:"body_fill"`
			Completeness struct {
				Dimensions []struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				} `json:"dimensions"`
			} `json:"completeness_receipt"`
		} `json:"report"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Report.BodyFill.CandidateGeneration.AssignmentsRetained != 16 || report.Report.BodyFill.CandidateGeneration.Coverage <= 0 {
		t.Fatalf("CLI omitted Gooo's bounded candidate-space receipt: %+v", report.Report.BodyFill.CandidateGeneration)
	}
	foundProgress := false
	for _, dimension := range report.Report.Completeness.Dimensions {
		if dimension.ID == "body_fill_assignment_space_coverage" && dimension.Status == "PROGRESS" {
			foundProgress = true
		}
	}
	if !foundProgress {
		t.Fatalf("truncated assignment space was not visible as incomplete: %+v", report.Report.Completeness.Dimensions)
	}
	if strings.Contains(report.GoooSource, "derive grammar") || strings.Contains(report.GoooSource, "__GOOO_BODY_HOLE_") {
		t.Fatalf("CLI output is not a directly reusable completed Gooo program:\n%s", report.GoooSource)
	}
}
