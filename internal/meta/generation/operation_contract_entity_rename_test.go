package generation

import (
	"reflect"
	"strings"
	"testing"
)

func TestNativeContractEntityRenamePreservesRelations(t *testing.T) {
	raw := string(operationContractFixture(t))
	before, err := parseOperationInputContract([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	for index, entity := range before.Entities {
		name := entity.Name
		t.Run(name, func(t *testing.T) {
			renamed := strings.ReplaceAll(raw, name, "Renamed"+name)
			after, err := parseOperationInputContract([]byte(renamed))
			if err != nil {
				t.Fatalf("same-ID source rename did not reach native admission: %v", err)
			}
			if after.Entities[index] != (contractEntityReference{ID: entity.ID, Name: "Renamed" + name}) {
				t.Fatal("admitted entity lost its authored label or stable identity")
			}
			if before.SourceDigest == after.SourceDigest || before.SemanticDigest != after.SemanticDigest ||
				!reflect.DeepEqual(before.Bindings, after.Bindings) || !reflect.DeepEqual(before.Facts, after.Facts) ||
				!reflect.DeepEqual(before.ObligationFacts, after.ObligationFacts) ||
				!reflect.DeepEqual(before.PolicyFacts, after.PolicyFacts) {
				t.Fatal("entity spelling changed the admitted native relations")
			}
		})
	}
}

func TestNativeContractRefusesChangedEntityIdentityWithSameLabels(t *testing.T) {
	raw := string(operationContractFixture(t))
	function := "gooo://meta-operation-input-contract/entity/function-input"
	file := "gooo://meta-operation-input-contract/entity/file-input"
	for name, source := range map[string]string{
		"unknown ID":   strings.ReplaceAll(raw, function, function+"-unknown"),
		"duplicate ID": strings.ReplaceAll(raw, function, file),
		"swapped IDs":  strings.NewReplacer(function, file, file, function).Replace(raw),
		"missing ID":   strings.ReplaceAll(raw, ` id "`+function+`"`, ""),
		"wrong kind":   strings.Replace(raw, "entity FunctionSource", "agent FunctionSource", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if source == raw {
				t.Fatal("rejection fixture did not mutate the source")
			}
			got, err := parseOperationInputContract([]byte(source))
			if err == nil || !reflect.DeepEqual(got, operationInputContract{}) {
				t.Fatalf("invalid identity produced a partial or accepted contract: %v", err)
			}
		})
	}
}
