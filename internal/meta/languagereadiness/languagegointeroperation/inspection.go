package languagegointeroperation

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"sort"
)

type inspection struct {
	Canonical       []byte
	APIDigest       string
	ExportedObjects int
	GenericMethods  int
	AliasNodes      int
	TypesBound      bool
}

type inspectionFailure struct {
	Stage string
	Code  string
}

func inspectSource(source []byte) (inspection, *inspectionFailure) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "reified.go", source, parser.ParseComments)
	if err != nil {
		return inspection{}, &inspectionFailure{Stage: "PARSE", Code: "GO_PARSE_REJECTED"}
	}
	if len(file.Imports) != 0 {
		return inspection{}, &inspectionFailure{Stage: "AUTHORITY", Code: "AMBIENT_IMPORT_REJECTED"}
	}
	canonical, err := formatAST(fileSet, file)
	if err != nil {
		return inspection{}, &inspectionFailure{Stage: "FORMAT", Code: "GO_FORMAT_REJECTED"}
	}
	config := types.Config{GoVersion: RequiredGoVersion}
	pkg, err := config.Check("gooo/interop/"+file.Name.Name, fileSet, []*ast.File{file}, nil)
	if err != nil {
		return inspection{}, &inspectionFailure{Stage: "TYPE", Code: "GO_TYPES_REJECTED"}
	}
	entries, typesBound := apiSurface(pkg)
	if len(entries) == 0 {
		return inspection{}, &inspectionFailure{Stage: "API", Code: "EXPORTED_API_EMPTY"}
	}
	methods, aliases := syntaxFeatureCounts(file)
	return inspection{Canonical: canonical, APIDigest: digestJSON(entries), ExportedObjects: len(entries),
		GenericMethods: methods, AliasNodes: aliases, TypesBound: typesBound}, nil
}

func formatAST(fileSet *token.FileSet, file *ast.File) ([]byte, error) {
	var output bytes.Buffer
	if err := format.Node(&output, fileSet, file); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

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
	for index := 0; index < named.NumMethods(); index++ {
		method := named.Method(index)
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
