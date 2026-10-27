package syntaxregistration

import (
	"fmt"
	"go/ast"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

type profileNativeCase struct {
	definition languagesyntax.CaseDefinition
	literal    *ast.CompositeLit
}

// Read the finite Cases expression. Incidental literals are not case origins.
func readProfileNativeCases(source *goSource) ([]profileNativeCase, error) {
	root, err := source.function("expectedRegistry")
	if err != nil {
		return nil, err
	}
	origins, err := source.registryFieldOrigins(root, "Cases", "CaseDefinition", false)
	if err != nil {
		return nil, err
	}
	var result []profileNativeCase
	ids, paths := map[string]bool{}, map[string]bool{}
	for _, leaf := range origins.leaves {
		for _, element := range leaf.Elts {
			literal, err := origins.literal(element, "CaseDefinition")
			if err != nil {
				return nil, err
			}
			definition, err := readProfileCaseLiteral(source, literal)
			if err != nil {
				return nil, err
			}
			if definition.ID == "" || definition.Path == "" || ids[definition.ID] || paths[definition.Path] {
				return nil, fmt.Errorf("PROFILE_NATIVE_CASE_IDENTITY_AMBIGUOUS")
			}
			ids[definition.ID], paths[definition.Path] = true, true
			result = append(result, profileNativeCase{definition, literal})
		}
	}
	return result, nil
}

func readProfileCaseLiteral(source *goSource, literal *ast.CompositeLit) (languagesyntax.CaseDefinition, error) {
	var definition languagesyntax.CaseDefinition
	target := reflect.ValueOf(&definition).Elem()
	seen := map[string]bool{}
	for _, element := range literal.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			return definition, fmt.Errorf("PROFILE_NATIVE_CASE_FIELDS_UNRESOLVED")
		}
		key, ok := pair.Key.(*ast.Ident)
		if !ok || seen[key.Name] {
			return definition, fmt.Errorf("PROFILE_NATIVE_CASE_FIELD_AMBIGUOUS")
		}
		field := target.FieldByName(key.Name)
		if !field.IsValid() || !field.CanSet() {
			return definition, fmt.Errorf("PROFILE_NATIVE_CASE_FIELD_UNKNOWN")
		}
		value, err := profileConstantValue(source, pair.Value, map[string]bool{})
		if err != nil {
			return definition, err
		}
		if err := assignProfileCaseValue(field, value); err != nil {
			return definition, err
		}
		seen[key.Name] = true
	}
	return definition, nil
}
