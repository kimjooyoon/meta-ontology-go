package syntaxregistration

import (
	"go/ast"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func readNativePackageUnits(source *goSource) ([]languagesyntax.PackageDefinition, []*ast.CompositeLit, error) {
	root, err := source.function("expectedRegistry")
	if err != nil {
		return nil, nil, err
	}
	origins, err := source.registryFieldOrigins(root, "PackageUnits", "PackageDefinition", true)
	if err != nil {
		return nil, nil, err
	}
	var units []languagesyntax.PackageDefinition
	var literals []*ast.CompositeLit
	for _, leaf := range origins.leaves {
		for _, element := range leaf.Elts {
			literal, err := origins.literal(element, "PackageDefinition")
			if err != nil {
				return nil, nil, err
			}
			unit, err := readNativePackageLiteral(source, literal)
			if err != nil {
				return nil, nil, err
			}
			units, literals = append(units, unit), append(literals, literal)
		}
	}
	return units, literals, nil
}

func readNativePackageLiteral(source *goSource, literal *ast.CompositeLit) (languagesyntax.PackageDefinition, error) {
	var unit languagesyntax.PackageDefinition
	err := readPartitionRecord(source, literal, &unit)
	return unit, err
}
