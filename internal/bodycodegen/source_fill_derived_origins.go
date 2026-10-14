package bodycodegen

import (
	"context"
	"fmt"
)

// SourceFillDerivedOrigins binds an enumerated assignment to its grammar. The
// generated expression has no direct literal or candidate declaration location.
type SourceFillDerivedOrigins struct {
	Schema         string                                `json:"schema"`
	Filename       string                                `json:"filename"`
	SourceSHA256   string                                `json:"source_sha256"`
	Activity       string                                `json:"activity"`
	ActivityID     string                                `json:"activity_id"`
	CandidateID    string                                `json:"candidate_id"`
	PlanSHA256     string                                `json:"plan_sha256"`
	Intent         string                                `json:"intent"`
	IntentSource   FillOriginRange                       `json:"intent_source"`
	Status         string                                `json:"status"`
	UnresolvedRows int                                   `json:"unresolved_rows"`
	Rows           []SourceFillDerivedOrigin             `json:"rows"`
	Generation     *IRBodyFillCandidateGenerationReceipt `json:"candidate_generation"`
	Scope          string                                `json:"scope"`
}

type SourceFillDerivedOrigin struct {
	HoleID          string           `json:"hole_id"`
	Expression      string           `json:"expression"`
	Grammar         string           `json:"grammar"`
	GrammarSource   FillOriginRange  `json:"grammar_source"`
	HoleDeclaration FillOriginRange  `json:"hole_declaration"`
	TemplateDecoded *FillOriginRange `json:"template_decoded,omitempty"`
	TemplateSource  *FillOriginRange `json:"template_source,omitempty"`
	Status          string           `json:"status"`
	Reason          string           `json:"reason,omitempty"`
}

func InspectSourceFillDerivedOrigins(ctx context.Context, filename string, source []byte, activity, candidateID string) (SourceFillDerivedOrigins, error) {
	r := SourceFillDerivedOrigins{Schema: "gooo/source-fill-derived-origins/v1", Filename: filename,
		SourceSHA256: digest(source), Activity: activity, CandidateID: candidateID, Status: "RESOLVED",
		Scope: "retained finite grammar assignment, grammar declaration and template token locations; no direct expression literal, candidate selection, body materialization, case evaluation or model call"}
	_, set, err := sourceFillCandidatePlan(ctx, filename, source, activity)
	if err != nil {
		return r, err
	}
	if set.Generation == nil {
		return r, fmt.Errorf("derived origins require a source-declared grammar")
	}
	candidate, found := candidateByID(set.Candidates, candidateID)
	if !found {
		return r, fmt.Errorf("derived candidate is absent from the retained source assignment set")
	}
	declaration, err := InspectSourceFillDeclarationOrigins(filename, source, activity, candidateID)
	if err != nil {
		return r, err
	}
	r.ActivityID, err = fillActivityID(filename, source, activity)
	if err != nil {
		return r, err
	}
	r.PlanSHA256, r.Generation = set.PlanSHA256, set.Generation
	r.Intent, r.IntentSource = declaration.Intent, declaration.IntentSource
	for _, hole := range declaration.Holes {
		row, err := sourceFillDerivedOriginRow(source, declaration, hole, candidate.Fills[hole.ID])
		if err != nil {
			return r, err
		}
		if row.Status != "MAPPED" {
			r.Status, r.UnresolvedRows = "UNRESOLVED", r.UnresolvedRows+1
		}
		r.Rows = append(r.Rows, row)
	}
	return r, ctx.Err()
}

func sourceFillDerivedOriginRow(source []byte, declaration SourceFillDeclarationOrigins, hole SourceFillDeclarationHole, expression string) (SourceFillDerivedOrigin, error) {
	row := SourceFillDerivedOrigin{HoleID: hole.ID, Expression: expression,
		Grammar: hole.Grammar, HoleDeclaration: hole.Declaration, Status: "MAPPED"}
	if expression == "" || hole.GrammarSource == nil || hole.ExpressionSource != nil {
		return row, fmt.Errorf("derived hole %q requires its grammar and enumerated expression", hole.ID)
	}
	row.GrammarSource = *hole.GrammarSource
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
