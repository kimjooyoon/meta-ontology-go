package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestBodyContextCLIExportsExplicitInputs(t *testing.T) {
	source, _ := os.ReadFile("../../examples/body-codegen/typed-path-compound.gooo.fixture")
	plan, _ := os.ReadFile("../../examples/body-codegen/typed-path-compound-plan.json")
	reader := mapSourceReader{"fixture.gooo": source, "plan.json": plan}
	args := []string{"--plan", "plan.json", "--activity", "Combined", "fixture.gooo"}
	t.Setenv("GOOO_LAYA_URL", "http://invalid.example/never")
	var stdout, stderr bytes.Buffer
	code := runBodyContext(args, reader, &stdout, &stderr)
	var exported bodycodegen.TypedPathContextExport
	if err := json.Unmarshal(stdout.Bytes(), &exported); err != nil {
		t.Fatal(err)
	}
	if code != exitOK || stderr.Len() != 0 || len(exported.Inputs) != 3 || !exported.SourceBinding.Equivalent ||
		exported.ModelPredictions != 0 || exported.Context.Schema != "gooo/compiler-typed-path-context/v2" {
		t.Fatalf("context command failed: %s %s", stdout.String(), stderr.String())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	stdout.Reset()
	if runBodyContextWithContext(ctx, args, reader, &stdout, &stderr) != exitFailure ||
		!strings.Contains(stdout.String(), `"status":"FAIL_CLOSED"`) || strings.Contains(stdout.String(), `"text":`) {
		t.Fatal("cancelled request exported text")
	}
	stdout.Reset()
	stderr.Reset()
	if run([]string{"body-context"}, &stdout, &stderr) != exitUsage || !strings.Contains(stderr.String(), bodyContextUsage) {
		t.Fatal("command not dispatched")
	}
}

func TestBodyContextCLIRejectsAmbiguousFlags(t *testing.T) {
	for _, flags := range [][]string{
		{}, {"--json"}, {"--plan"}, {"--plan", "p", "--plan", "q"},
		{"--activity", "A", "--activity", "B"}, {"--path-model", "m"}, {"one", "two"},
	} {
		var out, stderr bytes.Buffer
		if runBodyContext(flags, mapSourceReader{}, &out, &stderr) != exitUsage || out.Len() != 0 {
			t.Fatalf("ambiguous flags accepted: %v", flags)
		}
	}
}
