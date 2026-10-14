package bodycodegen

import "fmt"

// SourceFillCandidateOrigins describes a declared assignment before selection.
// It carries no selected-body coordinates, type result or functional score.
type SourceFillCandidateOrigins struct {
	Schema          string                      `json:"schema"`
	Filename        string                      `json:"filename"`
	SourceSHA256    string                      `json:"source_sha256"`
	Activity        string                      `json:"activity"`
	ActivityID      string                      `json:"activity_id"`
	CandidateID     string                      `json:"candidate_id"`
	CandidateSource FillOriginRange             `json:"candidate_source"`
	Intent          string                      `json:"intent"`
	IntentSource    FillOriginRange             `json:"intent_source"`
	Status          string                      `json:"status"`
	UnresolvedRows  int                         `json:"unresolved_rows"`
	Rows            []SourceFillCandidateOrigin `json:"rows"`
	Scope           string                      `json:"scope"`
}

type SourceFillCandidateOrigin struct {
	HoleID           string           `json:"hole_id"`
	Expression       string           `json:"expression"`
	HoleDeclaration  FillOriginRange  `json:"hole_declaration"`
	ExpressionSource FillOriginRange  `json:"expression_source"`
	TemplateDecoded  *FillOriginRange `json:"template_decoded,omitempty"`
	TemplateSource   *FillOriginRange `json:"template_source,omitempty"`
	Status           string           `json:"status"`
	Reason           string           `json:"reason,omitempty"`
}

// InspectSourceFillCandidateOrigins reads one explicit source candidate, even
// when its expression would fail type checking. No candidate is materialized.
func InspectSourceFillCandidateOrigins(filename string, source []byte, activity, candidate string) (SourceFillCandidateOrigins, error) {
	r := SourceFillCandidateOrigins{Schema: "gooo/source-fill-candidate-origins/v1", Filename: filename,
		SourceSHA256: digest(source), Activity: activity, CandidateID: candidate, Status: "RESOLVED",
		Scope: "declared source assignment and original template token locations only; no selection, type checking, case evaluation or model call"}
	declaration, err := InspectSourceFillDeclarationOrigins(filename, source, activity, candidate)
	if err != nil {
		return r, err
	}
	if declaration.CandidateSource == nil || len(declaration.Holes) < 1 || len(declaration.Holes) > 8 {
		return r, fmt.Errorf("candidate origins require an explicit source candidate with 1..8 holes; generated grammar candidates have no declared assignment")
	}
	r.ActivityID, err = fillActivityID(filename, source, activity)
	if err != nil {
		return r, err
	}
	r.Intent, r.IntentSource, r.CandidateSource = declaration.Intent, declaration.IntentSource, *declaration.CandidateSource
	for _, hole := range declaration.Holes {
		row, err := sourceFillCandidateOriginRow(source, declaration, hole)
		if err != nil {
			return r, err
		}
		if row.Status != "MAPPED" {
			r.Status, r.UnresolvedRows = "UNRESOLVED", r.UnresolvedRows+1
		}
		r.Rows = append(r.Rows, row)
	}
	return r, nil
}

func sourceFillCandidateOriginRow(source []byte, declaration SourceFillDeclarationOrigins, hole SourceFillDeclarationHole) (SourceFillCandidateOrigin, error) {
	row := SourceFillCandidateOrigin{HoleID: hole.ID, Expression: hole.Expression,
		HoleDeclaration: hole.Declaration, Status: "MAPPED"}
	if hole.ExpressionSource == nil {
		return row, fmt.Errorf("explicit candidate hole %q has no declared expression", hole.ID)
	}
	row.ExpressionSource = *hole.ExpressionSource
	start, end, err := fillOriginIdentifier(declaration.TemplateBody, bodyFillHoleToken(hole.ID))
	if err != nil {
		row.Status, row.Reason = "UNRESOLVED", "TEMPLATE_HOLE_NOT_UNIQUE"
		return row, nil
	}
	literal := declaration.TemplateLiteral
	rawStart, rawEnd, err := SourceLiteralByteRange(string(source[literal.Start:literal.End]), declaration.TemplateBody, start, end)
	if err != nil {
		return row, err
	}
	row.TemplateDecoded = &FillOriginRange{Start: start, End: end}
	row.TemplateSource = &FillOriginRange{Start: literal.Start + rawStart, End: literal.Start + rawEnd}
	return row, nil
}
