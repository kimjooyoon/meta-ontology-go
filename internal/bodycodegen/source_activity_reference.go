package bodycodegen

import (
	"context"
	"fmt"
	"go/ast"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type SourceActivityReference struct {
	ActivityID string `json:"activity_id"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Start      int    `json:"start_byte"`
	End        int    `json:"end_byte"`
}

// ResolveActivityReference follows a declaration or a free direct computes call.
// A matching spelling in prose, a local variable or an argument is not a target.
func ResolveActivityReference(ctx context.Context, filename string, source []byte, offset int) (SourceActivityReference, bool, error) {
	var empty SourceActivityReference
	if ctx == nil {
		ctx = context.Background()
	}
	if len(source) > 128<<10 || offset < 0 || offset > len(source) {
		return empty, false, fmt.Errorf("activity reference requires bounded source and a valid byte position")
	}
	if err := ctx.Err(); err != nil {
		return empty, false, err
	}
	file, diagnostics := ParseBodyFile(filename, source)
	if diagnostics.HasErrors() {
		return empty, false, diagnostics.Error()
	}
	var name, kind string
	var start, end int
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		if offset >= activity.NameSpan.Start.Offset && offset < activity.NameSpan.End.Offset {
			name, kind = activity.Name, "declaration"
			start, end = activity.NameSpan.Start.Offset, activity.NameSpan.End.Offset
			break
		}
		span := activity.ValueProgramSpan
		if !activity.ValueProgramPresent || offset <= span.Start.Offset || offset >= span.End.Offset-1 {
			continue
		}
		var err error
		name, start, end, err = bodyActivityReference(source, activity, offset)
		if err != nil {
			return empty, false, err
		}
		kind = "direct_call"
		break
	}
	if name == "" {
		return empty, false, nil
	}
	if len(file.Imports) != 0 {
		return empty, false, fmt.Errorf("activity reference currently requires one file without imports")
	}
	model, _, err := resolveBodyModel(file)
	if err != nil {
		return empty, false, err
	}
	if err := ctx.Err(); err != nil {
		return empty, false, err
	}
	for _, node := range model.Nodes {
		if node.Kind == bidir.ActivityKind && node.Name == name {
			return SourceActivityReference{ActivityID: string(node.ID), Name: name, Kind: kind, Start: start, End: end}, true, nil
		}
	}
	return empty, false, nil
}

func bodyActivityReference(source []byte, activity *syntax.ActivityDecl, offset int) (string, int, int, error) {
	span := activity.ValueProgramSpan
	raw := string(source[span.Start.Offset:span.End.Offset])
	position, err := sourceLiteralPosition(raw, activity.ValueProgram, offset-span.Start.Offset, true)
	if err != nil {
		return "", 0, 0, err
	}
	file, set, header, err := parseRenameBody(activity.ValueProgram, len(activity.Inputs), syntax.ActivityDeclaredInputNames(activity)...)
	if err != nil {
		return "", 0, 0, err
	}
	_, mapping, _ := rewriteBodySurface(activity.ValueProgram)
	var found *ast.Ident
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		id, ok := ast.Unparen(call.Fun).(*ast.Ident)
		if !ok || id.Obj != nil {
			return true
		}
		start := mapping.originalOffset(set.Position(id.Pos()).Offset - header)
		if position >= start && position < start+len(id.Name) {
			found = id
		}
		return true
	})
	if found == nil {
		return "", 0, 0, nil
	}
	start := mapping.originalOffset(set.Position(found.Pos()).Offset - header)
	first, err := sourceLiteralPosition(raw, activity.ValueProgram, start, false)
	if err != nil {
		return "", 0, 0, err
	}
	last, err := sourceLiteralPosition(raw, activity.ValueProgram, start+len(found.Name), false)
	return found.Name, span.Start.Offset + first, span.Start.Offset + last, err
}
