package syntaxregistration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

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
	return bytes.Replace(raw, old, next, 1), nil
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
	return nil
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
