package syntaxregistration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func generateCorpus(raw []byte, request Request) ([]byte, error) {
	var document map[string]json.RawMessage
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	old := document["cases"]
	if len(old) == 0 || bytes.Count(raw, old) != 1 {
		return nil, fmt.Errorf("corpus case array is missing or ambiguous")
	}
	end := bytes.LastIndexByte(old, ']')
	prefix := bytes.TrimRight(old[:end], " \t\r\n")
	if len(prefix) < 2 {
		return nil, fmt.Errorf("registration requires a nonempty baseline corpus")
	}
	item, err := json.MarshalIndent(request.Case, "    ", "  ")
	if err != nil {
		return nil, err
	}
	next := append(bytes.Clone(prefix), []byte(",\n    ")...)
	next = append(next, item...)
	next = append(next, []byte("\n  ]")...)
	result := bytes.Replace(raw, old, next, 1)
	if request.PromoteMetaSource {
		return removeCorpusMetaSource(result, request.Case.Path)
	}
	return result, nil
}

func generateRegistry(source *goSource, request Request) error {
	function, err := source.function("expectedRegistry")
	if err != nil {
		return err
	}
	count := 0
	ast.Inspect(function, func(node ast.Node) bool {
		list, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}
		array, ok := list.Type.(*ast.ArrayType)
		if !ok || array.Len != nil || source.text(array.Elt) != "CaseDefinition" {
			return true
		}
		value := request.Case
		row := fmt.Sprintf("\n{ID: %q, Path: %q, Kind: KindValid, ExpectedDecision: DecisionPass, ProofChoice: %q, MetaOperation: %q, Scope: ScopeLanguageCapability, EntityFields: %t, ImplicitActivityPorts: %t},\n",
			value.ID, value.Path, value.ProofChoice, value.MetaOperation, value.EntityFields, value.ImplicitActivityPorts)
		source.insert(list.Rbrace, row)
		count++
		return false
	})
	if count != 1 {
		return fmt.Errorf("expected exactly one native case registry, got %d", count)
	}
	if request.PromoteMetaSource {
		return removeRegistryMetaSource(source, function, request.Case.Path)
	}
	return nil
}

func removeCorpusMetaSource(raw []byte, path string) ([]byte, error) {
	var document map[string]json.RawMessage
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	old := document["meta_sources"]
	var paths []string
	if len(old) == 0 || bytes.Count(raw, old) != 1 {
		return nil, fmt.Errorf("corpus meta-source array is missing or ambiguous")
	}
	if err := json.Unmarshal(old, &paths); err != nil {
		return nil, err
	}
	remaining, err := removeMetaSource(paths, path)
	if err != nil {
		return nil, err
	}
	next, err := json.MarshalIndent(remaining, "  ", "  ")
	if err != nil {
		return nil, err
	}
	return bytes.Replace(raw, old, next, 1), nil
}

func removeMetaSource(paths []string, target string) ([]string, error) {
	found := -1
	for index, path := range paths {
		if path != target {
			continue
		}
		if found != -1 {
			return nil, fmt.Errorf("meta-source membership is not unique")
		}
		found = index
	}
	if found == -1 {
		return nil, fmt.Errorf("requested meta source is not registered")
	}
	remaining := append([]string{}, paths[:found]...)
	return append(remaining, paths[found+1:]...), nil
}

func removeRegistryMetaSource(source *goSource, function *ast.FuncDecl, path string) error {
	count := 0
	var problem error
	ast.Inspect(function, func(node ast.Node) bool {
		field, ok := node.(*ast.KeyValueExpr)
		if !ok || source.text(field.Key) != "MetaSources" {
			return true
		}
		count++
		list, ok := field.Value.(*ast.CompositeLit)
		if !ok {
			problem = fmt.Errorf("native meta sources are not an explicit literal")
			return false
		}
		paths, err := nativeMetaSourceValues(source, list)
		if err != nil {
			problem = err
			return false
		}
		remaining, err := removeMetaSource(paths, path)
		if err != nil {
			problem = err
			return false
		}
		for index := range remaining {
			remaining[index] = strconv.Quote(remaining[index])
		}
		source.replace(field.Value, "[]string{"+strings.Join(remaining, ", ")+"}")
		return false
	})
	if count != 1 {
		return fmt.Errorf("expected exactly one native meta-source registry, got %d", count)
	}
	return problem
}

func nativeMetaSourceValues(source *goSource, list *ast.CompositeLit) ([]string, error) {
	array, ok := list.Type.(*ast.ArrayType)
	if !ok || array.Len != nil || source.text(array.Elt) != "string" {
		return nil, fmt.Errorf("native meta sources are not a string slice")
	}
	paths := make([]string, 0, len(list.Elts))
	for _, element := range list.Elts {
		literal, ok := element.(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return nil, fmt.Errorf("native meta-source path is not a string literal")
		}
		path, err := strconv.Unquote(literal.Value)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func corpusTotals(raw []byte) (total, valid, capability int, err error) {
	var registry languagesyntax.Registry
	if err = decodeStrict(raw, &registry); err != nil {
		return
	}
	total = len(registry.Cases)
	for _, item := range registry.Cases {
		if item.Kind == languagesyntax.KindValid {
			valid++
		}
		if item.Scope == languagesyntax.ScopeLanguageCapability {
			capability++
		}
	}
	return
}

func generateModel(source *goSource, corpus []byte) error {
	total, valid, capability, err := corpusTotals(corpus)
	if err != nil {
		return err
	}
	expected := map[string]int{"totalCases": total, "validCases": valid, "FixedCapabilityTotal": capability}
	seen := map[string]int{}
	for _, declaration := range source.file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok != token.CONST {
			continue
		}
		for _, spec := range group.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok || len(value.Names) != 1 || len(value.Values) != 1 {
				continue
			}
			name := value.Names[0].Name
			old, wanted := expected[name]
			if !wanted {
				continue
			}
			actual, literal := integer(value.Values[0])
			if !literal || actual != old {
				return fmt.Errorf("baseline syntax denominator mismatch: %s", name)
			}
			seen[name]++
			source.replace(value.Values[0], strconv.Itoa(old+1))
		}
	}
	for _, name := range sortedPaths(expected) {
		if seen[name] != 1 {
			return fmt.Errorf("syntax denominator %s is not unique", name)
		}
	}
	return nil
}
