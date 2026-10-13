package workspaceexecution

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func bindWorkspaceFillInputSources(program Program, runtime *bodyexecution.CompositionRuntime) {
	for i := range runtime.ConditionSites {
		site := &runtime.ConditionSites[i]
		if site.Source == nil || site.Source.Construction == nil {
			continue
		}
		link := site.Source.Construction
		input, found := program.fillInputs[site.ActivityID]
		if !found {
			continue
		}
		link.Input = workspaceFillInputOrigin(input, link)
	}
}

func workspaceFillInputOrigin(input workspaceFillInput, link *bodycodegen.SourceFillConditionOrigin) *bodycodegen.SourceFillInputOrigin {
	out := &bodycodegen.SourceFillInputOrigin{Schema: "gooo/source-fill-input-origin/v1",
		Status: "UNMAPPED", Reason: "INPUT_DECLARATION_UNAVAILABLE", View: "workspace_input_file",
		PackagePath: input.pkg, Filename: input.filename, SourceSHA256: sourceSHA256([]byte(input.source)),
		ActivityID: input.activityID, CandidateID: link.CandidateID}
	decl, err := bodycodegen.InspectSourceFillDeclarationOrigins(input.filename, []byte(input.source), input.activity, link.CandidateID)
	if err != nil {
		return out
	}
	out.Status, out.Reason = "MAPPED", ""
	out.ActivitySource, out.AssemblySource = decl.ActivitySource, decl.AssemblySource
	out.Intent, out.IntentSource, out.CandidateSource = decl.Intent, decl.IntentSource, decl.CandidateSource
	for _, fragment := range link.Fragments {
		row := workspaceFillInputFragment(input, decl, fragment)
		out.Fragments = append(out.Fragments, row)
		if row.Status != "MAPPED" {
			out.Status, out.Reason = "PARTIAL", "FRAGMENT_ORIGIN_UNAVAILABLE"
		}
	}
	return out
}

func workspaceFillInputFragment(input workspaceFillInput, decl bodycodegen.SourceFillDeclarationOrigins,
	fragment bodycodegen.SourceFillFragment) bodycodegen.SourceFillInputFragment {
	row := bodycodegen.SourceFillInputFragment{Status: "UNMAPPED", Reason: "FRAGMENT_DECLARATION_UNAVAILABLE",
		Kind: fragment.Kind, HoleID: fragment.HoleID, SelectedDecoded: fragment.SelectedDecoded}
	row.TemplateSource, _ = originalFillTemplateRange(input, decl, fragment.TemplateDecoded)
	if fragment.Kind == "retained_template" {
		if row.TemplateSource != nil {
			row.Status, row.Reason = "MAPPED", ""
		}
		return row
	}
	for _, hole := range decl.Holes {
		if hole.ID != fragment.HoleID {
			continue
		}
		declaration := hole.Declaration
		row.HoleDeclaration = &declaration
		switch fragment.Kind {
		case "declared_candidate":
			row.Expression, row.ExpressionSource = hole.Expression, hole.ExpressionSource
			if row.ExpressionSource != nil {
				row.Status, row.Reason = "MAPPED", ""
			}
		case "derived_grammar":
			row.Grammar, row.GrammarSource = hole.Grammar, hole.GrammarSource
			if row.GrammarSource != nil {
				row.Status, row.Reason = "MAPPED", ""
			}
		}
		break
	}
	return row
}
