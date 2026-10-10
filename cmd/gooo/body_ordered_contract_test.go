package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
)

// Zero weights check the CLI route and replay, not learned model quality.
func cliOrderedContractModel(t *testing.T) string {
	t.Helper()
	model, err := contractdecision.NewOrderedRequirementConditioned(
		[contractdecision.OrderedRequirementParameterCount]float32{}, contractdecision.ExtremePooling)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := model.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "ordered-contract.json")
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestOrderedContractCLIContextNativeConstructionAndReplay(t *testing.T) {
	for _, source := range []string{interactionContractSource,
		"../../examples/body-codegen/source-interaction-assignment-cases.gooo.fixture"} {
		t.Run(filepath.Base(source), func(t *testing.T) {
			model := cliOrderedContractModel(t)
			checkContractModelCLIContextAndCodegenSource(t, model, contractdecision.OrderedRequirementSchema, source)
			checkContractModelCLINativeConstructionAndReplaySource(t, model, contractdecision.OrderedRequirementSchema, source)
		})
	}
}

func TestOrderedContractUnsupportedCLIChoiceUsesDeterministicSearch(t *testing.T) {
	checkOrderedContractUnsupportedCLIChoice(t, cliOrderedContractModel(t))
}
