package bodycodegen

import (
	"fmt"
	"go/scanner"
	"go/token"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type fillOriginEdit struct {
	id, expression, replacement string
	start, end                  int
}

func sourceFillFragments(source, selected []byte, input, output *syntax.ActivityDecl,
	receipt *IRBodyFillReceipt, index fillOriginIndex) ([]SourceFillFragment, string, error) {
	body, err := rewriteLetDeclarations(input.ValueProgram)
	if err != nil {
		return nil, "", err
	}
	edits, err := prepareFillOriginEdits(body, input, receipt)
	if err != nil {
		return nil, "", err
	}
	fragments, assembled, err := assembleFillOriginFragments(body, edits, index)
	if err != nil {
		return nil, "", err
	}
	actual, err := rewriteLetDeclarations(output.ValueProgram)
	if err != nil || actual != assembled {
		return nil, "", fmt.Errorf("source fill origin selected body differs from its substitutions")
	}
	for i := range fragments {
		f := &fragments[i]
		f.TemplateSource, err = fillBodySourceRange(source, input, f.TemplateDecoded)
		if err != nil {
			return nil, "", err
		}
		f.SelectedSource, err = fillBodySourceRange(selected, output, f.SelectedDecoded)
		if err != nil {
			return nil, "", err
		}
		if f.Kind == "retained_template" {
			f.ExpressionSource = cloneFillOriginRange(&f.TemplateSource)
		}
	}
	return fragments, digest([]byte(actual)), nil
}

func prepareFillOriginEdits(body string, activity *syntax.ActivityDecl, receipt *IRBodyFillReceipt) ([]fillOriginEdit, error) {
	fills := make(map[string]string, len(receipt.HoleFills))
	for _, fill := range receipt.HoleFills {
		fills[fill.HoleID] = fill.Expression
	}
	var edits []fillOriginEdit
	for _, hole := range activity.Assembly.Spec.FillPlan.Holes {
		expression, ok := fills[hole.ID]
		if !ok {
			return nil, fmt.Errorf("source fill origin has no selected expression for %q", hole.ID)
		}
		name := bodyFillHoleToken(hole.ID)
		start, end, err := fillOriginIdentifier(body, name)
		if err != nil {
			return nil, err
		}
		replacement, err := replaceIdentifier(name, name, expression)
		if err != nil {
			return nil, err
		}
		edits = append(edits, fillOriginEdit{hole.ID, expression, replacement, start, end})
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	return edits, nil
}

func fillOriginIdentifier(body, name string) (int, int, error) {
	fset := token.NewFileSet()
	file := fset.AddFile("fill-origin", fset.Base(), len(body))
	var scan scanner.Scanner
	scan.Init(file, []byte(body), nil, scanner.ScanComments)
	start, count := -1, 0
	for {
		position, kind, literal := scan.Scan()
		if kind == token.EOF {
			break
		}
		if kind == token.IDENT && literal == name {
			start, count = file.Offset(position), count+1
		}
	}
	if count != 1 {
		return 0, 0, fmt.Errorf("source fill origin needs one occurrence of %q, got %d", name, count)
	}
	return start, start + len(name), nil
}

func assembleFillOriginFragments(body string, edits []fillOriginEdit, index fillOriginIndex) ([]SourceFillFragment, string, error) {
	var fragments []SourceFillFragment
	var out strings.Builder
	cursor := 0
	appendTemplate := func(end int) {
		if end <= cursor {
			return
		}
		start := out.Len()
		out.WriteString(body[cursor:end])
		fragments = append(fragments, SourceFillFragment{Kind: "retained_template",
			TemplateDecoded: FillOriginRange{cursor, end}, SelectedDecoded: FillOriginRange{start, out.Len()}})
	}
	for _, edit := range edits {
		if edit.start < cursor {
			return nil, "", fmt.Errorf("source fill origin edits overlap")
		}
		appendTemplate(edit.start)
		fragment, err := substitutedFillOrigin(edit, index)
		if err != nil {
			return nil, "", err
		}
		start := out.Len()
		out.WriteString(edit.replacement)
		fragment.TemplateDecoded, fragment.SelectedDecoded = FillOriginRange{edit.start, edit.end}, FillOriginRange{start, out.Len()}
		fragments = append(fragments, fragment)
		cursor = edit.end
	}
	appendTemplate(len(body))
	return fragments, out.String(), nil
}

func substitutedFillOrigin(edit fillOriginEdit, index fillOriginIndex) (SourceFillFragment, error) {
	fragment := SourceFillFragment{HoleID: edit.id, Expression: edit.expression}
	hole, ok := index.holes[edit.id]
	if !ok {
		return fragment, fmt.Errorf("source fill origin hole declaration missing: %s", edit.id)
	}
	fragment.HoleDeclaration = &hole
	if literal, ok := index.expressions[edit.id]; ok {
		if literal.value != edit.expression {
			return fragment, fmt.Errorf("source fill origin selected expression differs")
		}
		fragment.Kind, fragment.ExpressionSource = "declared_candidate", &literal.span
		return fragment, nil
	}
	grammar, ok := index.grammars[edit.id]
	if !ok {
		grammar, ok = index.grammars[""]
	}
	if !ok {
		return fragment, fmt.Errorf("source fill origin has neither candidate nor grammar for %s", edit.id)
	}
	fragment.Kind, fragment.Grammar, fragment.GrammarSource = "derived_grammar", grammar.value, &grammar.span
	return fragment, nil
}

func fillBodySourceRange(source []byte, activity *syntax.ActivityDecl, decoded FillOriginRange) (FillOriginRange, error) {
	span := activity.ValueProgramSpan
	if span.Start.Offset < 0 || span.End.Offset > len(source) || span.End.Offset <= span.Start.Offset {
		return FillOriginRange{}, fmt.Errorf("source fill origin body literal is outside source")
	}
	start, end, err := SourceLiteralByteRange(string(source[span.Start.Offset:span.End.Offset]), activity.ValueProgram, decoded.Start, decoded.End)
	return FillOriginRange{span.Start.Offset + start, span.Start.Offset + end}, err
}
