package bodypathstream

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestSourceAssemblyStreamRunsParallelWithoutDocument(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	var input bytes.Buffer
	for _, activity := range []string{"Clamp", "Qualified", "Clamp", "Qualified"} {
		raw, err := json.Marshal(Request{Schema: RequestSchema, CorrelationID: activity, Activity: activity, Source: string(source)})
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(raw, []byte(`"document"`)) {
			t.Fatal("source assembly still needs a JSON plan")
		}
		input.Write(raw)
		input.WriteByte('\n')
	}
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := Run(context.Background(), g, io.NopCloser(&input), asWriteCloser(&output), 2); err != nil {
		t.Fatal(err)
	}
	results := decodeResults(t, output.Bytes())
	if len(results) != 4 {
		t.Fatal("source-only parallel requests lost", len(results))
	}
	for _, r := range results {
		if r.Status != "completed" || r.Response == nil || r.Response.Report.BodyPaths.FunctionalCompleteness != 100 ||
			r.Response.Report.BodyPaths.Search.Selection.ModelCalls != 0 {
			t.Fatal("source-only worker generation failed", r)
		}
	}
}

func TestSourceAssemblyFileRunSavesAndReplaysWithoutRecipe(t *testing.T) {
	out := filepath.Join(t.TempDir(), "observations")
	args := []string{"--source", "../../examples/body-codegen/source-assembly.gooo.fixture", "--activity", "Clamp",
		"--cases", "../../examples/body-codegen/source-assembly-clamp-cases.json", "--out", out, "--repeat", "2", "--timing"}
	var stdout, stderr bytes.Buffer
	if code := RunFilesCommand(context.Background(), "files", args, &stdout, &stderr); code != 0 {
		t.Fatal(code, stderr.String())
	}
	rows := fileSummary(t, out)
	if len(rows) != 2 || rows[0].Passed != 7 || rows[1].Passed != 7 || !rows[1].ArtifactReused {
		t.Fatal("source-only native observations or reuse failed", rows)
	}
	if _, err := os.Stat(filepath.Join(out, "recipe.json")); !os.IsNotExist(err) {
		t.Fatal("source-only run saved an invented recipe", err)
	}
	stdout.Reset()
	if code := RunFilesCommand(context.Background(), "files", []string{"--out", out, "--verify-timing"}, &stdout, &stderr); code != 0 {
		t.Fatal("source-only saved timing does not verify", code, stderr.String())
	}
}
