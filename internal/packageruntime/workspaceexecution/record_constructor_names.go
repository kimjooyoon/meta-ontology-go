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

func (n workspaceRecordNames) rewriteConstructors(pkg, source string, body bool) (string, error) {
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
		name, ok := literal.Type.(*ast.Ident)
		if !ok || !n.known[name.Name] {
			return true
		}
		lowered, visible := n.visible[pkg][name.Name]
		if !visible || lowered == "" {
			err = fmt.Errorf("record constructor %q is undeclared or ambiguous in package %q", name.Name, pkg)
			return false
		}
		if lowered != name.Name {
			start := set.Position(name.Pos()).Offset - len(prefix)
			changes = append(changes, replacement{start, start + len(name.Name), lowered})
		}
		return true
	})
	if err != nil {
		return "", err
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].start > changes[j].start })
	for _, change := range changes {
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

func (n workspaceRecordNames) rewriteAssembly(pkg string, assembly *syntax.AssemblyDecl) error {
	if assembly == nil {
		return nil
	}
	spec := &assembly.Spec
	var err error
	spec.Baseline, err = n.rewriteConstructors(pkg, spec.Baseline, true)
	if err != nil {
		return err
	}
	for i := range spec.Choices {
		choice := &spec.Choices[i]
		if choice.Kind == "field_value" || choice.Kind == "field_update" {
			choice.Alternative, err = n.rewriteConstructors(pkg, choice.Alternative, false)
			if err != nil {
				return err
			}
		}
	}
	if spec.FillPlan != nil {
		for i := range spec.FillPlan.Candidates {
			for j := range spec.FillPlan.Candidates[i].Fills {
				fill := &spec.FillPlan.Candidates[i].Fills[j]
				fill.Expression, err = n.rewriteConstructors(pkg, fill.Expression, false)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (n workspaceRecordNames) rewriteFillPlan(pkg string, plan bodycodegen.IRBodyFillPlan) (bodycodegen.IRBodyFillPlan, error) {
	plan.Candidates = append([]bodycodegen.IRBodyFillCandidate(nil), plan.Candidates...)
	for i := range plan.Candidates {
		candidate := &plan.Candidates[i]
		var err error
		candidate.Expression, err = n.rewriteConstructors(pkg, candidate.Expression, false)
		if err != nil {
			return plan, err
		}
		fills := make(map[string]string, len(candidate.Fills))
		for hole, expression := range candidate.Fills {
			fills[hole], err = n.rewriteConstructors(pkg, expression, false)
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
