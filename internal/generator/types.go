package generator

import (
	"fmt"
	"go/importer"
	"go/token"
	gotypes "go/types"
)

func validateGoTypes(ir SemanticIR) error {
	packageScope := gotypes.NewPackage(ir.Package, ir.Package)
	if err := insertEntityTypes(packageScope, ir.Entities); err != nil {
		return err
	}
	if err := insertImportTypes(packageScope, ir.Imports); err != nil {
		return err
	}
	for _, entity := range ir.Entities {
		for _, field := range entity.Fields {
			if err := validateTypeExpr(packageScope, field.GoType, "field "+field.ID); err != nil {
				return err
			}
		}
	}
	for _, activity := range ir.Activities {
		for _, port := range append(append([]Port(nil), activity.Inputs...), activity.Outputs...) {
			if err := validateTypeExpr(packageScope, port.GoType, "port "+port.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func insertEntityTypes(packageScope *gotypes.Package, entities []Entity) error {
	for _, entity := range entities {
		name := gotypes.NewTypeName(token.NoPos, packageScope, entity.GoName, nil)
		gotypes.NewNamed(name, gotypes.NewStruct(nil, nil), nil)
		if existing := packageScope.Scope().Insert(name); existing != nil {
			return fmt.Errorf("generator: Go type name %q collides with %s", entity.GoName, existing.Name())
		}
	}
	return nil
}

func insertImportTypes(packageScope *gotypes.Package, imports []Import) error {
	for _, item := range imports {
		if item.Name == "_" {
			continue
		}
		imported, err := importer.Default().Import(item.Path)
		if err != nil {
			return fmt.Errorf("generator: import %q cannot be resolved for type validation: %w", item.Path, err)
		}
		name := item.Name
		if name == "" {
			name = imported.Name()
		}
		object := gotypes.NewPkgName(token.NoPos, packageScope, name, imported)
		if existing := packageScope.Scope().Insert(object); existing != nil {
			return fmt.Errorf("generator: import name %q collides with %s", name, existing.Name())
		}
	}
	return nil
}

func validateTypeExpr(packageScope *gotypes.Package, expression, context string) error {
	if _, err := gotypes.Eval(token.NewFileSet(), packageScope, token.NoPos, expression); err != nil {
		return fmt.Errorf("generator: invalid Go type %q for %s: %w", expression, context, err)
	}
	return nil
}
