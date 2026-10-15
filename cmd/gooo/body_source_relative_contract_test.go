package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
)

func cliSourceRelativeContractModel(t *testing.T) string {
	t.Helper()
	model, err := contractdecision.NewSourceRelativeRequirementConditioned(
		[contractdecision.InteractionRequirementParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := model.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "source-relative-contract.json")
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestSourceRelativeContractCLIContextCodegenNativeConstructionAndReplay(t *testing.T) {
	for _, source := range []string{interactionContractSource, "../../examples/body-codegen/source-interaction-assignment-cases.gooo.fixture"} {
		model := cliSourceRelativeContractModel(t)
		checkContractModelCLIContextAndCodegenSource(t, model, contractdecision.SourceRelativeRequirementSchema, source)
		checkContractModelCLINativeConstructionAndReplaySource(t, model, contractdecision.SourceRelativeRequirementSchema, source)
	}
}

func TestSourceRelativeContractUnsupportedCLIChoiceUsesDeterministicSearch(t *testing.T) {
	checkOrderedContractUnsupportedCLIChoice(t, cliSourceRelativeContractModel(t))
}
