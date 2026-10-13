package workspaceexecution

import (
	"context"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type renameFillLiteral struct {
	span  syntax.Span
	value string
}

// sourceFillReferences inspects complete, source-declared bodies. Neither a
// candidate score nor a model result is needed to resolve a hole's lexical scope.
func (index workspaceRenameIndex) sourceFillReferences(ctx context.Context, key string,
	d workspaceCallDeclaration) ([]renameReference, error) {
	plan := d.activity.Assembly.Spec.FillPlan
	if plan.Generation != nil {
		return nil, fmt.Errorf("derived source_fill rename requires a generated candidate-scope map")
	}
	literals, err := sourceFillLiteralSpans(index.sources[d.ref.PackagePath][d.source], d.activity.Assembly)
	if err != nil {
		return nil, err
	}
	var refs []renameReference
	seen := map[renameReference]bool{}
	for _, candidate := range plan.Candidates {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		body, parts, err := materializeRenameFill(d.body, d.activity.ValueProgramSpan, candidate, literals[candidate.ID])
		if err != nil {
			return nil, fmt.Errorf("source_fill candidate %q: %w", candidate.ID, err)
		}
		_, sites, err := index.calls.rewriteCalls(key, d, body, "source_fill:"+candidate.ID, true)
		if err != nil {
			return nil, fmt.Errorf("source_fill candidate %q: %w", candidate.ID, err)
		}
		for _, site := range sites {
			ref, err := index.mappedFillReference(key, d, body, site, parts)
			if err != nil {
				return nil, err
			}
			if !seen[ref] {
				refs, seen[ref] = append(refs, ref), true
			}
		}
	}
	return refs, nil
}

func (index workspaceRenameIndex) mappedFillReference(key string, d workspaceCallDeclaration,
	body string, site WorkspaceCallSite, parts []renameFillPart) (renameReference, error) {
	start := site.Start
	spelling := body[start:site.End]
	if dot := strings.LastIndexByte(spelling, '.'); dot >= 0 {
		start += dot + 1
	}
	for _, part := range parts {
		if start < part.start || site.End > part.end {
			continue
		}
		raw := index.sources[d.ref.PackagePath][d.source][part.literal.span.Start.Offset:part.literal.span.End.Offset]
		first, last, err := bodycodegen.SourceLiteralByteRange(raw, part.literal.value,
			part.decodedStart+start-part.start, part.decodedStart+site.End-part.start)
		if err != nil {
			return renameReference{}, err
		}
		return renameReference{d.ref.PackagePath, d.source, index.identities[key], index.identities[site.Callee],
			index.calls[site.Callee].ref.PackagePath, part.surface, part.literal.span.Start.Offset + first,
			part.literal.span.Start.Offset + last, start != site.Start, true}, nil
	}
	return renameReference{}, fmt.Errorf("source_fill call has no single original literal range")
}

func sourceFillLiteralSpans(source string, assembly *syntax.AssemblyDecl) (map[string]map[string]renameFillLiteral, error) {
	start, end := assembly.Span.Start.Offset, assembly.Span.End.Offset
	tokens, diagnostics := syntax.Lex(source[start:end])
	if diagnostics.HasErrors() {
		return nil, diagnostics.Error()
	}
	literals := map[string]map[string]renameFillLiteral{}
	candidate := ""
	for i, item := range tokens {
		if item.Kind != syntax.TokenIdentifier {
			continue
		}
		if item.Value == "candidate" && i+1 < len(tokens) && tokens[i+1].Kind == syntax.TokenString {
			candidate = tokens[i+1].Value
			literals[candidate] = map[string]renameFillLiteral{}
		}
		if item.Value != "fill" || i+2 >= len(tokens) || tokens[i+1].Kind != syntax.TokenString ||
			tokens[i+2].Kind != syntax.TokenString || candidate == "" {
			continue
		}
		expression := tokens[i+2]
		span := expression.Span
		span.Start.Offset, span.End.Offset = start+span.Start.Offset, start+span.End.Offset
		literals[candidate][tokens[i+1].Value] = renameFillLiteral{span, expression.Value}
	}
	return literals, nil
}
