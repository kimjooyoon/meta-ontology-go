package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestBodyContextOwnModelExplainsTypedDeclineBeforeGeneration(t *testing.T) {
	_, model := recordModelContextFixture(t)
	source, err := os.ReadFile("../../examples/caller-typed-paths/unary.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	args := []string{"--model", model, "--activity", "Choose", "--include-plan", "f.gooo"}
	code := runBodyContext(args, mapSourceReader{"f.gooo": source}, &out, &stderr)
	var result struct {
		Compatibility struct {
			Status string `json:"status"`
			Reason string `json:"reason"`
		} `json:"model_compatibility"`
		ModelPredictions int             `json:"model_predictions"`
		CandidateTests   int             `json:"candidate_tests"`
		Inputs           []any           `json:"inputs"`
		ExpandedPlan     json.RawMessage `json:"expanded_plan"`
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if code != exitOK || stderr.Len() != 0 || result.Compatibility.Status != "DECLINED_TO_DETERMINISTIC" ||
		result.Compatibility.Reason != "THREE_DECISION_COUNT_UNSUPPORTED" || result.ModelPredictions != 0 ||
		result.CandidateTests != 0 || len(result.Inputs) != 0 || !strings.Contains(string(result.ExpandedPlan), "branch_layout") {
		t.Fatal("typed model inspection must retain its native decline and source choices", out.String(), stderr.String())
	}
}

func TestBodyContextTypedModelSelectsFeatureAndReadsExternalPlan(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-compound.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := os.ReadFile("../../examples/body-codegen/typed-path-compound-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	model := writeCLIPathFeedbackModel(t)
	raw, _ := os.ReadFile(model)
	var metadata decision.Metadata
	if err := json.Unmarshal(raw, &metadata); err != nil {
		t.Fatal(err)
	}
	metadata.FeatureVersion = decision.SemanticContextIntentFeatureVersion
	raw, _ = json.Marshal(metadata)
	if err := os.WriteFile(model, raw, 0600); err != nil {
		t.Fatal(err)
	}
	reader := mapSourceReader{"f.gooo": source, "p.json": plan}
	args := []string{"--model", model, "--plan", "p.json", "--activity", "Combined", "f.gooo"}
	var out, stderr bytes.Buffer
	if runBodyContext(args, reader, &out, &stderr) != exitOK {
		t.Fatal(out.String(), stderr.String())
	}
	var got bodycodegen.TypedPathContextExport
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ModelCompatibility == nil || got.ModelCompatibility.Status != "READY_FOR_RANKING" ||
		got.Context.FeatureVersion != decision.SemanticContextIntentFeatureVersion || len(got.Inputs) != 3 ||
		got.ModelPredictions != 0 || got.CandidateTests != 0 || strings.Contains(out.String(), "expanded_plan") {
		t.Fatal("CLI did not select the model's feature contract", out.String())
	}
	out.Reset()
	args = append([]string{"--feature-version", decision.SplitContextIntentFeatureVersion}, args...)
	if runBodyContext(args, reader, &out, &stderr) != exitFailure || strings.Contains(out.String(), "model_compatibility") ||
		!strings.Contains(out.String(), "explicit feature version differs") {
		t.Fatal("mismatched feature advertised compatibility", out.String())
	}
}
