package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodypathstream"
)

func TestBodyPathStreamCLIRecipe(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/path-recipe.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	recipe, err := os.ReadFile("../../examples/body-codegen/path-recipe.json")
	if err != nil {
		t.Fatal(err)
	}
	request, err := json.Marshal(bodypathstream.Request{Schema: bodypathstream.RequestSchema,
		CorrelationID: "recipe", Source: string(source), Activity: "Compose", Document: recipe})
	if err != nil {
		t.Fatal(err)
	}
	output, err := os.Create(filepath.Join(t.TempDir(), "results.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	var diagnostics bytes.Buffer
	code := runWithInput([]string{"body-path-stream", "--workers", "1"},
		io.NopCloser(bytes.NewReader(append(request, '\n'))), output, &diagnostics)
	if code != exitOK {
		t.Fatalf("code=%d diagnostics=%s", code, &diagnostics)
	}
	if _, err := output.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	var result bodypathstream.Result
	if err := json.NewDecoder(output).Decode(&result); err != nil || result.Status != "completed" ||
		result.CorrelationID != "recipe" || result.Response == nil || !json.Valid(bytes.TrimSpace(diagnostics.Bytes())) {
		t.Fatalf("result=%+v err=%v diagnostics=%s", result, err, &diagnostics)
	}
}

func TestBodyPathStreamCLIHelp(t *testing.T) {
	var out, diagnostics bytes.Buffer
	if code := runWithInput([]string{"body-path-stream", "--help"}, nil, &out, &diagnostics); code != exitOK || !bytes.Contains(diagnostics.Bytes(), []byte("gooo body-path-stream")) || out.Len() != 0 {
		t.Fatalf("code=%d diagnostics=%s", code, &diagnostics)
	}
}
