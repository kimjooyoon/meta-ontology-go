package workspaceexecution

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// Edits use coordinates before each rewrite stage. Reverse constructor edits
// first, then call edits. A boundary inside a replaced name remains unresolved.
func originalFillTemplateRange(input workspaceFillInput, declarations bodycodegen.SourceFillDeclarationOrigins,
	span bodycodegen.FillOriginRange) (*bodycodegen.FillOriginRange, bool) {
	if input.reason != "" || span.Start < 0 || span.End < span.Start || span.End > len(input.loweredBody) {
		return nil, false
	}
	var ok bool
	for _, stage := range [][]BodySourceEdit{input.recordEdits, input.callEdits} {
		span, ok = undoFillSourceEdits(span, stage)
		if !ok {
			return nil, false
		}
	}
	literal := declarations.TemplateLiteral
	if literal.Start < 0 || literal.End <= literal.Start || literal.End > len(input.source) {
		return nil, false
	}
	start, end, err := bodycodegen.SourceLiteralByteRange(input.source[literal.Start:literal.End], declarations.TemplateBody, span.Start, span.End)
	if err != nil {
		return nil, false
	}
	return &bodycodegen.FillOriginRange{Start: literal.Start + start, End: literal.Start + end}, true
}

func undoFillSourceEdits(span bodycodegen.FillOriginRange, edits []BodySourceEdit) (bodycodegen.FillOriginRange, bool) {
	ordered := append([]BodySourceEdit(nil), edits...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Start < ordered[j].Start })
	translate := func(offset int) (int, bool) {
		delta, end := 0, 0
		for _, edit := range ordered {
			if edit.Start < end || edit.End < edit.Start {
				return 0, false
			}
			first, last := edit.Start+delta, edit.Start+delta+len(edit.Replacement)
			if offset <= first {
				return offset - delta, true
			}
			if offset < last {
				return 0, false
			}
			delta += len(edit.Replacement) - (edit.End - edit.Start)
			end = edit.End
		}
		return offset - delta, true
	}
	start, a := translate(span.Start)
	end, b := translate(span.End)
	return bodycodegen.FillOriginRange{Start: start, End: end}, a && b && start >= 0 && end >= start
}
