package bodyexecution

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

func readDriverRegistrations(prior Composition, prefix string) map[string]string {
	result := map[string]string{}
	root, input := 0, 0
	for _, node := range prior.Plan.Activities {
		for _, slot := range node.inputSlots() {
			if slot.From >= 0 {
				continue
			}
			name := "input" + strconv.Itoa(root)
			root++
			result[name] = ""
			for _, record := range prior.Plan.Records {
				if record.Name != slot.Type {
					continue
				}
				for _, field := range record.Fields {
					if field.Cardinality == "many" {
						value := name + "." + field.GoName
						if field.Presence == "optional" {
							result[name] += fmt.Sprintf("if %s!=nil{%sRegister(*%s,%d)};", value, prefix, value, input)
						} else {
							result[name] += fmt.Sprintf("%sRegister(%s,%d);", prefix, value, input)
						}
						input++
					}
				}
			}
		}
	}
	return result
}

func readDriverStatements(source string) ([]ast.Stmt, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "statements.go", "package main;func placeholder(){"+source+"}", 0)
	if err != nil {
		return nil, err
	}
	return file.Decls[0].(*ast.FuncDecl).Body.List, nil
}

func observedReadDriver(source, prefix string, prior Composition, constructions ...bool) (string, error) {
	observer := readObserver
	if len(constructions) > 0 && constructions[0] {
		observer = constructionReadObserver
	}
	return observedReadDriverWithObserver(source, prefix, prior, observer)
}

func observedReadDriverWithObserver(source, prefix string, prior Composition, observer string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "driver.go", source, 0)
	if err != nil {
		return "", err
	}
	registrations := readDriverRegistrations(prior, prefix)
	ranges, encoders, inputs := 0, 0, 0
	var failure error
	ast.Inspect(file, func(node ast.Node) bool {
		if loop, ok := node.(*ast.RangeStmt); ok {
			if name, ok := loop.X.(*ast.Ident); ok && name.Name == "in" {
				key, ok := loop.Key.(*ast.Ident)
				if !ok {
					return false
				}
				statements, parseErr := readDriverStatements(prefix + "Reset(" + key.Name + ",len(in));")
				if parseErr != nil {
					failure = parseErr
					return false
				}
				for _, stmt := range loop.Body.List {
					statements = append(statements, stmt)
					name := readUnmarshaledInput(stmt)
					if registration, found := registrations[name]; found {
						added, parseErr := readDriverStatements(registration)
						if parseErr != nil {
							failure = parseErr
							return false
						}
						statements = append(statements, added...)
						inputs++
						delete(registrations, name)
					}
				}
				loop.Body.List = statements
				ranges++
			}
		}
		if call, ok := node.(*ast.CallExpr); ok && len(call.Args) == 1 {
			if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "Encode" {
				call.Args[0] = &ast.CallExpr{Fun: ast.NewIdent(prefix + "Save"), Args: []ast.Expr{call.Args[0]}}
				encoders++
			}
		}
		return failure == nil
	})
	if failure != nil {
		return "", failure
	}
	if ranges != 1 || encoders != 1 || inputs == 0 || len(registrations) != 0 {
		return "", fmt.Errorf("collection observation driver differs from its decoded input slots")
	}
	var out bytes.Buffer
	if err := format.Node(&out, fset, file); err != nil {
		return "", err
	}
	out.WriteString(strings.ReplaceAll(observer, "GoooRead", prefix))
	raw, err := format.Source(out.Bytes())
	return string(raw), err
}

func readUnmarshaledInput(stmt ast.Stmt) string {
	branch, ok := stmt.(*ast.IfStmt)
	if !ok {
		return ""
	}
	comparison, ok := branch.Cond.(*ast.BinaryExpr)
	if !ok || comparison.Op != token.NEQ {
		return ""
	}
	call, ok := comparison.X.(*ast.CallExpr)
	if !ok || len(call.Args) != 2 {
		return ""
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Unmarshal" {
		return ""
	}
	owner, ok := selector.X.(*ast.Ident)
	if !ok || owner.Name != "json" {
		return ""
	}
	address, ok := call.Args[1].(*ast.UnaryExpr)
	if !ok || address.Op != token.AND {
		return ""
	}
	name, ok := address.X.(*ast.Ident)
	if !ok {
		return ""
	}
	return name.Name
}
