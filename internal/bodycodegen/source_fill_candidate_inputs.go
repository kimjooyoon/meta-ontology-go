package bodycodegen

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SourceFillCandidateInput transports primitive facts. Gooo owns interpretation.
type SourceFillCandidateInput struct {
	Filename        string `json:"filename"`
	SourceSHA256    string `json:"source_sha256"`
	ActivityID      string `json:"activity_id"`
	CandidateID     string `json:"candidate_id"`
	HoleID          string `json:"hole_id"`
	Expression      string `json:"expression"`
	Status          string `json:"status"`
	Reason          string `json:"reason"`
	HasUse          bool   `json:"has_use"`
	HoleStart       int    `json:"hole_start"`
	HoleEnd         int    `json:"hole_end"`
	UseStart        int    `json:"use_start"`
	UseEnd          int    `json:"use_end"`
	ExpressionStart int    `json:"expression_start"`
	ExpressionEnd   int    `json:"expression_end"`
}

func SourceFillCandidateOriginInputs(report SourceFillCandidateOrigins, target string) ([]byte, error) {
	pkg, activity, ok := strings.Cut(target, ":")
	if !ok || pkg == "" || activity == "" || strings.ContainsAny(target, " \t\r\n") || strings.Contains(activity, ":") {
		return nil, fmt.Errorf("consumer input needs package/path:Activity")
	}
	if report.Schema != "gooo/source-fill-candidate-origins/v1" || len(report.Rows) < 1 || len(report.Rows) > 8 {
		return nil, fmt.Errorf("candidate origin input requires 1..8 declared hole rows")
	}
	out := struct {
		Schema string                                `json:"schema"`
		Inputs []map[string]SourceFillCandidateInput `json:"inputs"`
	}{Schema: "gooo/body-composition-inputs/v1"}
	for _, row := range report.Rows {
		input := SourceFillCandidateInput{Filename: report.Filename, SourceSHA256: report.SourceSHA256,
			ActivityID: report.ActivityID, CandidateID: report.CandidateID, HoleID: row.HoleID,
			Expression: row.Expression, Status: row.Status, Reason: row.Reason,
			HoleStart: row.HoleDeclaration.Start, HoleEnd: row.HoleDeclaration.End,
			ExpressionStart: row.ExpressionSource.Start, ExpressionEnd: row.ExpressionSource.End}
		if row.TemplateSource != nil {
			input.HasUse, input.UseStart, input.UseEnd = true, row.TemplateSource.Start, row.TemplateSource.End
		}
		out.Inputs = append(out.Inputs, map[string]SourceFillCandidateInput{target: input})
	}
	raw, err := json.Marshal(out)
	if err == nil && len(raw) > 32<<10 {
		return nil, fmt.Errorf("candidate origin inputs exceed 32 KiB")
	}
	return raw, err
}
