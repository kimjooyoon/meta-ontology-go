package bodyexecution

import (
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type localSiteBinding struct {
	name, scalar string
	declaration  *bodycodegen.BodyExpressionSource
}

func validateLocalSiteContract(sites []CompositionLocalSite, uses []CompositionLocalUseSite) error {
	bindings := map[conditionOriginKey]localSiteBinding{}
	sourceUses := map[conditionOriginKey]map[[2]int]bool{}
	writeIDs, useIDs := map[conditionOriginKey]bool{}, map[conditionOriginKey]bool{}
	valid := func(root, activity, name, scalar string, variable, ordinal, start, end int) bool {
		return root != "" && activity != "" && name != "" && variable >= 0 && ordinal >= 0 && start >= 0 && end > start && (scalar == "int64" || scalar == "bool" || scalar == "string")
	}
	for _, s := range sites {
		if !valid(s.RootActivityID, s.ActivityID, s.Name, s.ScalarType, s.VariableOrdinal, s.NativeOrdinal, s.Start, s.End) {
			return fmt.Errorf("local write site identity or scalar coordinates differ")
		}
		id := conditionOriginKey{s.RootActivityID, s.ActivityID, s.NativeOrdinal}
		if writeIDs[id] {
			return fmt.Errorf("duplicate local write site identity")
		}
		writeIDs[id] = true
		key := conditionOriginKey{s.RootActivityID, s.ActivityID, s.VariableOrdinal}
		binding := localSiteBinding{s.Name, s.ScalarType, s.Source.Declaration}
		if prior, found := bindings[key]; found && !reflect.DeepEqual(prior, binding) {
			return fmt.Errorf("local lexical binding has contradictory source declarations")
		}
		bindings[key] = binding
	}
	for _, s := range uses {
		if !valid(s.RootActivityID, s.ActivityID, s.Name, s.ScalarType, s.VariableOrdinal, s.NativeOrdinal, s.Start, s.End) {
			return fmt.Errorf("local use site identity or scalar coordinates differ")
		}
		id := conditionOriginKey{s.RootActivityID, s.ActivityID, s.NativeOrdinal}
		if useIDs[id] {
			return fmt.Errorf("duplicate local use site identity")
		}
		useIDs[id] = true
		if s.Source != nil {
			if s.VariableUseOrdinal < 0 {
				return fmt.Errorf("local source-use ordinal differs")
			}
			key := conditionOriginKey{s.RootActivityID, s.ActivityID, 0}
			if sourceUses[key] == nil {
				sourceUses[key] = map[[2]int]bool{}
			}
			sourceID := [2]int{s.VariableOrdinal, s.VariableUseOrdinal}
			if sourceUses[key][sourceID] {
				return fmt.Errorf("duplicate lexical source-use ordinal")
			}
			sourceUses[key][sourceID] = true
			if s.Source.Status == "MAPPED" && (s.Declaration == nil || s.Source.SourceSHA256 != s.Declaration.SourceSHA256 || s.Source.Expression != s.Name || s.Source.Start < 0 || s.Source.End <= s.Source.Start || s.Source.Start == s.Declaration.Start) {
				return fmt.Errorf("local use source differs from its declared carrier")
			}
		}
		binding, found := bindings[conditionOriginKey{s.RootActivityID, s.ActivityID, s.VariableOrdinal}]
		if !found || !reflect.DeepEqual(binding, localSiteBinding{s.Name, s.ScalarType, s.Declaration}) {
			return fmt.Errorf("local use differs from its declared lexical binding")
		}
	}
	return nil
}
