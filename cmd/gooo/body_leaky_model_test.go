package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/flowdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestLeakyModelCLIConstructAndReplay(t *testing.T) {
	m, err := flowdecision.NewForActivation([flowdecision.ParameterCount]float32{}, decision.RelationalFlowFeatureVersion, flowdecision.LeakyReLUActivation)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := m.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "leaky.json")
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	args := []string{"body-context", "--model", name, "--feature-version", decision.RelationalFlowFeatureVersion,
		"--activity", "Choose", "../../examples/body-codegen/source-output-feedback.gooo.fixture"}
	if code := run(args, &out, &diagnostics); code != exitOK {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	var before bodycodegen.TypedPathContextExport
	if err := json.Unmarshal(out.Bytes(), &before); err != nil {
		t.Fatal(err)
	}
	if before.ModelCompatibility.Model.ModelSchema != flowdecision.ActivationSchema || before.ModelPredictions != 0 || before.CandidateTests != 0 {
		t.Fatal("explicit computation missing from CLI preflight")
	}
	candidateModelCLIConstructAndReplayWithModel(t, decision.RelationalFlowFeatureVersion, name)
}
