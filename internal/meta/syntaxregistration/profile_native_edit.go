package syntaxregistration

import (
	"bytes"
	"fmt"
	"go/ast"
	"maps"
	"reflect"
	"strconv"
)

func (plan ProfileMigrationPlan) generateProfileNative() (string, []byte, error) {
	source, err := parseSourceUnits(plan.inputs, syntaxRoot, "languagesyntax", false)
	if err != nil {
		return "", nil, err
	}
	rows, err := readProfileNativeCases(source)
	if err != nil {
		return "", nil, err
	}
	if len(rows) != len(plan.registry.Cases) || !reflect.DeepEqual(rows[plan.index].definition, plan.request.Case) {
		return "", nil, fmt.Errorf("PROFILE_NATIVE_CASE_DIVERGED")
	}
	literal := rows[plan.index].literal
	path, _ := source.location(literal.Pos())
	source.activity = plan.binding.activities[1]
	editProfileLiteral(source, literal, plan.request.TargetProfile)
	outputs, err := source.finish()
	if err != nil {
		return "", nil, err
	}
	if len(outputs) != 1 || outputs[path] == nil {
		return "", nil, fmt.Errorf("PROFILE_NATIVE_WRITE_SET_DIVERGED")
	}
	if err := plan.validateProfileNative(source, literal, path, outputs[path]); err != nil {
		return "", nil, err
	}
	return path, outputs[path], nil
}

func editProfileLiteral(source *goSource, literal *ast.CompositeLit, target string) {
	profilePresent := false
	for _, element := range literal.Elts {
		pair := element.(*ast.KeyValueExpr)
		switch pair.Key.(*ast.Ident).Name {
		case "EntityFields", "ImplicitActivityPorts", "OrderedCollections":
			source.replace(pair.Value, "false")
		case "EntityFieldsProfile":
			source.replace(pair.Value, strconv.Quote(target))
			profilePresent = true
		}
	}
	if !profilePresent {
		prefix := " "
		if source.registryTailNeedsComma(literal) {
			end := literal.Elts[len(literal.Elts)-1].End()
			if end == literal.Rbrace {
				prefix = ", "
			} else {
				source.insert(end, ",")
			}
		}
		source.insert(literal.Rbrace, prefix+"EntityFieldsProfile: "+strconv.Quote(target)+",")
	}
}

func (plan ProfileMigrationPlan) validateProfileNative(before *goSource, literal *ast.CompositeLit,
	path string, output []byte) error {
	inputs := maps.Clone(plan.inputs)
	inputs[path] = output
	after, err := parseSourceUnits(inputs, syntaxRoot, "languagesyntax", false)
	if err != nil {
		return err
	}
	rows, err := readProfileNativeCases(after)
	if err != nil {
		return err
	}
	if len(rows) != len(plan.registry.Cases) {
		return fmt.Errorf("PROFILE_NATIVE_POPULATION_CHANGED")
	}
	for index, row := range rows {
		want := plan.registry.Cases[index]
		if index == plan.index {
			want = profileAfterCase(plan.request)
		}
		if !reflect.DeepEqual(row.definition, want) {
			return fmt.Errorf("PROFILE_NATIVE_UNRELATED_CASE_CHANGED")
		}
	}
	if !bytes.Equal(profileOutsideLiteral(before, literal), profileOutsideLiteral(after, rows[plan.index].literal)) {
		return fmt.Errorf("PROFILE_NATIVE_UNRELATED_BYTES_CHANGED")
	}
	return nil
}

func profileOutsideLiteral(source *goSource, literal *ast.CompositeLit) []byte {
	path, start := source.location(literal.Pos())
	_, end := source.location(literal.End())
	raw := source.units[path]
	return append(bytes.Clone(raw[:start]), raw[end:]...)
}
