package workspaceexecution

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestLengthIntegerNative20809(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_FIELD_ROOT")
	if root == "" {
		root = filepath.Join("..", "..", "..")
	}
	path := filepath.Join(root, "examples", "length-aware-summary")
	source, err := os.ReadFile(filepath.Join(path, "summary.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(path, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	cases, err := bodyexecution.DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry:    packageruntime.EntrySpec{PackagePath: "origin/lengthsummary20809", Activity: "Measure"},
		Packages: []packageruntime.PackageSpec{{Path: "origin/lengthsummary20809", Name: "lengthsummary", Sources: []packageruntime.Source{{Filename: "길이집계20809.gooo", Content: string(source)}}}}}
	result, err := ExecuteWorkspace(context.Background(), manifest, cases, "", "")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Runtime
	if r.FinitePassed != 3 || r.FiniteTotal != 3 || !r.RuntimeReplayed || r.ModelCalls != 0 {
		t.Fatal("new length-input execution scope", r.FinitePassed, r.FiniteTotal, r.RuntimeReplayed, r.ModelCalls)
	}
	if output := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); output != "" {
		file, err := os.Create(filepath.Join(output, "length-native20809.json.gz"))
		if err != nil {
			t.Fatal(err)
		}
		zip := gzip.NewWriter(file)
		if err := json.NewEncoder(zip).Encode(r); err != nil {
			t.Fatal(err)
		}
		if err := zip.Close(); err != nil {
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if r.FieldFileSources == nil {
		t.Fatal("field file locations absent")
	}
	mapped, authoredLengths := 0, 0
	for _, binding := range r.FieldFileSources.Bindings {
		value := binding.Value
		if value.Status != "MAPPED" {
			continue
		}
		mapped++
		if value.Filename != "길이집계20809.gooo" || string(source[value.Start:value.End]) != value.Expression || strings.Contains(value.Expression, "int64(") {
			t.Fatal("implicit widening claimed as source text", value)
		}
	}
	for _, site := range r.LocalSites {
		if site.Source.Value != nil && strings.Contains(site.Source.Value.Expression, "len(input.labels") {
			authoredLengths++
		}
	}
	if mapped == 0 || authoredLengths < 2 {
		t.Fatal("original length and field coordinates absent", mapped, authoredLengths)
	}

	t.Log("new length-aware Gooo collection metaprogram3/3;empty/Korean+emoji/partial9-row inputs;Integer guard/helper/field assignment without authored casts;exact two-run observation;model0")
}
