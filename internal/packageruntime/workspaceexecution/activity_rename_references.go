package workspaceexecution

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func (index workspaceRenameIndex) references(ctx context.Context) ([]renameReference, error) {
	keys := make([]string, 0, len(index.calls))
	for key := range index.calls {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var references []renameReference
	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		d := index.calls[key]
		var body []renameReference
		var err error
		if d.activity.Assembly != nil && d.activity.Assembly.Spec.FillPlan != nil {
			body, err = index.sourceFillReferences(ctx, key, d)
		} else {
			body, err = index.literalReferences(key, d, d.activity.ValueProgramSpan, d.body, "computes", true, false)
		}
		if err != nil {
			return nil, fmt.Errorf("activity %s: %w", key, err)
		}
		references = append(references, body...)
		assembly, err := index.assemblyReferences(key, d)
		if err != nil {
			return nil, fmt.Errorf("activity %s: %w", key, err)
		}
		references = append(references, assembly...)
	}
	bindings, err := index.bindingReferences()
	return append(references, bindings...), err
}

func (index workspaceRenameIndex) literalReferences(key string, d workspaceCallDeclaration,
	span syntax.Span, value, surface string, body, fragment bool) ([]renameReference, error) {
	if value == "" {
		return nil, nil
	}
	_, sites, err := index.calls.rewriteCalls(key, d, value, surface, body)
	if err != nil {
		return nil, err
	}
	raw := index.sources[d.ref.PackagePath][d.source]
	literal := raw[span.Start.Offset:span.End.Offset]
	var refs []renameReference
	for _, site := range sites {
		spelling := value[site.Start:site.End]
		if fragment {
			if err := checkRenameFragmentScope(d, spelling); err != nil {
				return nil, err
			}
		}
		start := site.Start
		if dot := strings.LastIndexByte(spelling, '.'); dot >= 0 {
			start += dot + 1
		}
		first, last, err := bodycodegen.SourceLiteralByteRange(literal, value, start, site.End)
		if err != nil {
			return nil, err
		}
		refs = append(refs, renameReference{d.ref.PackagePath, d.source, index.identities[key], index.identities[site.Callee], index.calls[site.Callee].ref.PackagePath,
			surface, span.Start.Offset + first, span.Start.Offset + last, start != site.Start, true})
	}
	return refs, nil
}

func checkRenameFragmentScope(d workspaceCallDeclaration, spelling string) error {
	name, _, _ := strings.Cut(spelling, ".")
	refs, err := bodycodegen.InspectBodySourceReferences("return "+name+"()", len(d.activity.Inputs),
		d.body, d.activity.Assembly.Spec.Baseline)
	if err != nil {
		return err
	}
	if len(refs) != 1 || !refs[0].Direct || refs[0].AmbiguousScope {
		return fmt.Errorf("alternative call %q needs its precise enclosing lexical scope", spelling)
	}
	return nil
}

func (index workspaceRenameIndex) assemblyReferences(key string, d workspaceCallDeclaration) ([]renameReference, error) {
	assembly := d.activity.Assembly
	if assembly == nil {
		return nil, nil
	}
	if assembly.Spec.Search != nil {
		return nil, fmt.Errorf("search rename requires a candidate-scope map")
	}
	raw := index.sources[d.ref.PackagePath][d.source]
	start, end := assembly.Span.Start.Offset, assembly.Span.End.Offset
	tokens, diagnostics := syntax.Lex(raw[start:end])
	if diagnostics.HasErrors() {
		return nil, diagnostics.Error()
	}
	var refs []renameReference
	choice := -1
	for i, token := range tokens {
		if token.Kind == syntax.TokenIdentifier && token.Value == "choice" {
			choice++
		}
		if i == 0 || token.Kind != syntax.TokenString {
			continue
		}
		span := token.Span
		span.Start.Offset, span.End.Offset = span.Start.Offset+start, span.End.Offset+start
		current, err := index.assemblyTokenReferences(key, d, tokens[i-1], token, span, choice)
		if err != nil {
			return nil, err
		}
		refs = append(refs, current...)
	}
	return refs, nil
}
