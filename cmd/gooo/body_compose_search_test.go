package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestBodyComposeCLISourceSearchRunsImmediately(t *testing.T) {
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	var stdout, stderr bytes.Buffer
	args := []string{"body-compose", "--source", "../../examples/body-codegen/source-search-composition.gooo.fixture",
		"--cases", "../../examples/body-codegen/source-search-composition-cases.json", "--go-bin", tool}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatalf("compose source search(%d): %s", code, stderr.String())
	}
	var result bodyCompositionOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Composition.Stage != "COMPLETE" || result.Runtime.FinitePassed != 15 || !result.Runtime.RuntimeReplayed ||
		result.Runtime.ModelCalls != 0 || strings.Contains(result.Composition.GoooSource, "__GOOO_BODY_HOLE_") {
		t.Fatal("source search CLI did not connect generation to native execution", result.Runtime.Stage)
	}
}

func TestBodyRealizeCLISourceSearchProducesReusableCheckpoint(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	source := "../../examples/body-codegen/ir-search-source.gooo.fixture"
	var generated, stderr bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", "ClampNegativeToZero", source}, &generated, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	generation := filepath.Join(t.TempDir(), "generation.json")
	if err := os.WriteFile(generation, generated.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	if code := run([]string{"body-realize", "--source", source, "--generation", generation,
		"--out", filepath.Join(t.TempDir(), "realized")}, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var realized bodycodegen.Realization
	if err := json.Unmarshal(stdout.Bytes(), &realized); err != nil || realized.FinitePassed != 5 || realized.ModelCalls != 0 ||
		strings.Contains(realized.Source, "assembling") || strings.Contains(realized.Source, "__GOOO_BODY_HOLE_") {
		t.Fatal("source search checkpoint was not returned", err, realized)
	}
}
