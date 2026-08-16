package semanticbinding

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"sort"
	"strings"
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

func collectRecords(
	parsed []parsedSource,
	fileSet *token.FileSet,
	info *types.Info,
	registered map[string]struct{},
) ([]Binding, []Obligation, error) {
	state := recordState{ids: make(map[string]Span), bindingTargets: make(map[string]Span)}
	for _, source := range parsed {
		attachments := attachmentsFor(source.file, source.input.PackagePath)
		for _, group := range source.file.Comments {
			for _, comment := range group.List {
				if err := state.addComment(comment, group, attachments, fileSet, info, registered); err != nil {
					return nil, nil, err
				}
			}
		}
	}
	return state.bindings, state.obligations, nil
}

func (state *recordState) addComment(
	comment *ast.Comment,
	group *ast.CommentGroup,
	attachments map[*ast.CommentGroup][]declarationTarget,
	fileSet *token.FileSet,
	info *types.Info,
	registered map[string]struct{},
) error {
	current, found, err := parseDirective(comment, spanFor(fileSet, comment))
	if err != nil || !found {
		return err
	}
	targets := attachments[group]
	if len(targets) == 0 {
		return bindingError(CodeDetachedComment, current.span, "directive is not attached to a declaration")
	}
	if len(targets) != 1 {
		return bindingError(CodeAmbiguousBinding, current.span, "directive is attached to multiple declarations")
	}
	current, err = validateDirective(current)
	if err != nil {
		return err
	}
	if err := ensureRegistered(current, registered); err != nil {
		return err
	}
	target := targets[0]
	key, err := targetObjectKey(target.node, info)
	if err != nil {
		return withErrorSpan(err, current.span)
	}
	declarationSpan := spanFor(fileSet, target.node)
	id := current.fields["id"]
	if previous, exists := state.ids[id]; exists {
		return bindingError(CodeDuplicateID, current.span, fmt.Sprintf("identity is already bound at %s", previous))
	}
	state.ids[id] = current.span
	return state.addRecord(current, target, key, declarationSpan)
}

func (state *recordState) addRecord(
	current directive,
	target declarationTarget,
	key string,
	declarationSpan Span,
) error {
	id := current.fields["id"]
	if current.kind == "bind" {
		targetID := target.inputKey(key)
		if previous, exists := state.bindingTargets[targetID]; exists {
			return bindingError(
				CodeAmbiguousBinding, current.span,
				fmt.Sprintf("declaration already has a binding at %s", previous),
			)
		}
		state.bindingTargets[targetID] = current.span
		state.bindings = append(state.bindings, Binding{
			ID: id, Role: Role(current.fields["role"]),
			PackagePath: target.packagePathValue(), DeclarationKey: key,
			Span: declarationSpan, DirectiveSpan: current.span,
		})
		return nil
	}
	state.obligations = append(state.obligations, Obligation{
		ID: id, Subject: current.fields["subject"], Pressure: current.fields["pressure"],
		PackagePath: target.packagePathValue(), DeclarationKey: key,
		Span: declarationSpan, DirectiveSpan: current.span,
	})
	return nil
}

func attachmentsFor(file *ast.File, packagePath string) map[*ast.CommentGroup][]declarationTarget {
	result := make(map[*ast.CommentGroup][]declarationTarget)
	for _, declaration := range file.Decls {
		switch current := declaration.(type) {
		case *ast.FuncDecl:
			if current.Doc != nil {
				result[current.Doc] = append(result[current.Doc], declarationTarget{
					node: current, file: file, packagePath: packagePath,
				})
			}
		case *ast.GenDecl:
			if current.Doc != nil {
				result[current.Doc] = append(result[current.Doc], declarationTarget{
					node: current, file: file, packagePath: packagePath,
				})
			}
			for _, specification := range current.Specs {
				if typeSpec, ok := specification.(*ast.TypeSpec); ok && typeSpec.Doc != nil {
					result[typeSpec.Doc] = append(result[typeSpec.Doc], declarationTarget{
						node: typeSpec, file: file, packagePath: packagePath,
					})
				}
			}
		}
	}
	return result
}

func (t declarationTarget) packagePathValue() string {
	return t.packagePath
}

func (t declarationTarget) inputKey(key string) string {
	return t.packagePath + "\x00" + key
}

func unknownResult(err error) (Result, error) {
	typed, ok := err.(*Error)
	if !ok {
		typed = bindingError(CodeInput, Span{}, err.Error())
	}
	unknown := Unknown{Code: typed.Code, Message: typed.Message, Span: typed.Span, FullSuiteFallback: true}
	return Result{Status: StatusUnknown, Unknowns: []Unknown{unknown}, FullSuiteFallback: true}, typed
}
