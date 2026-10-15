package bodycodegen

import (
	"encoding/json"
	"fmt"
	"strings"
)

type SourceFillDerivedInput struct {
	SourceFillCandidateInput
	OriginKind   string `json:"origin_kind"`
	Grammar      string `json:"grammar"`
	GrammarStart int    `json:"grammar_start"`
	GrammarEnd   int    `json:"grammar_end"`
}

// Derived inputs retain zero direct-expression coordinates alongside an
// explicit origin kind and real grammar coordinates. Gooo owns interpretation.
func SourceFillDerivedOriginInputs(report SourceFillDerivedOrigins, target string) ([]byte, error) {
	pkg, activity, ok := strings.Cut(target, ":")
	if !ok || pkg == "" || activity == "" || strings.ContainsAny(target, " \t\r\n") || strings.Contains(activity, ":") {
		return nil, fmt.Errorf("consumer input needs package/path:Activity")
	}
	if report.Schema != "gooo/source-fill-derived-origins/v1" || len(report.Rows) < 1 || len(report.Rows) > 8 {
		return nil, fmt.Errorf("derived origin input requires 1..8 grammar hole rows")
	}
	out := struct {
		Schema string                              `json:"schema"`
		Inputs []map[string]SourceFillDerivedInput `json:"inputs"`
	}{Schema: "gooo/body-composition-inputs/v1"}
	for _, row := range report.Rows {
		input := SourceFillDerivedInput{
			Filename: report.Filename, SourceSHA256: report.SourceSHA256, ActivityID: report.ActivityID,
			CandidateID: report.CandidateID, HoleID: row.HoleID, Expression: row.Expression,
			Status: row.Status, Reason: row.Reason, HoleStart: row.HoleDeclaration.Start, HoleEnd: row.HoleDeclaration.End,
			OriginKind: "derived_grammar", Grammar: row.Grammar, GrammarStart: row.GrammarSource.Start, GrammarEnd: row.GrammarSource.End}
		if row.TemplateSource != nil {
			input.HasUse, input.UseStart, input.UseEnd = true, row.TemplateSource.Start, row.TemplateSource.End
		}
		out.Inputs = append(out.Inputs, map[string]SourceFillDerivedInput{target: input})
	}
	raw, err := json.Marshal(out)
	if err == nil && len(raw) > 32<<10 {
		return nil, fmt.Errorf("derived origin inputs exceed 32 KiB")
	}
	return raw, err
}
