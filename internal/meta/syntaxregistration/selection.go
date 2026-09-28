package syntaxregistration

import (
	"fmt"
	"go/ast"
)

func generateSelection(source *goSource, version, capability int) error {
	for _, item := range []struct{ function, result string }{
		{"activeDenominator", evidenceName(version)},
		{"activeDenominatorDigest", digestName(version)},
	} {
		function, err := source.function(item.function)
		if err != nil {
			return err
		}
		count := 0
		ast.Inspect(function, func(node ast.Node) bool {
			selection, ok := node.(*ast.SwitchStmt)
			if !ok || selection.Tag == nil || source.text(selection.Tag) != "languagesyntax.FixedCapabilityTotal" {
				return true
			}
			matched := false
			for _, statement := range selection.Body.List {
				clause, ok := statement.(*ast.CaseClause)
				if !ok {
					continue
				}
				for _, expression := range clause.List {
					value, literal := integer(expression)
					if literal && value == capability+1 {
						count = -100
						return false
					}
					if literal && value == capability {
						matched = true
					}
				}
			}
			if matched {
				source.insert(selection.Body.Rbrace, fmt.Sprintf("\ncase %d:\nreturn %s\n", capability+1, item.result))
				count++
			}
			return false
		})
		if count != 1 {
			return fmt.Errorf("denominator selection %s lacks one exact baseline", item.function)
		}
	}
	var anchor ast.Node
	for _, declaration := range source.file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range group.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if ok && len(value.Names) == 1 && value.Names[0].Name == evidenceName(version-1) {
				if anchor != nil {
					return fmt.Errorf("baseline embedded evidence is ambiguous")
				}
				anchor = value
			}
		}
	}
	if anchor == nil {
		return fmt.Errorf("baseline embedded evidence is missing")
	}
	if err := source.requireImport(anchor, "embed", "_"); err != nil {
		return err
	}
	source.appendAt(anchor, fmt.Sprintf("\n//go:embed evidence/denominator-v%d.json\nvar %s []byte\n", version, evidenceName(version)))
	return nil
}
