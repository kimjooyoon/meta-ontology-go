package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
)

func cliCanonicalContractModel(t *testing.T) string {
	t.Helper()
	model, err := contractdecision.NewCanonicalInteractionRequirementConditioned(
		[contractdecision.InteractionRequirementParameterCount]float32{}, contractdecision.MeanPooling)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := model.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "canonical-contract.json")
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestCanonicalContractCLIContextCodegenNativeConstructionAndReplay(t *testing.T) {
	for _, source := range []string{interactionContractSource, "../../examples/body-codegen/source-interaction-assignment-cases.gooo.fixture"} {
		model := cliCanonicalContractModel(t)
		checkContractModelCLIContextAndCodegenSource(t, model, contractdecision.CanonicalInteractionRequirementSchema, source)
		checkContractModelCLINativeConstructionAndReplaySource(t, model, contractdecision.CanonicalInteractionRequirementSchema, source)
	}
}

func TestCanonicalContractUnsupportedCLIChoiceUsesDeterministicSearch(t *testing.T) {
	checkOrderedContractUnsupportedCLIChoice(t, cliCanonicalContractModel(t))
}
