package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestBodyRealizeCLIProducesReusableSourceAndRetainsInputs(t *testing.T) {
	source := "../../examples/body-codegen/source-assembly.gooo.fixture"
	var generated, diagnostics bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", "Qualified", source}, &generated, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	root := t.TempDir()
	generation := filepath.Join(root, "generation.json")
	if err := os.WriteFile(generation, generated.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "realized")
	args := []string{"body-realize", "--source", source, "--generation", generation, "--out", output}
	var stdout bytes.Buffer
	if code := run(args, &stdout, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var realization bodycodegen.Realization
	if err := json.Unmarshal(stdout.Bytes(), &realization); err != nil || realization.ModelCalls != 0 ||
		realization.FinitePassed != 5 || realization.FiniteTotal != 5 {
		t.Fatal(err, realization)
	}
	for _, name := range []string{"original.gooo", "generation.json", "realized.gooo", "realization.json"} {
		data, err := os.ReadFile(filepath.Join(output, name))
		if err != nil || len(data) == 0 {
			t.Fatal("realization lost an input or result", name, err)
		}
		if name == "generation.json" && !bytes.Equal(data, generated.Bytes()) {
			t.Fatal("generation observation bytes changed")
		}
	}
	stdout.Reset()
	if code := run([]string{"body-codegen", "--json", "--activity", "Qualified", filepath.Join(output, "realized.gooo")},
		&stdout, &diagnostics); code != exitOK {
		t.Fatal("realized source cannot be used for the next generation", code, diagnostics.String())
	}
	var next bodycodegen.Result
	if err := json.Unmarshal(stdout.Bytes(), &next); err != nil || next.GoooSource != realization.Source {
		t.Fatal("checkpoint failed fixed-point generation", err)
	}
	stdout.Reset()
	diagnostics.Reset()
	if code := run(args, &stdout, &diagnostics); code != exitFailure || !strings.Contains(diagnostics.String(), "new directory") {
		t.Fatal("realization overwrote an existing directory", code)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	args[len(args)-1] = filepath.Join(root, "cancelled")
	if code := runBodyRealizeContext(ctx, args[1:], &stdout, &diagnostics); code != exitFailure {
		t.Fatal("cancelled realization succeeded")
	}
	if _, err := os.Stat(args[len(args)-1]); !os.IsNotExist(err) {
		t.Fatal("cancelled realization created output")
	}
}

func TestBodyRealizeCLIRequiresCompleteArguments(t *testing.T) {
	for _, args := range [][]string{{}, {"--source"}, {"--source", "x"}, {"--bad", "x"},
		{"--source", "x", "--source", "y"}, {"--out", "--source"}} {
		var stdout, stderr bytes.Buffer
		if runBodyRealize(args, &stdout, &stderr) != exitUsage || stdout.Len() != 0 {
			t.Fatal("incomplete realization arguments accepted", args)
		}
	}
}
