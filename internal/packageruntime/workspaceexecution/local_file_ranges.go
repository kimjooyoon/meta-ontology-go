package workspaceexecution

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func localFileLocation(input workspaceFillInput, found bool, selected *bodycodegen.BodyExpressionSource, fill localFillSelection) bodyexecution.CompositionLocalFileLocation {
	out := bodyexecution.CompositionLocalFileLocation{Status: "UNMAPPED", Reason: "INPUT_SOURCE_UNAVAILABLE", View: "workspace_input_file"}
	if !found {
		return out
	}
	out.PackagePath, out.Filename, out.ActivityID = input.pkg, input.filename, input.activityID
	out.SourceSHA256 = input.fileSHA
	if input.reason != "" {
		out.Reason = input.reason
		return out
	}
	if selected == nil || selected.Status != "MAPPED" {
		out.Reason = "SELECTED_RANGE_UNAVAILABLE"
		return out
	}
	span := bodycodegen.FillOriginRange{Start: selected.DecodedStart, End: selected.DecodedEnd}
	if input.loweredBodySHA == "" {
		out.Reason = "TEMPLATE_BODY_UNAVAILABLE"
		return out
	}
	if selected.BodySHA256 != input.loweredBodySHA {
		if fill.ambiguous {
			out.Reason = "AMBIGUOUS_ASSEMBLY_INPUT"
			return out
		}
		var ok bool
		span, ok = retainedLocalTemplateRange(span, fill.origin)
		if !ok {
			out.Reason = "ASSEMBLED_RANGE_HAS_NO_TEMPLATE_LOCATION"
			return out
		}
	}
	if span.Start < 0 || span.End <= span.Start || span.End > len(input.loweredBody) {
		out.Reason = "LOWERED_RANGE_DIFFERS"
		return out
	}
	if input.loweredBody[span.Start:span.End] != selected.Expression {
		out.Reason = "SELECTED_TEMPLATE_EXPRESSION_DIFFERS"
		return out
	}
	for _, edits := range [][]BodySourceEdit{input.recordEdits, input.callEdits} {
		var ok bool
		span, ok = undoFillSourceEdits(span, edits)
		if !ok {
			out.Reason = "LOWERING_BOUNDARY_CROSSES_REPLACED_NAME"
			return out
		}
	}
	literal := input.bodyLiteral
	if literal.Start < 0 || literal.End <= literal.Start || literal.End > len(input.source) || span.End > len(input.originalBody) {
		out.Reason = "INPUT_LITERAL_UNAVAILABLE"
		return out
	}
	start, end, err := bodycodegen.SourceLiteralByteRange(input.source[literal.Start:literal.End], input.originalBody, span.Start, span.End)
	if err != nil {
		out.Reason = "INPUT_LITERAL_RANGE_UNAVAILABLE"
		return out
	}
	out.Status, out.Reason, out.Start, out.End = "MAPPED", "", literal.Start+start, literal.Start+end
	out.Expression = input.originalBody[span.Start:span.End]
	return out
}

func retainedLocalTemplateRange(span bodycodegen.FillOriginRange, fill *bodycodegen.SourceFillOrigin) (bodycodegen.FillOriginRange, bool) {
	if fill == nil {
		return span, false
	}
	for _, fragment := range fill.Fragments {
		s, t := fragment.SelectedDecoded, fragment.TemplateDecoded
		if fragment.Kind == "retained_template" && span.Start >= s.Start && span.End <= s.End && s.End-s.Start == t.End-t.Start {
			return bodycodegen.FillOriginRange{Start: t.Start + span.Start - s.Start, End: t.Start + span.End - s.Start}, true
		}
	}
	return span, false
}

func localFileConstruction(input workspaceFillInput, found bool, selected *bodycodegen.BodyExpressionSource, fill localFillSelection) *bodycodegen.SourceFillInputOrigin {
	if !found || selected == nil || selected.Status != "MAPPED" || fill.origin == nil || fill.ambiguous {
		return nil
	}
	link := &bodycodegen.SourceFillConditionOrigin{CandidateID: fill.origin.CandidateID}
	for _, fragment := range fill.origin.Fragments {
		if fragment.SelectedDecoded.Start < selected.DecodedEnd && fragment.SelectedDecoded.End > selected.DecodedStart {
			fragment.Relation = "local_expression"
			link.Fragments = append(link.Fragments, fragment)
		}
	}
	if len(link.Fragments) == 0 {
		return nil
	}
	return workspaceFillInputOrigin(input, link)
}
