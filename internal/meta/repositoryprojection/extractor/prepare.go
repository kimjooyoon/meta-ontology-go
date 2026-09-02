package extractor

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
)

func prepareParsedSource(root, logical string, source []byte, fset *token.FileSet, file *ast.File) ([]byte, *token.FileSet, *ast.File, error) {
	prepared, err := prepareOversizedFunctions(root, logical, source, fset, file)
	if err != nil {
		return nil, nil, nil, err
	}
	if bytes.Equal(prepared, source) {
		return source, fset, file, nil
	}
	preparedSet := token.NewFileSet()
	preparedFile, err := parser.ParseFile(preparedSet, logical, prepared, parser.ParseComments)
	if err != nil {
		return nil, nil, nil, fail("rewrite-source", "parse-decomposed-source", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
	}
	return prepared, preparedSet, preparedFile, nil
}
