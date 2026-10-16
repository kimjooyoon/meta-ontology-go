package syntaxregistration

import (
	"fmt"
	"go/ast"
	"reflect"
)

func bindProfileMetaSources(source *goSource, expected []string) error {
	count := 0
	for _, path := range expected {
		if path == profileContractPath {
			count++
		}
	}
	if count != 1 {
		return profileIssue("UNKNOWN", "bind-contract", "PROFILE_CONTRACT_MEMBERSHIP_MISSING_OR_AMBIGUOUS")
	}
	root, err := source.function("expectedRegistry")
	if err != nil {
		return err
	}
	if root.Body == nil || len(root.Body.List) != 1 {
		return fmt.Errorf("PROFILE_NATIVE_REGISTRY_UNRESOLVED")
	}
	returned, ok := root.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(returned.Results) != 1 {
		return fmt.Errorf("PROFILE_NATIVE_REGISTRY_UNRESOLVED")
	}
	registry, ok := returned.Results[0].(*ast.CompositeLit)
	if !ok {
		return fmt.Errorf("PROFILE_NATIVE_REGISTRY_UNRESOLVED")
	}
	seen := 0
	for _, element := range registry.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok || !namedRegistryType(pair.Key, "MetaSources") {
			continue
		}
		seen++
		list, ok := pair.Value.(*ast.CompositeLit)
		if !ok {
			return fmt.Errorf("PROFILE_NATIVE_META_SOURCES_UNRESOLVED")
		}
		paths, err := nativeMetaSourceValues(source, list)
		if err != nil || !reflect.DeepEqual(paths, expected) {
			return fmt.Errorf("PROFILE_NATIVE_META_SOURCES_DIVERGED")
		}
	}
	if seen != 1 {
		return fmt.Errorf("PROFILE_NATIVE_META_SOURCES_UNRESOLVED")
	}
	return nil
}
