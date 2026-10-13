package bodycodegen

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// FillOriginRange uses byte offsets with an exclusive end. Source ranges refer
// to the origin's input file; decoded ranges refer to the computes body.
type FillOriginRange struct {
	Start int `json:"start_byte"`
	End   int `json:"end_byte"`
}

type SourceFillFragment struct {
	Relation         string           `json:"relation,omitempty"`
	Kind             string           `json:"kind"`
	HoleID           string           `json:"hole_id,omitempty"`
	Expression       string           `json:"expression,omitempty"`
	Grammar          string           `json:"grammar,omitempty"`
	TemplateDecoded  FillOriginRange  `json:"template_decoded"`
	SelectedDecoded  FillOriginRange  `json:"selected_decoded"`
	TemplateSource   FillOriginRange  `json:"template_source"`
	SelectedSource   FillOriginRange  `json:"selected_source"`
	HoleDeclaration  *FillOriginRange `json:"hole_declaration,omitempty"`
	ExpressionSource *FillOriginRange `json:"expression_source,omitempty"`
	GrammarSource    *FillOriginRange `json:"grammar_source,omitempty"`
}

// SourceFillOrigin connects the materialized body to the assembly input before
// its source_fill block is removed. It makes no original-workspace-file claim.
type SourceFillOrigin struct {
	Schema               string               `json:"schema"`
	Filename             string               `json:"filename"`
	View                 string               `json:"view"`
	SourceSHA256         string               `json:"source_sha256"`
	SelectedSourceSHA256 string               `json:"selected_source_sha256"`
	SelectedBodySHA256   string               `json:"selected_body_sha256"`
	ActivityID           string               `json:"activity_id"`
	ActivitySource       FillOriginRange      `json:"activity_source"`
	AssemblySource       FillOriginRange      `json:"assembly_source"`
	Intent               string               `json:"intent"`
	IntentSource         FillOriginRange      `json:"intent_source"`
	CandidateID          string               `json:"candidate_id"`
	CandidateSource      *FillOriginRange     `json:"candidate_source,omitempty"`
	PlanSHA256           string               `json:"plan_sha256"`
	Fragments            []SourceFillFragment `json:"fragments"`
	DataFlow             *BodyValueOrigins    `json:"data_flow,omitempty"`
}

type SourceFillConditionOrigin struct {
	Schema               string                 `json:"schema"`
	Filename             string                 `json:"filename"`
	View                 string                 `json:"view"`
	SourceSHA256         string                 `json:"source_sha256"`
	SelectedSourceSHA256 string                 `json:"selected_source_sha256"`
	SelectedBodySHA256   string                 `json:"selected_body_sha256"`
	ActivityID           string                 `json:"activity_id"`
	ActivitySource       FillOriginRange        `json:"activity_source"`
	AssemblySource       FillOriginRange        `json:"assembly_source"`
	Intent               string                 `json:"intent"`
	IntentSource         FillOriginRange        `json:"intent_source"`
	CandidateID          string                 `json:"candidate_id"`
	CandidateSource      *FillOriginRange       `json:"candidate_source,omitempty"`
	PlanSHA256           string                 `json:"plan_sha256"`
	Fragments            []SourceFillFragment   `json:"fragments"`
	ConditionDecoded     FillOriginRange        `json:"condition_decoded"`
	Input                *SourceFillInputOrigin `json:"input_source,omitempty"`
	DataFlow             *BodyValueOrigins      `json:"data_flow,omitempty"`
}

// ConditionOrigin includes direct fragments and, when recorded, bounded static
// value/control dependencies. Template ranges retain their original scope.
func (o SourceFillOrigin) ConditionOrigin(condition BodyConditionOrigin) *SourceFillConditionOrigin {
	if condition.ActivityID != o.ActivityID || condition.BodySHA256 != o.SelectedBodySHA256 {
		return nil
	}
	link := &SourceFillConditionOrigin{Schema: o.Schema, Filename: o.Filename, View: o.View,
		SourceSHA256: o.SourceSHA256, SelectedSourceSHA256: o.SelectedSourceSHA256,
		SelectedBodySHA256: o.SelectedBodySHA256, ActivityID: o.ActivityID,
		ActivitySource: o.ActivitySource, AssemblySource: o.AssemblySource,
		Intent: o.Intent, IntentSource: o.IntentSource, CandidateID: o.CandidateID,
		CandidateSource: o.CandidateSource, PlanSHA256: o.PlanSHA256,
		ConditionDecoded: FillOriginRange{condition.Start, condition.End}}
	link.Schema = "gooo/source-fill-condition-origin/v1"
	link.CandidateSource = cloneFillOriginRange(o.CandidateSource)
	link.Fragments = nil
	link.DataFlow = o.DataFlow.conditionSlice(link.ConditionDecoded, o.SelectedBodySHA256)
	for _, fragment := range o.Fragments {
		relation := fillFragmentRelation(fragment.SelectedDecoded, link.ConditionDecoded, link.DataFlow)
		if relation != "" {
			if link.DataFlow != nil {
				fragment.Relation = relation
			}
			fragment.HoleDeclaration = cloneFillOriginRange(fragment.HoleDeclaration)
			fragment.ExpressionSource = cloneFillOriginRange(fragment.ExpressionSource)
			fragment.GrammarSource = cloneFillOriginRange(fragment.GrammarSource)
			link.Fragments = append(link.Fragments, fragment)
		}
	}
	if len(link.Fragments) == 0 {
		return nil
	}
	return link
}

func cloneFillOriginRange(value *FillOriginRange) *FillOriginRange {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func sourceFillActivity(filename string, source []byte, name string) (*syntax.ActivityDecl, error) {
	file, diagnostics := ParseBodyFile(filename, source)
	if diagnostics.HasErrors() {
		return nil, diagnostics.Error()
	}
	for _, declaration := range file.Declarations {
		if activity, ok := declaration.(*syntax.ActivityDecl); ok && activity.Name == name {
			return activity, nil
		}
	}
	return nil, fmt.Errorf("source fill origin: activity %q is absent", name)
}

func fillSpan(span syntax.Span) FillOriginRange {
	return FillOriginRange{span.Start.Offset, span.End.Offset}
}

func retainSourceFillOrigin(filename string, source []byte, activity string, result *Result) error {
	input, err := sourceFillActivity(filename, source, activity)
	if err != nil || input.Assembly == nil || input.Assembly.Spec.FillPlan == nil || result.Report.BodyFill == nil {
		return fmt.Errorf("source fill origin requires its assembly: %w", err)
	}
	selected := []byte(result.GoooSource)
	output, err := sourceFillActivity(filename, selected, activity)
	if err != nil {
		return err
	}
	index, err := sourceFillLiteralOrigins(source, input, result.Report.BodyFill.SelectedCandidateID)
	if err != nil {
		return err
	}
	fragments, bodySHA, err := sourceFillFragments(source, selected, input, output, result.Report.BodyFill, index)
	if err != nil {
		return err
	}
	flow := InspectBodyValueOrigins(output.ValueProgram, len(output.Inputs))
	r := result.Report.BodyFill
	r.SourceOrigin = &SourceFillOrigin{Schema: "gooo/source-fill-origin/v1", Filename: filename,
		View: "assembly_input", SourceSHA256: digest(source), SelectedSourceSHA256: digest(selected),
		SelectedBodySHA256: bodySHA, ActivityID: result.Report.ActivityID,
		ActivitySource: fillSpan(input.Span), AssemblySource: fillSpan(input.Assembly.Span),
		Intent: r.Intent, IntentSource: index.intent, CandidateID: r.SelectedCandidateID,
		CandidateSource: index.candidate, PlanSHA256: r.IRPlanSHA256, Fragments: fragments, DataFlow: &flow}
	return nil
}
