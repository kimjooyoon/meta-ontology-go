package syntaxregistration

import (
	"fmt"
	"go/ast"
	"strings"
)

func admissionHelperCount(source *goSource, prefix string) int {
	count := 0
	for _, declaration := range source.file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && strings.HasPrefix(function.Name.Name, prefix) {
			count++
		}
		if group, ok := declaration.(*ast.GenDecl); ok {
			for _, spec := range group.Specs {
				if value, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range value.Names {
						if strings.HasPrefix(name.Name, prefix) {
							count++
						}
					}
				}
				if value, ok := spec.(*ast.TypeSpec); ok && strings.HasPrefix(value.Name.Name, prefix) {
					count++
				}
			}
		}
	}
	return count
}

func (set *admissionSet) readChunks(source *goSource) error {
	if source.text(set.root) != set.rootCall() {
		return fmt.Errorf("denominator admission root differs")
	}
	count := admissionHelperCount(source, set.prefix)
	for index := range count {
		helper, err := source.function(set.helperName(index))
		if err != nil {
			return err
		}
		if !set.validHelper(source, helper) {
			return fmt.Errorf("denominator admission helper signature differs")
		}
		selection, ok := helper.Body.List[0].(*ast.SwitchStmt)
		if !ok || selection.Init != nil || selection.Tag == nil || source.text(selection.Tag) != set.tag ||
			len(selection.Body.List) == 0 || len(selection.Body.List) > admissionChunkRows {
			return fmt.Errorf("denominator admission helper selection differs")
		}
		for _, statement := range selection.Body.List {
			if err := set.readClause(source, statement); err != nil {
				return err
			}
		}
		expected := "false"
		if index+1 < count {
			expected = set.helperName(index+1) + "(" + set.argument + ")"
		}
		result, ok := helper.Body.List[1].(*ast.ReturnStmt)
		if !ok || len(result.Results) != 1 || source.text(result.Results[0]) != expected {
			return fmt.Errorf("denominator admission helper continuation differs")
		}
		set.helpers = append(set.helpers, helper)
	}
	return nil
}

func (set admissionSet) validHelper(source *goSource, helper *ast.FuncDecl) bool {
	if helper.Recv != nil || helper.Type.TypeParams != nil || helper.Type.Params.NumFields() != 1 ||
		helper.Type.Results.NumFields() != 1 || helper.Body == nil || len(helper.Body.List) != 2 {
		return false
	}
	parameter := helper.Type.Params.List[0]
	return len(parameter.Names) == 1 && parameter.Names[0].Name == set.argument &&
		source.text(parameter.Type) == set.parameter && source.text(helper.Type.Results.List[0].Type) == "bool"
}

func (set *admissionSet) readClause(source *goSource, statement ast.Stmt) error {
	clause, ok := statement.(*ast.CaseClause)
	if !ok || len(clause.List) != 1 || len(clause.Body) != 1 {
		return fmt.Errorf("denominator admission clause differs")
	}
	result, ok := clause.Body[0].(*ast.ReturnStmt)
	if !ok || len(result.Results) != 1 {
		return fmt.Errorf("denominator admission result differs")
	}
	var row admissionRow
	if set.argument == "digest" {
		name, named := clause.List[0].(*ast.Ident)
		if !named || source.text(result.Results[0]) != "true" {
			return fmt.Errorf("denominator admission digest case differs")
		}
		version, valid := digestVersion(name.Name)
		if !valid {
			return fmt.Errorf("denominator admission digest identity differs")
		}
		row = admissionRow{version, name.Name, "true"}
	} else {
		version, literal := integer(clause.List[0])
		row = headerAdmissionRow(version)
		if !literal || version < 21 || source.text(result.Results[0]) != row.result {
			return fmt.Errorf("denominator admission header case differs")
		}
	}
	return set.retainRow(row)
}
