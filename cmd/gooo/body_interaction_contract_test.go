package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const interactionContractSource = "../../examples/body-codegen/source-interaction-condition-cases.gooo.fixture"

func cliInteractionContractModel(t *testing.T) string {
	t.Helper()
	m, err := contractdecision.NewInteractionRequirementConditioned(
		[contractdecision.InteractionRequirementParameterCount]float32{}, contractdecision.ExtremePooling)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "interaction-contract.json")
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestInteractionContractModelCLIContextAndCodegen(t *testing.T) {
	checkContractModelCLIContextAndCodegenSource(t, cliInteractionContractModel(t), contractdecision.InteractionRequirementSchema, interactionContractSource)
}

func TestInteractionContractModelCLINativeConstructionAndModelFreeReplay(t *testing.T) {
	checkContractModelCLINativeConstructionAndReplaySource(t, cliInteractionContractModel(t), contractdecision.InteractionRequirementSchema, interactionContractSource)
}

func TestInteractionContractUnsupportedCLIChoiceUsesDeterministicSearch(t *testing.T) {
	model := cliInteractionContractModel(t)
	source := "../../examples/body-codegen/source-condition-cases.gooo.fixture"
	var out, diagnostics bytes.Buffer
	if code := run([]string{"body-context", "--model", model, "--activity", "Choose", source}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var before bodycodegen.TypedPathContextExport
	if err := json.Unmarshal(out.Bytes(), &before); err != nil {
		t.Fatal(err)
	}
	if before.ModelCompatibility.Status != "DECLINED_TO_DETERMINISTIC" || before.ModelPredictions != 0 ||
		len(before.Inputs) != 0 || before.ContractConditions != nil {
		t.Fatal("unsupported arithmetic choice advertised a complete input")
	}
	out.Reset()
	diagnostics.Reset()
	if code := run([]string{"body-codegen", "--json", "--path-model", model, "--activity", "Choose", source}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var result bodycodegen.Result
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if p.FunctionalCompleteness != 100 || p.Search.Selection.ModelCalls != 0 || p.ContractRanking != nil ||
		p.ModelContext.Status != "DECLINED_TO_DETERMINISTIC" {
		t.Fatal("unsupported arithmetic choice changed the deterministic continuation")
	}
}
