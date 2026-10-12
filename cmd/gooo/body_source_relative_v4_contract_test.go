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

func freshV4Fixture(t *testing.T) (string, string) {
	t.Helper()
	source := "../../examples/body-codegen/branch-owned-output-v4.gooo.fixture"
	model := filepath.Join(t.TempDir(), "zero-v4.json")
	m, err := contractdecision.NewSourceRelativeV4([contractdecision.InteractionRequirementParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(model, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return source, model
}

// New authored cases and zero weights exercise wiring, not fitted-model quality.
func TestSourceRelativeV4CompilerContextAndFiniteConstruction(t *testing.T) {
	source, model := freshV4Fixture(t)
	call := func(args ...string) []byte {
		t.Helper()
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitOK {
			t.Fatal(code, diagnostics.String(), out.String())
		}
		return out.Bytes()
	}
	var before bodycodegen.TypedPathContextExport
	if err := json.Unmarshal(call("body-context", "--model", model, "--activity", "Charge", source), &before); err != nil {
		t.Fatal(err)
	}
	if before.Context == nil || before.Context.Schema != "gooo/compiler-source-relative-contract-model-context/v4" ||
		before.Context.FeatureVersion != contractdecision.SourceRelativeV4FeatureVersion ||
		before.Context.IntentPolicy != contractdecision.SourceRelativeV3ProseNeutralIntentPolicy ||
		before.Context.ClearedIntentCells == nil || *before.Context.ClearedIntentCells != ([2]int{64, 192}) ||
		before.ContractCases == nil || before.ContractCases.Count != 6 || before.ContractConditions == nil ||
		before.ContractConditions.Count != 3 || before.ModelPredictions != 0 || before.CandidateTests != 0 ||
		before.SelectedEmission || len(before.Inputs) != 2 {
		t.Fatal("V4 preflight lost source requirements, input policy or zero-call boundary")
	}
	for _, row := range before.Inputs {
		owned := row.BranchOutputs
		if row.SourceRelativeFeatures == nil || row.SourceRelativeOutputs == nil || owned == nil ||
			!owned.Available || owned.CaseCount != 6 || owned.CaseSHA256 != row.SourceRelativeOutputs.CaseSHA256 ||
			len(owned.Literals[0]) != 1 || len(owned.Literals[1]) != 1 ||
			owned.Literals[0][0] != 23 || owned.Literals[1][0] != 9 {
			t.Fatal("then/else literal ownership or all output rows lost", owned)
		}
		for slot, value := range owned.Features {
			if row.SourceRelativeFeatures[256+slot] != value {
				t.Fatal("branch explanation differs from actual array", slot)
			}
		}
		for _, value := range row.SourceRelativeFeatures[64:192] {
			if value != 0 {
				t.Fatal("V4 intent channel is not empty")
			}
		}
	}
	var result bodycodegen.Result
	if err := json.Unmarshal(call("body-codegen", "--json", "--activity", "Charge", "--path-model", model,
		"--path-step-attempts", "1", source), &result); err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if p == nil || p.ContractRanking == nil || p.ContractRanking.Calls != 1 || p.FunctionalCompleteness != 100 ||
		p.Conditions.Passed != 3 || p.ModelContext == nil || p.ModelContext.FeatureVersion != before.Context.FeatureVersion ||
		p.ContractRanking.Schema != "gooo/source-relative-branch-output-path-ranking/v4" {
		t.Fatal("V4 finite construction or one initial prediction differs")
	}
	for i, row := range before.Inputs {
		if row.InputSHA != "sha256:"+p.ContractRanking.FeatureSHA[i] {
			t.Fatal("preflight array differs from model input", i)
		}
	}
	result = bodycodegen.Result{}
	if err := json.Unmarshal(call("body-codegen", "--json", "--activity", "Charge", source), &result); err != nil {
		t.Fatal(err)
	}
	if result.Report.BodyPaths == nil || result.Report.BodyPaths.ContractRanking != nil ||
		result.Report.BodyPaths.FunctionalCompleteness != 100 {
		t.Fatal("model omission changed deterministic construction")
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{"body-context", "--model", model, "--feature-version",
		contractdecision.SourceRelativeV3FeatureVersion, "--activity", "Charge", source}, &out, &diagnostics); code == exitOK {
		t.Fatal("V4 artifact accepted an explicit V3 input contract")
	}
}
