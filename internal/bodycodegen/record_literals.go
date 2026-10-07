package bodycodegen

import (
	"fmt"
	"go/ast"
)

func validateRecordLiterals(body *ast.BlockStmt, records []RecordType) error {
	var failure error
	ast.Inspect(body, func(node ast.Node) bool {
		if failure != nil {
			return false
		}
		if literal, ok := node.(*ast.CompositeLit); ok {
			failure = validateRecordLiteral(literal, records)
		}
		return failure == nil
	})
	return failure
}

func validateRecordLiteral(literal *ast.CompositeLit, records []RecordType) error {
	name, ok := literal.Type.(*ast.Ident)
	if !ok {
		return fmt.Errorf("record construction requires a declared record name")
	}
	for _, record := range records {
		if record.Name != name.Name {
			continue
		}
		seen := make(map[string]bool, len(record.Fields))
		for _, element := range literal.Elts {
			pair, ok := element.(*ast.KeyValueExpr)
			if !ok {
				return fmt.Errorf("record %q requires explicit named fields", record.Name)
			}
			key, ok := pair.Key.(*ast.Ident)
			if !ok || seen[key.Name] {
				return fmt.Errorf("record %q field is duplicated or unnamed", record.Name)
			}
			seen[key.Name] = true
		}
		for _, field := range record.Fields {
			if !seen[field.Name] && field.Presence != "optional" {
				return fmt.Errorf("record %q requires field %q", record.Name, field.Name)
			}
		}
		for name := range seen {
			declared := false
			for _, field := range record.Fields {
				if name == field.Name {
					declared = true
					break
				}
			}
			if !declared {
				return fmt.Errorf("record %q contains an undeclared field", record.Name)
			}
		}
		return nil
	}
	return fmt.Errorf("record %q is not declared in the body profile", name.Name)
}

func validateRecordExpression(literal *ast.CompositeLit) error {
	if _, ok := literal.Type.(*ast.Ident); !ok {
		return fmt.Errorf("record construction requires a declared record name")
	}
	for _, element := range literal.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			return fmt.Errorf("record construction requires named fields")
		}
		if _, ok := pair.Key.(*ast.Ident); !ok {
			return fmt.Errorf("record field requires an identifier")
		}
		if err := validateExpression(pair.Value); err != nil {
			return err
		}
	}
	return nil
}
