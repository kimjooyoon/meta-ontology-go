package semanticbinding

import (
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

func parseInputs(sources []SourceFile) ([]parsedSource, *token.FileSet, string, error) {
	sort.Slice(sources, func(left, right int) bool { return sources[left].Filename < sources[right].Filename })
	packagePath := sources[0].PackagePath
	seenFiles := make(map[string]struct{}, len(sources))
	parsed := make([]parsedSource, 0, len(sources))
	fileSet := token.NewFileSet()
	packageName := ""
	for _, source := range sources {
		if _, exists := seenFiles[source.Filename]; exists {
			return nil, nil, "", bindingError(CodeInput, Span{}, "source filenames must be unique")
		}
		seenFiles[source.Filename] = struct{}{}
		if source.PackagePath != packagePath {
			return nil, nil, "", bindingError(CodeInput, Span{}, "all source files must use one explicit package path")
		}
		file, err := parser.ParseFile(fileSet, source.Filename, source.Source, parser.ParseComments)
		if err != nil {
			return nil, nil, "", bindingError(CodeParse, Span{}, err.Error())
		}
		if packageName == "" {
			packageName = file.Name.Name
		} else if packageName != file.Name.Name {
			return nil, nil, "", bindingError(CodeInput, spanFor(fileSet, file), "source files must use one Go package name")
		}
		parsed = append(parsed, parsedSource{input: source, file: file})
	}
	if strings.TrimSpace(packagePath) == "" {
		return nil, nil, "", bindingError(CodeInput, Span{}, "package path is required")
	}
	return parsed, fileSet, packagePath, nil
}
func registeredIDs(values []string) (map[string]struct{}, error) {
	if values == nil {
		return nil, nil
	}
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		canonical, err := normalizeIdentity(value)
		if err != nil {
			return nil, withErrorSpan(err, Span{})
		}
		if _, exists := result[canonical]; exists {
			return nil, bindingError(CodeDuplicateID, Span{}, "registered identity is duplicated")
		}
		result[canonical] = struct{}{}
	}
	return result, nil
}
