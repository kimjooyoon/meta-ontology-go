package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/conditiondecision"
	"github.com/kimjooyoon/gooo-decision-runtime/executiondecision"
	"github.com/kimjooyoon/gooo-decision-runtime/flowdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func cliVersionedCandidateModel(t *testing.T, version string) string {
	t.Helper()
	var raw []byte
	var err error
	if version == decision.ExecutionFlowFeatureVersion || version == decision.SemanticFlowFeatureVersion {
		m, failure := flowdecision.NewForFeatures([flowdecision.ParameterCount]float32{}, version)
		if failure != nil {
			t.Fatal(failure)
		}
		raw, err = m.Marshal()
	} else if version == decision.ExecutionFeatureVersion {
		m, failure := executiondecision.New([executiondecision.ParameterCount]float32{})
		if failure != nil {
			t.Fatal(failure)
		}
		raw, err = m.Marshal()
	} else {
		m, failure := conditiondecision.NewForFeatures([conditiondecision.ParameterCount]float32{}, version)
		if failure != nil {
			t.Fatal(failure)
		}
		raw, err = m.Marshal()
	}
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "model.json")
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestCandidateModelCLIExplicitVersionAndArrays(t *testing.T) {
	for _, version := range []string{decision.ConditionChannelFeatureVersion, decision.ConditionBranchFeatureVersion, decision.ExecutionFeatureVersion, decision.ExecutionFlowFeatureVersion, decision.SemanticFlowFeatureVersion} {
		t.Run(version, func(t *testing.T) {
			model := cliVersionedCandidateModel(t, version)
			source := "../../examples/body-codegen/source-output-feedback.gooo.fixture"
			var out, diagnostics bytes.Buffer
			if code := run([]string{"body-context", "--model", model, "--feature-version", version, "--activity", "Choose", source}, &out, &diagnostics); code != exitOK {
				t.Fatal(code, out.String(), diagnostics.String())
			}
			var before bodycodegen.TypedPathContextExport
			if err := json.Unmarshal(out.Bytes(), &before); err != nil {
				t.Fatal(err)
			}
			if before.ModelCompatibility.Model.FeatureVersion != version || before.ModelPredictions != 0 || before.CandidateTests != 0 || len(before.Inputs) != 2 {
				t.Fatal("CLI did not bind exact artifact version", before)
			}
			out.Reset()
			diagnostics.Reset()
			if code := run([]string{"body-codegen", "--json", "--activity", "Choose", "--path-model", model, "--path-step-attempts", "1", "--path-feedback-rounds", "3", source}, &out, &diagnostics); code != exitOK {
				t.Fatal(code, out.String(), diagnostics.String())
			}
			var result bodycodegen.Result
			if err := json.Unmarshal(out.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			p := result.Report.BodyPaths
			if p.FunctionalCompleteness != 100 || p.Conditions.Passed != 3 {
				t.Fatal("versioned model bypassed codegen validation")
			}
			for i, input := range before.Inputs {
				legacy := version == decision.ConditionChannelFeatureVersion || version == decision.ConditionBranchFeatureVersion
				if input.InputSHA != "sha256:"+p.ConditionProgress[0].Ranking.FeatureSHA[i] ||
					(input.ExecutionFeatures != nil) != (version == decision.ExecutionFeatureVersion) ||
					(input.FlowFeatures != nil) != (version == decision.ExecutionFlowFeatureVersion || version == decision.SemanticFlowFeatureVersion) || (input.Features != nil) != legacy {
					t.Fatal("preflight array or digest differs", i)
				}
			}
		})
	}
}

func TestExecutionModelCLIConstructAndReplayWithoutArtifact(t *testing.T) {
	for _, version := range []string{decision.ExecutionFeatureVersion, decision.ExecutionFlowFeatureVersion, decision.SemanticFlowFeatureVersion} {
		t.Run(version, func(t *testing.T) { candidateModelCLIConstructAndReplay(t, version) })
	}
}

func candidateModelCLIConstructAndReplay(t *testing.T, version string) {
	t.Helper()
	model := cliVersionedCandidateModel(t, version)
	root := filepath.Join(t.TempDir(), "construction")
	base := "../../examples/body-codegen/source-output-"
	goBin := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goBin += ".exe"
	}
	args := []string{"body-construct", "--source", base + "feedback.gooo.fixture", "--entry", "Main", "--model", model,
		"--construction-cases", base + "construction-cases.json", "--cases", base + "evaluation-cases.json", "--attempts", "4", "--go-bin", goBin, "--out", root}
	var out, diagnostics bytes.Buffer
	if code := run(args, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var result bodyConstructOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	p := result.Construction.Initial.Preparations[0].Generation.Report.BodyPaths
	if result.Evaluation.Runtime.FinitePassed != 8 || p.Conditions.Passed != 3 || p.Search.Selection.ModelCalls != 1 || p.ModelRetention.FeatureVersion != version {
		t.Fatal("construction did not retain execution model identity", result)
	}
	if err := os.Remove(model); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diagnostics.Reset()
	args = []string{"body-construct", "--source", filepath.Join(root, "original.gooo"), "--construction", filepath.Join(root, "construction.json"), "--cases", base + "evaluation-cases.json", "--go-bin", goBin}
	if code := run(args, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.GeneratedNow || !result.Evaluation.ConstructionReplayed || result.Evaluation.NewModelCalls != 0 || result.Evaluation.Runtime.FinitePassed != 8 {
		t.Fatal("saved execution-model construction requires artifact", err, result)
	}
}
