package syntaxregistration

import (
	"fmt"
	"go/ast"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func bindNativeCaseSources(source *goSource, expected []languagesyntax.CaseSourceSet) error {
	root, err := source.function("expectedRegistry")
	if err != nil {
		return err
	}
	leaves, err := source.registryFieldLeaves(root, "CaseSources", "CaseSourceSet", true)
	if err != nil {
		return err
	}
	var groups []languagesyntax.CaseSourceSet
	for _, leaf := range leaves {
		for _, element := range leaf.Elts {
			literal, ok := element.(*ast.CompositeLit)
			if !ok || (literal.Type != nil && !namedRegistryType(literal.Type, "CaseSourceSet")) {
				return fmt.Errorf("native source family is not an explicit CaseSourceSet")
			}
			group, err := readCaseSourceLiteral(source, literal)
			if err != nil {
				return err
			}
			groups = append(groups, group)
		}
	}
	if !reflect.DeepEqual(groups, expected) {
		return fmt.Errorf("native source families differ from the corpus")
	}
	return nil
}

func readCaseSourceLiteral(source *goSource, literal *ast.CompositeLit) (languagesyntax.CaseSourceSet, error) {
	var group languagesyntax.CaseSourceSet
	seen := map[string]bool{}
	for _, element := range literal.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			return group, fmt.Errorf("native source family requires named fields")
		}
		key, ok := pair.Key.(*ast.Ident)
		if !ok || seen[key.Name] {
			return group, fmt.Errorf("native source family field is ambiguous")
		}
		seen[key.Name] = true
		switch key.Name {
		case "CaseID":
			value, err := profileConstantValue(source, pair.Value, map[string]bool{})
			if err != nil {
				return group, err
			}
			if err := assignProfileCaseValue(reflect.ValueOf(&group.CaseID).Elem(), value); err != nil {
				return group, err
			}
		case "Members":
			list, ok := pair.Value.(*ast.CompositeLit)
			if !ok {
				return group, fmt.Errorf("native source family members are not an explicit slice")
			}
			var err error
			group.Members, err = nativeMetaSourceValues(source, list)
			if err != nil {
				return group, err
			}
		default:
			return group, fmt.Errorf("native source family field is unknown")
		}
	}
	if len(seen) != 2 {
		return group, fmt.Errorf("native source family is incomplete")
	}
	return group, nil
}
