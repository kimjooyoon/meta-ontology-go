package bodycodegen

import "fmt"

// SourceFillInputOrigin retains the first package-file view alongside the
// assembly input. Mapping failures remain visible without changing execution.
type SourceFillInputOrigin struct {
	Schema          string                    `json:"schema"`
	Status          string                    `json:"status"`
	Reason          string                    `json:"reason,omitempty"`
	View            string                    `json:"view"`
	PackagePath     string                    `json:"package_path"`
	Filename        string                    `json:"filename"`
	SourceSHA256    string                    `json:"source_sha256"`
	ActivityID      string                    `json:"activity_id"`
	ActivitySource  FillOriginRange           `json:"activity_source"`
	AssemblySource  FillOriginRange           `json:"assembly_source"`
	Intent          string                    `json:"intent"`
	IntentSource    FillOriginRange           `json:"intent_source"`
	CandidateID     string                    `json:"candidate_id"`
	CandidateSource *FillOriginRange          `json:"candidate_source,omitempty"`
	Fragments       []SourceFillInputFragment `json:"fragments"`
}

type SourceFillInputFragment struct {
	Relation         string           `json:"relation,omitempty"`
	Status           string           `json:"status"`
	Reason           string           `json:"reason,omitempty"`
	Kind             string           `json:"kind"`
	HoleID           string           `json:"hole_id,omitempty"`
	Expression       string           `json:"expression,omitempty"`
	Grammar          string           `json:"grammar,omitempty"`
	SelectedDecoded  FillOriginRange  `json:"selected_decoded"`
	TemplateSource   *FillOriginRange `json:"template_source,omitempty"`
	HoleDeclaration  *FillOriginRange `json:"hole_declaration,omitempty"`
	ExpressionSource *FillOriginRange `json:"expression_source,omitempty"`
	GrammarSource    *FillOriginRange `json:"grammar_source,omitempty"`
}

// SourceFillDeclarationOrigins is a static syntax view. It does not select a
// candidate or generate expressions. TemplateBody is decoded literal text.
type SourceFillDeclarationOrigins struct {
	ActivitySource  FillOriginRange
	AssemblySource  FillOriginRange
	TemplateLiteral FillOriginRange
	TemplateBody    string
	Intent          string
	IntentSource    FillOriginRange
	CandidateSource *FillOriginRange
	Holes           []SourceFillDeclarationHole
}

type SourceFillDeclarationHole struct {
	ID               string
	Declaration      FillOriginRange
	Expression       string
	ExpressionSource *FillOriginRange
	Grammar          string
	GrammarSource    *FillOriginRange
}

func InspectSourceFillDeclarationOrigins(filename string, source []byte, activity, candidate string) (SourceFillDeclarationOrigins, error) {
	var out SourceFillDeclarationOrigins
	if len(source) > 128<<10 {
		return out, fmt.Errorf("source fill declaration source exceeds 128 KiB")
	}
	decl, err := sourceFillActivity(filename, source, activity)
	if err != nil {
		return out, err
	}
	if decl.Assembly == nil || decl.Assembly.Spec.FillPlan == nil {
		return out, fmt.Errorf("source fill declaration is absent")
	}
	index, err := sourceFillLiteralOrigins(source, decl, candidate)
	if err != nil {
		return out, err
	}
	plan := decl.Assembly.Spec.FillPlan
	if plan.Generation == nil && index.candidate == nil {
		return out, fmt.Errorf("source fill candidate is absent")
	}
	out = SourceFillDeclarationOrigins{ActivitySource: fillSpan(decl.Span), AssemblySource: fillSpan(decl.Assembly.Span),
		TemplateLiteral: fillSpan(decl.ValueProgramSpan), TemplateBody: decl.ValueProgram,
		Intent: plan.Intent, IntentSource: index.intent, CandidateSource: index.candidate}
	for _, hole := range plan.Holes {
		row := SourceFillDeclarationHole{ID: hole.ID, Declaration: index.holes[hole.ID]}
		if expression, found := index.expressions[hole.ID]; found {
			row.Expression, row.ExpressionSource = expression.value, cloneFillOriginRange(&expression.span)
		} else {
			grammar, found := index.grammars[hole.ID]
			if !found {
				grammar, found = index.grammars[""]
			}
			if !found {
				return out, fmt.Errorf("source fill hole has no expression or grammar")
			}
			row.Grammar, row.GrammarSource = grammar.value, cloneFillOriginRange(&grammar.span)
		}
		out.Holes = append(out.Holes, row)
	}
	return out, nil
}
