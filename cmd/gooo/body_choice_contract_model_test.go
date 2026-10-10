package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
)

func cliChoiceContractModel(t *testing.T) string {
	t.Helper()
	m, err := contractdecision.NewChoiceConditioned([contractdecision.ParameterCount]float32{},
		[contractdecision.ChoiceContextParameterCount]float32{}, contractdecision.ExtremePooling)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "choice-contract.json")
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestChoiceContractModelCLIContextAndCodegen(t *testing.T) {
	checkContractModelCLIContextAndCodegen(t, cliChoiceContractModel(t), contractdecision.ChoiceSchema)
}

func TestChoiceContractModelCLINativeConstructionAndModelFreeReplay(t *testing.T) {
	checkContractModelCLINativeConstructionAndReplay(t, cliChoiceContractModel(t), contractdecision.ChoiceSchema)
}
