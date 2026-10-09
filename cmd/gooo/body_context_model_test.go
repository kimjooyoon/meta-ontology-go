package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type testedRecordModelContext struct {
	bodycodegen.RecordAssemblyContextExport
	Compatibility *struct {
		Status string                        `json:"status"`
		Reason string                        `json:"reason"`
		Model  bodycodegen.RetainedModelInfo `json:"model"`
	} `json:"model_compatibility"`
}

func recordModelContextFixture(t *testing.T) ([]byte, string) {
	t.Helper()
	source, err := os.ReadFile("../../examples/scalar-identity/source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	model, err := filepath.Abs("../../examples/scalar-identity/model/model.json")
	if err != nil {
		t.Fatal(err)
	}
	return source, model
}

func TestBodyContextModelSelectsItsInputContractWithoutPredicting(t *testing.T) {
	source, model := recordModelContextFixture(t)
	var out, stderr bytes.Buffer
	args := []string{"--activity", "Describe", "--model", model, "r.gooo"}
	if runBodyContext(args, mapSourceReader{"r.gooo": source}, &out, &stderr) != exitOK {
		t.Fatal(out.String(), stderr.String())
	}
	var got testedRecordModelContext
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Compatibility == nil || got.Compatibility.Status != "READY_FOR_RANKING" || got.Compatibility.Reason != "SOURCE_INPUT_ENCODED" ||
		!got.Compatibility.Model.Loaded || got.Compatibility.Model.WeightsSHA256 != "76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f" ||
		got.Compatibility.Model.FeatureVersion != jointdecision.RecordGraphSharedFeatureVersion || got.Context.FeatureVersion != jointdecision.RecordGraphSharedFeatureVersion ||
		got.ModelPredictions != 0 || got.CandidateTests != 0 || got.ExpandedPlan != nil || got.ValueFlow != nil || stderr.Len() != 0 {
		t.Fatal("model preflight identity or zero-work boundary differs", out.String(), stderr.String())
	}
	if _, err := jointdecision.DecodeRecordGraphThree(got.Context.Text); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got.Context.Text, "hello!") {
		t.Fatal("case expectation leaked into model input")
	}
}

func TestBodyContextModelReportsRepresentationDecline(t *testing.T) {
	source, model := recordModelContextFixture(t)
	var lines []string
	for _, line := range strings.Split(string(source), "\n") {
		if !strings.Contains(line, `choice "active"`) && !strings.Contains(line, `choice "count"`) {
			lines = append(lines, line)
		}
	}
	var out, stderr bytes.Buffer
	if runBodyContext([]string{"--model", model, "--activity", "Describe", "r.gooo"},
		mapSourceReader{"r.gooo": []byte(strings.Join(lines, "\n"))}, &out, &stderr) != exitOK {
		t.Fatal(out.String(), stderr.String())
	}
	var got testedRecordModelContext
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Compatibility == nil || got.Compatibility.Status != "DECLINED_TO_DETERMINISTIC" ||
		got.Compatibility.Reason != "THREE_FIELD_CHOICES_REQUIRED" || got.Context.Text != "" || got.ModelPredictions != 0 || got.CandidateTests != 0 {
		t.Fatal("decline must retain its reason and zero predictions", out.String())
	}
}

func TestBodyContextModelRejectsMismatchedFeatureAndMissingModel(t *testing.T) {
	source, model := recordModelContextFixture(t)
	for _, flags := range [][]string{
		{"--model", model, "--feature-version", jointdecision.RecordSharedFeatureVersion}, {"--model", model + ".missing"},
	} {
		var out, stderr bytes.Buffer
		args := append(flags, "--activity", "Describe", "r.gooo")
		if runBodyContext(args, mapSourceReader{"r.gooo": source}, &out, &stderr) != exitFailure ||
			!strings.Contains(out.String(), `"status":"FAIL_CLOSED"`) || strings.Contains(out.String(), `"model_compatibility"`) {
			t.Fatal("invalid model input advertised as compatible", out.String(), stderr.String())
		}
	}
	for _, flags := range [][]string{{"--model"}, {"--model", model, "--model", model}} {
		var out, stderr bytes.Buffer
		if runBodyContext(append(flags, "--activity", "Describe", "r.gooo"), mapSourceReader{}, &out, &stderr) != exitUsage || out.Len() != 0 {
			t.Fatal("ambiguous model flags accepted")
		}
	}
}

func TestBodyContextModelPreservesExplicitFlowAndCancellation(t *testing.T) {
	source, model := recordModelContextFixture(t)
	args := []string{"--model", model, "--activity", "Describe", "--value-flow", "--feature-version", jointdecision.RecordGraphSharedFeatureVersion, "r.gooo"}
	var out, stderr bytes.Buffer
	if runBodyContext(args, mapSourceReader{"r.gooo": source}, &out, &stderr) != exitOK {
		t.Fatal(out.String(), stderr.String())
	}
	var got testedRecordModelContext
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ValueFlow == nil || got.ValueFlow.Status != "RESOLVED" || got.Context.ValueFlowSHA256 == "" {
		t.Fatal("flow missing")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out.Reset()
	if runBodyContextWithContext(ctx, args, mapSourceReader{"r.gooo": source}, &out, &stderr) != exitFailure || strings.Contains(out.String(), `"model_compatibility"`) {
		t.Fatal("cancelled preflight advertised compatibility", out.String())
	}
}

func TestBodyContextModelRejectsOtherAssemblyKinds(t *testing.T) {
	_, model := recordModelContextFixture(t)
	source, err := os.ReadFile("../../examples/body-codegen/path-recipe.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	if runBodyContext([]string{"--model", model, "--activity", "Compose", "r.gooo"},
		mapSourceReader{"r.gooo": source}, &out, &stderr) != exitFailure ||
		!strings.Contains(out.String(), "body-context --model requires record source assembly") ||
		strings.Contains(out.String(), `"model_compatibility"`) {
		t.Fatal("unsupported body advertised as compatible", out.String(), stderr.String())
	}
}
