package languagegointeroperation

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
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
