package workspaceexecution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func (n workspaceRecordNames) rewriteConstructors(pkg, source string, body bool, scopes ...recordConstructorScope) (string, error) {
	return n.rewriteConstructorsWithEdits(pkg, source, body, scopes, nil)
}

func (n workspaceRecordNames) rewriteConstructorsWithEdits(pkg, source string, body bool, scopes []recordConstructorScope, edits *[]BodySourceEdit) (string, error) {
	if source == "" || len(n.canonical) == 0 {
		return source, nil
	}
	prefix, parsed := "", source
	if body {
		prefix = "package scoped\nfunc body() {\n"
		parsed = prefix + recordBodyGoSyntax(source) + "\n}"
	}
	set := token.NewFileSet()
	var root ast.Node
	var err error
	if body {
		root, err = parser.ParseFile(set, "computes", parsed, parser.ParseComments)
	} else {
		root, err = parser.ParseExprFrom(set, "expression", parsed, parser.ParseComments)
	}
	if err != nil {
		return "", fmt.Errorf("parse scoped record constructor: %w", err)
	}
	type replacement struct {
		start, end int
		name       string
	}
	var changes []replacement
	ast.Inspect(root, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok || err != nil {
			return err == nil
		}
		var lowered string
		lowered, err = n.constructorTypeName(pkg, literal.Type, scopes)
		if err != nil {
			return false
		}
		if lowered != "" {
			start := set.Position(literal.Type.Pos()).Offset - len(prefix)
			end := set.Position(literal.Type.End()).Offset - len(prefix)
			if source[start:end] != lowered {
				changes = append(changes, replacement{start, end, lowered})
			}
		}
		return true
	})
	if err != nil {
		return "", err
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].start > changes[j].start })
	for _, change := range changes {
		if edits != nil {
			*edits = append(*edits, BodySourceEdit{change.start, change.end, change.name})
		}
		source = source[:change.start] + change.name + source[change.end:]
	}
	return source, nil
}

// let and var have the same byte length, preserving original source offsets.
func recordBodyGoSyntax(body string) string {
	raw := []byte(body)
	file := token.NewFileSet().AddFile("computes", -1, len(raw))
	var scan scanner.Scanner
	scan.Init(file, raw, nil, scanner.ScanComments)
	for {
		pos, kind, literal := scan.Scan()
		if kind == token.EOF {
			break
		}
		if kind == token.IDENT && literal == "let" {
			copy(raw[file.Offset(pos):], "var")
		}
	}
	return string(raw)
}

func (n workspaceRecordNames) rewriteAssembly(pkg string, assembly *syntax.AssemblyDecl, scopes ...recordConstructorScope) error {
	if assembly == nil {
		return nil
	}
	spec := &assembly.Spec
	var err error
	spec.Baseline, err = n.rewriteConstructors(pkg, spec.Baseline, true, scopes...)
	if err != nil {
		return err
	}
	for i := range spec.Choices {
		choice := &spec.Choices[i]
		if choice.Kind == "field_value" || choice.Kind == "field_update" {
			choice.Alternative, err = n.rewriteConstructors(pkg, choice.Alternative, false, scopes...)
			if err != nil {
				return err
			}
		}
	}
	if spec.FillPlan != nil {
		for i := range spec.FillPlan.Candidates {
			for j := range spec.FillPlan.Candidates[i].Fills {
				fill := &spec.FillPlan.Candidates[i].Fills[j]
				fill.Expression, err = n.rewriteConstructorFill(pkg, fill.HoleID, fill.Expression, scopes)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (n workspaceRecordNames) rewriteFillPlan(pkg string, plan bodycodegen.IRBodyFillPlan, scopes ...recordConstructorScope) (bodycodegen.IRBodyFillPlan, error) {
	plan.Candidates = append([]bodycodegen.IRBodyFillCandidate(nil), plan.Candidates...)
	for i := range plan.Candidates {
		candidate := &plan.Candidates[i]
		var err error
		candidate.Expression, err = n.rewriteConstructorFill(pkg, plan.HoleID, candidate.Expression, scopes)
		if err != nil {
			return plan, err
		}
		fills := make(map[string]string, len(candidate.Fills))
		for hole, expression := range candidate.Fills {
			fills[hole], err = n.rewriteConstructorFill(pkg, hole, expression, scopes)
			if err != nil {
				return plan, err
			}
		}
		if candidate.Fills != nil {
			candidate.Fills = fills
		}
	}
	return plan, nil
}
