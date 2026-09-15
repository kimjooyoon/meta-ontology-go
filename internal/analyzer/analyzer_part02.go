package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type parsedFile struct {
	filename    string
	packagePath string
	packageName string
	file        *ast.File
}

func parseSources(fileSet *token.FileSet, sources []SourceFile) ([]parsedFile, error) {
	parsed := make([]parsedFile, 0, len(sources))
	for _, source := range sources {
		filename := source.Filename
		if filename == "" {
			filename = "<source>"
		}
		file, err := parser.ParseFile(fileSet, filename, source.Source, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		packagePath := source.PackagePath
		if packagePath == "" {
			packagePath = file.Name.Name
		}
		parsed = append(parsed, parsedFile{
			filename:    filename,
			packagePath: packagePath,
			packageName: file.Name.Name,
			file:        file,
		})
	}
	return parsed, nil
}
