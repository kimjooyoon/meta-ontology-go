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
		binding, found := bindings[conditionOriginKey{s.RootActivityID, s.ActivityID, s.VariableOrdinal}]
		if !found || !reflect.DeepEqual(binding, localSiteBinding{s.Name, s.ScalarType, s.Declaration}) {
			return fmt.Errorf("local use differs from its declared lexical binding")
		}
	}
	return nil
}
