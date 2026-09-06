package syntaxregistration

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"strconv"
	"strings"
)

type boundary struct {
	ID            string `json:"id"`
	Schema        string `json:"schema"`
	MetaOperation string `json:"meta_operation"`
	Target        int    `json:"target"`
	LinkTarget    int    `json:"link_target"`
}

type denominator struct {
	Schema        string     `json:"schema"`
	DenominatorID string     `json:"denominator_id"`
	Version       int        `json:"version"`
	Boundaries    []boundary `json:"boundaries"`
}

func generateDenominator(raw []byte, version, capability int) ([]byte, error) {
	var value denominator
	if err := decodeStrict(raw, &value); err != nil {
		return nil, err
	}
	if value.Schema != "gooo/vertical-slice-boundary-denominator/v1" ||
		value.Version != version || value.DenominatorID != denominatorID(version) ||
		len(value.Boundaries) != 6 || value.Boundaries[0].ID != "syntax" ||
		value.Boundaries[0].Target != capability {
		return nil, fmt.Errorf("baseline denominator does not match the registered corpus")
	}
	links := 0
	for _, item := range value.Boundaries {
		links += item.LinkTarget
	}
	if links != 12 {
		return nil, fmt.Errorf("baseline denominator changed link obligations")
	}
	value.Version++
	value.DenominatorID = denominatorID(value.Version)
	value.Boundaries[0].Target++
	out, err := json.MarshalIndent(value, "", "  ")
	return append(out, '\n'), err
}

func denominatorID(version int) string {
	return fmt.Sprintf("gooo.denominator.capability.vertical-slice-closure.v%d", version)
}

func digestName(version int) string { return fmt.Sprintf("DenominatorMigrationV%dDigest", version) }
func evidenceName(version int) string { return fmt.Sprintf("embeddedDenominatorV%d", version) }

func generateAdmission(raw []byte, version int) ([]byte, error) {
	source, err := parseGo(raw)
	if err != nil {
		return nil, err
	}
	decode, err := source.function("decodeDenominator")
	if err != nil {
		return nil, err
	}
	header, err := source.function("validateDenominator")
	if err != nil {
		return nil, err
	}
	digests, headers := 0, 0
	ast.Inspect(decode, func(node ast.Node) bool {
		check, ok := node.(*ast.IfStmt)
		if ok && strings.Contains(source.text(check.Cond), "digest != DenominatorDigest") {
			source.replace(check.Cond, "("+source.text(check.Cond)+") && digest != "+digestName(version))
			digests++
			return false
		}
		return true
	})
	ast.Inspect(header, func(node ast.Node) bool {
		assignment, ok := node.(*ast.AssignStmt)
		if ok && len(assignment.Lhs) == 1 && source.text(assignment.Lhs[0]) == "validHeader" && len(assignment.Rhs) == 1 {
			extension := fmt.Sprintf(" || (value.DenominatorID == %q && value.Version == %d)", denominatorID(version), version)
			source.replace(assignment.Rhs[0], "("+source.text(assignment.Rhs[0])+")"+extension)
			headers++
			return false
		}
		return true
	})
	if digests != 1 || headers != 1 {
		return nil, fmt.Errorf("denominator admission anchors are not exact")
	}
	return source.finish()
}

func generateDigest(raw, previous []byte, version int, next []byte) ([]byte, error) {
	source, err := parseGo(raw)
	if err != nil {
		return nil, err
	}
	bound := 0
	for _, declaration := range source.file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range group.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok || len(value.Names) != 1 || len(value.Values) != 1 {
				continue
			}
			if value.Names[0].Name == digestName(version) {
				return nil, fmt.Errorf("new denominator digest is already declared")
			}
			if value.Names[0].Name == digestName(version-1) {
				literal, ok := value.Values[0].(*ast.BasicLit)
				if !ok {
					return nil, fmt.Errorf("baseline digest is not literal")
				}
				text, err := strconv.Unquote(literal.Value)
				if err != nil || text != digest(previous) {
					return nil, fmt.Errorf("baseline denominator digest mismatch")
				}
				bound++
			}
		}
	}
	if bound != 1 {
		return nil, fmt.Errorf("baseline denominator digest is not uniquely pinned")
	}
	source.edits = append(source.edits, sourceEdit{len(raw), len(raw),
		fmt.Sprintf("\nconst %s = %q\n", digestName(version), digest(next))})
	return source.finish()
}
