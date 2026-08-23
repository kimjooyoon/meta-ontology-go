package languagegointeroperation

import (
	"go/ast"
	"go/types"
	"sort"
)

func apiSurface(pkg *types.Package) ([]string, bool) {
	entries := []string{}
	typesBound := true
	qualifier := func(other *types.Package) string {
		if other == pkg {
			return ""
		}
		return other.Path()
	}
	for _, name := range pkg.Scope().Names() {
		object := pkg.Scope().Lookup(name)
		if !object.Exported() {
			continue
		}
		entries = append(entries, types.ObjectString(object, qualifier))
		entries = append(entries, namedMethods(object, qualifier)...)
		typesBound = typesBound && typeIdentityBound(object.Type())
	}
	sort.Strings(entries)
	return entries, typesBound
}

func namedMethods(object types.Object, qualifier types.Qualifier) []string {
	typeName, ok := object.(*types.TypeName)
	if !ok {
		return nil
	}
	named, ok := types.Unalias(typeName.Type()).(*types.Named)
	if !ok {
		return nil
	}
	methods := make([]string, 0, named.NumMethods())
	for method := range named.Methods() {
		method := method
		if method.Exported() {
			methods = append(methods, types.ObjectString(method, qualifier))
		}
	}
	return methods
}

func typeIdentityBound(value types.Type) bool {
	hasher := types.Hasher{}
	if alias, ok := value.(*types.Alias); ok {
		return hasher.Equal(alias, types.Unalias(alias))
	}
	return hasher.Equal(value, value)
}

func syntaxFeatureCounts(file *ast.File) (int, int) {
	methods, aliases := 0, 0
	ast.Inspect(file, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.FuncDecl:
			if value.Recv != nil && value.Type.TypeParams != nil && len(value.Type.TypeParams.List) > 0 {
				methods++
			}
		case *ast.TypeSpec:
			if value.Assign.IsValid() {
				aliases++
			}
		}
		return true
	})
	return methods, aliases
}
