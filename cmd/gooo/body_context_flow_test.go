package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestBodyContextCLIExportsOptionalSourceValueFlow(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/record-field-updates.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	args := []string{"--value-flow", "--activity", "Select", "r.gooo"}
	if runBodyContext(args, mapSourceReader{"r.gooo": source}, &out, &stderr) != exitOK {
		t.Fatal(out.String(), stderr.String())
	}
	var result bodycodegen.RecordAssemblyContextExport
	if err = json.Unmarshal(out.Bytes(), &result); err != nil || result.ValueFlow == nil || result.ValueFlow.Status != "RESOLVED" ||
		result.ModelPredictions != 0 || result.CandidateTests != 0 || result.ExpandedPlan != nil {
		t.Fatal("optional source observation differs", err)
	}
	out.Reset()
	originArgs := append([]string{"--feature-version", jointdecision.RecordOriginSharedFeatureVersion}, args...)
	if runBodyContext(originArgs, mapSourceReader{"r.gooo": source}, &out, &stderr) != exitOK {
		t.Fatal("origin context option failed", out.String())
	}
	if err = json.Unmarshal(out.Bytes(), &result); err != nil || result.Context.Status != "ENCODED" || result.Context.ValueFlowSHA256 == "" {
		t.Fatal("CLI did not expose the new real model input", err)
	}
	out.Reset()
	graphArgs := append([]string{"--feature-version", jointdecision.RecordGraphSharedFeatureVersion}, args...)
	if runBodyContext(graphArgs, mapSourceReader{"r.gooo": source}, &out, &stderr) != exitOK {
		t.Fatal("graph context option failed", out.String())
	}
	if err = json.Unmarshal(out.Bytes(), &result); err != nil || result.Context.Status != "ENCODED" ||
		result.Context.FeatureVersion != jointdecision.RecordGraphSharedFeatureVersion || result.ModelPredictions != 0 || result.CandidateTests != 0 {
		t.Fatal("CLI did not expose source graph input without predictions", err)
	}
	if _, err = jointdecision.DecodeRecordGraphThree(result.Context.Text); err != nil {
		t.Fatal("CLI graph input is not consumable by SDK", err)
	}
	out.Reset()
	if runBodyContext(append(args, "--value-flow"), mapSourceReader{}, &out, &stderr) != exitUsage || out.Len() != 0 {
		t.Fatal("duplicate value-flow accepted")
	}
	args = []string{"--value-flow", "--activity", "Compose", "scalar.gooo"}
	source, _ = os.ReadFile("../../examples/body-codegen/path-recipe.gooo.fixture")
	if runBodyContext(args, mapSourceReader{"scalar.gooo": source}, &out, &stderr) != exitFailure ||
		!strings.Contains(out.String(), "value flow requires record source assembly") {
		t.Fatal("scalar body accepted record observation", out.String())
	}
}
