package syntaxregistration

import (
	"fmt"
	"go/ast"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func generatePartitionNative(inputs map[string][]byte, before, after languagesyntax.Registry) (map[string][]byte, error) {
	if before.Schema != after.Schema || !reflect.DeepEqual(before.Cases, after.Cases) ||
		!reflect.DeepEqual(before.ProjectionUnits, after.ProjectionUnits) {
		return nil, fmt.Errorf("physical partition cannot change schema, cases or projection consumers")
	}
	source, err := parseSourceUnits(inputs, syntaxRoot, "languagesyntax", false)
	if err != nil {
		return nil, err
	}
	binding, err := bindPartitionNative(source, before)
	if err != nil {
		return nil, err
	}
	// An internal Go construction label, not a fabricated Gooo activity ID or
	// execution authority. The caller still owns operation admission/publication.
	source.activity = "native/partition-registry-membership"
	if err := editPartitionFamilies(source, binding, before.CaseSources, after.CaseSources); err != nil {
		return nil, err
	}
	if err := editPartitionPackages(source, binding, before.PackageUnits, after.PackageUnits); err != nil {
		return nil, err
	}
	if !slices.Equal(before.MetaSources, after.MetaSources) {
		if len(binding.meta) == 0 {
			return nil, fmt.Errorf("native meta-source sequence has no insertion origin")
		}
		if err := appendPartitionMember(source, binding.meta[len(binding.meta)-1], before.MetaSources, after.MetaSources); err != nil {
			return nil, err
		}
	}
	outputs, err := source.finish()
	if err != nil {
		return nil, err
	}
	projected := maps.Clone(inputs)
	maps.Copy(projected, outputs)
	native, err := parseSourceUnits(projected, syntaxRoot, "languagesyntax", false)
	if err != nil {
		return nil, err
	}
	if _, err := bindPartitionNative(native, after); err != nil {
		return nil, err
	}
	return outputs, nil
}

func editPartitionFamilies(source *goSource, binding partitionNativeBinding, before, after []languagesyntax.CaseSourceSet) error {
	if len(after) < len(before) || len(after) > len(before)+1 {
		return fmt.Errorf("partition changed source-family population")
	}
	for i, group := range before {
		if group.CaseID != after[i].CaseID {
			return fmt.Errorf("partition changed source-family identity or ordering")
		}
		if !slices.Equal(group.Members, after[i].Members) {
			list, err := partitionMembersLiteral(binding.families[i])
			if err != nil {
				return err
			}
			if err := appendPartitionMember(source, list, group.Members, after[i].Members); err != nil {
				return err
			}
		}
	}
	if len(after) == len(before) {
		return nil
	}
	group := after[len(before)]
	quoted := make([]string, len(group.Members))
	for i, name := range group.Members {
		quoted[i] = strconv.Quote(name)
	}
	row := "{CaseID: " + strconv.Quote(group.CaseID) + ", Members: []string{" + strings.Join(quoted, ", ") + "}}"
	root, err := source.function("expectedRegistry")
	if err != nil {
		return err
	}
	leaves, err := source.registryFieldLeaves(root, "CaseSources", "CaseSourceSet", true)
	if err != nil {
		return err
	}
	if len(leaves) == 0 {
		source.appendRegistryElement(binding.root, "CaseSources: []CaseSourceSet{"+row+"}")
	} else {
		source.appendRegistryElement(leaves[len(leaves)-1], row)
	}
	return nil
}

func editPartitionPackages(source *goSource, binding partitionNativeBinding, before, after []languagesyntax.PackageDefinition) error {
	if len(before) != len(after) {
		return fmt.Errorf("partition changed package population")
	}
	for i, unit := range before {
		want := after[i]
		want.Members = unit.Members
		if !reflect.DeepEqual(unit, want) {
			return fmt.Errorf("partition changed package identity or execution metadata")
		}
		if slices.Equal(unit.Members, after[i].Members) {
			continue
		}
		list, err := partitionMembersLiteral(binding.packages[i])
		if err != nil {
			return err
		}
		if err := appendPartitionMember(source, list, unit.Members, after[i].Members); err != nil {
			return err
		}
	}
	return nil
}

func partitionMembersLiteral(literal *ast.CompositeLit) (*ast.CompositeLit, error) {
	value, err := registryLiteralField(literal, "Members")
	if err != nil {
		return nil, err
	}
	list, ok := value.(*ast.CompositeLit)
	if !ok {
		return nil, fmt.Errorf("partition membership has no exact literal origin")
	}
	return list, nil
}

func appendPartitionMember(source *goSource, list *ast.CompositeLit, before, after []string) error {
	if len(after) != len(before)+1 || !slices.Equal(before, after[:len(before)]) {
		return fmt.Errorf("physical partition must append exactly one member without replacing existing addresses")
	}
	source.appendRegistryElement(list, strconv.Quote(after[len(before)]))
	return nil
}

func (source *goSource) appendRegistryElement(list *ast.CompositeLit, text string) {
	prefix := " "
	if len(list.Elts) != 0 && source.registryTailNeedsComma(list) {
		end := list.Elts[len(list.Elts)-1].End()
		if end == list.Rbrace {
			prefix = ", "
		} else {
			source.insert(end, ",")
		}
	}
	source.insert(list.Rbrace, prefix+text+",")
}
