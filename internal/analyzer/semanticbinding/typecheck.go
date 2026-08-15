package semanticbinding

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
)

func typeCheck(parsed []parsedSource, fileSet *token.FileSet, packagePath string) (*types.Info, error) {
	info := &types.Info{Defs: make(map[*ast.Ident]types.Object)}
	var firstError error
	config := types.Config{
		Importer: importer.Default(),
		Error: func(err error) {
			if firstError == nil {
				firstError = err
			}
		},
	}
	files := make([]*ast.File, 0, len(parsed))
	for _, source := range parsed {
		files = append(files, source.file)
	}
	_, err := config.Check(packagePath, fileSet, files, info)
	if err != nil {
		return nil, bindingError(CodeTypeCheck, Span{}, err.Error())
	}
	if firstError != nil {
		return nil, bindingError(CodeTypeCheck, Span{}, firstError.Error())
	}
	return info, nil
}

func targetObjectKey(node ast.Node, info *types.Info) (string, error) {
	identifiers := declarationIdentifiers(node)
	if len(identifiers) != 1 {
		return "", bindingError(CodeAmbiguousBinding, Span{}, "declaration does not identify exactly one named object")
	}
	object := info.Defs[identifiers[0]]
	if object == nil {
		return "", bindingError(CodeMissingObject, Span{}, "go/types did not register the declaration object")
	}
	return objectKey(object), nil
}

func declarationIdentifiers(node ast.Node) []*ast.Ident {
	switch current := node.(type) {
	case *ast.FuncDecl:
		return []*ast.Ident{current.Name}
	case *ast.TypeSpec:
		return []*ast.Ident{current.Name}
	case *ast.GenDecl:
		var result []*ast.Ident
		for _, specification := range current.Specs {
			switch currentSpec := specification.(type) {
			case *ast.TypeSpec:
				result = append(result, currentSpec.Name)
			case *ast.ValueSpec:
				result = append(result, currentSpec.Names...)
			}
		}
		return result
	default:
		return nil
	}
}

func objectKey(object types.Object) string {
	key := object.Id()
	if function, ok := object.(*types.Func); ok {
		if signature, ok := function.Type().(*types.Signature); ok && signature.Recv() != nil {
			key += "|receiver=" + types.TypeString(signature.Recv().Type(), func(pkg *types.Package) string {
				if pkg == nil {
					return ""
				}
				return pkg.Path()
			})
		}
	}
	if key == "" {
		key = fmt.Sprintf("%T:%s", object, object.Name())
	}
	return key
}
