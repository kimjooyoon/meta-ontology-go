package workspaceexecution

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func unmappedRecordChoiceLocation(input workspaceFillInput, reason string) bodyexecution.CompositionLocalFileLocation {
	return bodyexecution.CompositionLocalFileLocation{Status: "UNMAPPED", Reason: reason, View: "workspace_input_file",
		PackagePath: input.pkg, Filename: input.filename, ActivityID: input.activityID, SourceSHA256: input.fileSHA}
}

func mappedRecordChoiceLocation(input workspaceFillInput, span bodycodegen.FillOriginRange, expression string) bodyexecution.CompositionLocalFileLocation {
	result := unmappedRecordChoiceLocation(input, "CHOICE_RANGE_UNAVAILABLE")
	if input.reason == "" && span.Start >= 0 && span.End > span.Start && span.End <= len(input.source) {
		result.Status, result.Reason, result.Start, result.End, result.Expression = "MAPPED", "", span.Start, span.End, expression
	}
	return result
}

func recordChoiceTemplateLocation(input workspaceFillInput, templateSHA string, span bodycodegen.FillOriginRange) bodyexecution.CompositionLocalFileLocation {
	out := unmappedRecordChoiceLocation(input, "CHOICE_TEMPLATE_RANGE_UNAVAILABLE")
	if input.reason != "" || templateSHA != input.loweredBodySHA {
		return out
	}
	_, mapping, err := bodycodegen.NormalizeBodySource(input.loweredBody)
	if err != nil {
		return out
	}
	span = mapping.OriginalRange(span.Start, span.End)
	for _, edits := range [][]BodySourceEdit{input.recordEdits, input.callEdits} {
		var ok bool
		span, ok = undoFillSourceEdits(span, edits)
		if !ok {
			return out
		}
	}
	literal := input.bodyLiteral
	if literal.Start < 0 || literal.End > len(input.source) || span.Start < 0 || span.End > len(input.originalBody) || span.End <= span.Start {
		return out
	}
	start, end, err := bodycodegen.SourceLiteralByteRange(input.source[literal.Start:literal.End], input.originalBody, span.Start, span.End)
	if err != nil {
		return out
	}
	return mappedRecordChoiceLocation(input, bodycodegen.FillOriginRange{Start: literal.Start + start, End: literal.Start + end}, input.originalBody[span.Start:span.End])
}
