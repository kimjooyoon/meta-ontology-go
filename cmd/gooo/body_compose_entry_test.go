package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const savedEntrySource = "../../examples/body-codegen/native-single-activity.gooo.fixture"
const savedEntryCases = "../../examples/body-codegen/native-single-activity-cases.json"

func savedEntryFixture(t *testing.T) (string, bodyexecution.Composition) {
	t.Helper()
	source, err := os.ReadFile(savedEntrySource)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(savedEntryCases)
	if err != nil {
		t.Fatal(err)
	}
	suite, err := bodyexecution.DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := bodyexecution.GenerateCompositionWithOptions(context.Background(), savedEntrySource,
		source, suite, bodyexecution.CompositionOptions{EntryActivity: "Clamp"})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "composition.json")
	writeSavedEntryFixture(t, path, prior)
	return path, prior
}

func writeSavedEntryFixture(t *testing.T, path string, prior bodyexecution.Composition) {
	t.Helper()
	raw, err := json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestBodyComposeSavedEntryAssertionKeepsNativeReplay(t *testing.T) {
	path, prior := savedEntryFixture(t)
	var out, diagnostics bytes.Buffer
	args := []string{"body-compose", "--source", savedEntrySource, "--cases", savedEntryCases,
		"--composition", path, "--entry", "Clamp"}
	if code := run(args, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var result bodyCompositionOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.GeneratedNow || result.Composition.Plan.EntryActivity != "Clamp" ||
		result.Composition.GeneratedSHA256 != prior.GeneratedSHA256 || result.Runtime.ModelCalls != 0 ||
		result.Runtime.FinitePassed != 3 || result.Runtime.FiniteTotal != 3 || len(result.Runtime.Runs) != 2 {
		t.Fatal("repeated entry changed the saved plan or execution", result.Runtime.Stage)
	}
}

func TestBodyComposeSavedEntryMismatchHasNoExecutionOrOutput(t *testing.T) {
	path, _ := savedEntryFixture(t)
	for _, mode := range []string{"--composition", "--resume-composition"} {
		t.Run(mode, func(t *testing.T) { assertSavedEntryMismatch(t, path, mode) })
	}
}

func assertSavedEntryMismatch(t *testing.T, path, mode string) {
	t.Helper()
	directory := filepath.Join(t.TempDir(), "must-not-exist")
	args := []string{"body-compose", "--source", savedEntrySource, "--cases", savedEntryCases,
		mode, path, "--entry", "Different", "--go-bin", "go-must-not-run", "--out", directory}
	if mode == "--resume-composition" {
		args = append(args, "--assembly-policy", "../../examples/assembly-explainer/main.gooo.fixture",
			"--policy-activity", "Explain")
	}
	var out, diagnostics bytes.Buffer
	if code := run(args, &out, &diagnostics); code != exitFailure || out.Len() != 0 ||
		!strings.Contains(diagnostics.String(), `requested "Different", saved "Clamp"`) {
		t.Fatal("entry conflict was not reported before execution", code, diagnostics.String())
	}
	if _, err := os.Stat(directory); !os.IsNotExist(err) {
		t.Fatal("entry conflict wrote an output directory", err)
	}
}

func TestBodyComposeSavedEntryDoesNotAdmitTamperedProjection(t *testing.T) {
	path, prior := savedEntryFixture(t)
	prior.GeneratedSHA256 = "sha256:" + strings.Repeat("0", 64)
	writeSavedEntryFixture(t, path, prior)
	var out, diagnostics bytes.Buffer
	args := []string{"body-compose", "--source", savedEntrySource, "--cases", savedEntryCases,
		"--composition", path, "--entry", "Clamp", "--go-bin", "go-must-not-run"}
	if code := run(args, &out, &diagnostics); code != exitFailure {
		t.Fatal("matching entry admitted a changed saved projection", code)
	}
	var result bodyCompositionOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.GeneratedNow || result.Runtime.ProjectionReplayed || result.Runtime.Toolchain.Started ||
		result.Runtime.Build.Started || len(result.Runtime.Runs) != 0 || diagnostics.Len() == 0 {
		t.Fatal("saved integrity failure reached compilation or execution", diagnostics.String())
	}
}
