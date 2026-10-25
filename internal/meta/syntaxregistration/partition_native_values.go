package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/constant"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagepackageexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/receiptprojection"
)

func readPartitionRecord(source *goSource, literal *ast.CompositeLit, record any) error {
	target, seen := reflect.ValueOf(record).Elem(), map[string]bool{}
	for _, element := range literal.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			return fmt.Errorf("native registry row requires named fields")
		}
		key, ok := pair.Key.(*ast.Ident)
		if !ok || seen[key.Name] || !target.FieldByName(key.Name).IsValid() {
			return fmt.Errorf("native registry field is unknown or repeated")
		}
		seen[key.Name] = true
		field := target.FieldByName(key.Name)
		if field.Type() == reflect.TypeFor[[]string]() {
			list, ok := pair.Value.(*ast.CompositeLit)
			if !ok {
				return fmt.Errorf("native registry membership requires an explicit string slice")
			}
			values, err := nativeMetaSourceValues(source, list)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(values))
			continue
		}
		value, err := partitionFieldValue(source, key.Name, pair.Value)
		if err != nil {
			return err
		}
		if err := assignProfileCaseValue(field, value); err != nil {
			return err
		}
	}
	return nil
}

func partitionFieldValue(source *goSource, field string, expression ast.Expr) (constant.Value, error) {
	selector, selected := expression.(*ast.SelectorExpr)
	if !selected {
		return profileConstantValue(source, expression, map[string]bool{})
	}
	alias, ok := selector.X.(*ast.Ident)
	if !ok {
		return nil, fmt.Errorf("native registry constant has an unknown qualifier")
	}
	// Bind the two existing native ABI constants by import and export identity.
	// This does not execute or evaluate arbitrary imported Go expressions.
	var imported, value string
	switch {
	case field == "ReportSchema" && selector.Sel.Name == "ReportSchema":
		imported, value = "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagepackageexecution", languagepackageexecution.ReportSchema
	case field == "Profile" && selector.Sel.Name == "Profile":
		imported, value = "github.com/kimjooyoon/meta-ontology-go/internal/receiptprojection", receiptprojection.Profile
	default:
		return nil, fmt.Errorf("native registry constant has an unknown export")
	}
	if err := source.requireImport(selector, imported, alias.Name); err != nil {
		return nil, err
	}
	if registryImportShadowed(source, selector, alias.Name) {
		return nil, fmt.Errorf("native registry constant import is shadowed")
	}
	return constant.MakeString(value), nil
}

func registryImportShadowed(source *goSource, selector *ast.SelectorExpr, alias string) bool {
	for _, declaration := range source.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Pos() > selector.Pos() || function.End() < selector.End() || function.Body == nil {
			continue
		}
		for _, statement := range function.Body.List {
			if assignment, ok := statement.(*ast.AssignStmt); ok {
				for _, name := range assignment.Lhs {
					if namedRegistryType(name, alias) {
						return true
					}
				}
			}
		}
	}
	return false
}
