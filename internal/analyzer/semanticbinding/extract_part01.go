package semanticbinding

import (
	"go/ast"
)

type parsedSource struct {
	input SourceFile
	file  *ast.File
}
type declarationTarget struct {
	node        ast.Node
	file        *ast.File
	packagePath string
}
type recordState struct {
	bindings       []Binding
	obligations    []Obligation
	ids            map[string]Span
	bindingTargets map[string]Span
}

func Extract(input Input) (Result, error) {
	sources, err := input.sourceInputs()
	if err != nil {
		return unknownResult(err)
	}
	registered, err := registeredIDs(input.RegisteredIDs)
	if err != nil {
		return unknownResult(err)
	}
	parsed, fileSet, packagePath, err := parseInputs(sources)
	if err != nil {
		return unknownResult(err)
	}
	info, err := typeCheck(parsed, fileSet, packagePath)
	if err != nil {
		return unknownResult(err)
	}
	bindings, obligations, err := collectRecords(parsed, fileSet, info, registered)
	if err != nil {
		return unknownResult(err)
	}
	return makeResult(bindings, obligations), nil
}

// ExtractPackage is the package-oriented spelling of Extract.
func ExtractPackage(sources []SourceFile) (Result, error) {
	return Extract(Input{Sources: sources})
}

// ParsePackage is an alias retained for callers that use parser vocabulary.
func ParsePackage(sources []SourceFile) (Result, error) {
	return ExtractPackage(sources)
}
