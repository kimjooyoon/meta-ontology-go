package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestPackageMapSavedOriginReader(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	var output, diagnostics bytes.Buffer
	args := []string{"package", "execute", "--json", "--cases", filepath.Join(root, "examples/collection-map/cases.json"), filepath.Join(root, "examples/collection-map/gooo.workspace.json")}
	if code := run(args, &output, &diagnostics); code != exitOK {
		t.Fatal(code, output.String(), diagnostics.String())
	}
	var producer packageExecutionReceipt
	if err := bodyexecution.DecodeExecutionReceipt(output.Bytes(), &producer); err != nil {
		t.Fatal("full saved receipt", err)
	}
	if producer.Result == nil || producer.Decision != "PASS" || producer.Result.Runtime.FinitePassed != 2 {
		t.Fatal("authored producer cases", producer.Error)
	}
	for _, site := range producer.Result.Runtime.ReadSites {
		if site.Source == nil || site.Source.Status != "MAPPED" {
			t.Fatal("map read source", site)
		}
	}
	path := filepath.Join(t.TempDir(), "producer.json")
	if err := os.WriteFile(path, output.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	diagnostics.Reset()
	args = []string{"package", "execute", "--json", "--origin-receipt", path, "--origin-kind", "elements", filepath.Join(root, "examples/native-origin-reader/elements/elements.workspace.json")}
	if code := run(args, &output, &diagnostics); code != exitOK {
		t.Fatal(code, output.String(), diagnostics.String())
	}
	var reader packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &reader); err != nil {
		t.Fatal(err)
	}
	if reader.Decision != "OBSERVED" || reader.Result == nil || reader.OriginInput == nil || reader.OriginInput.ProducerExecutions != 0 || reader.OriginInput.ProducerModelCalls != 0 || len(reader.Result.Runtime.Traces) != 5 || !reader.Result.Runtime.RuntimeReplayed {
		t.Fatal("saved origin boundary", reader.Error, reader.OriginInput)
	}
	for _, trace := range reader.Result.Runtime.Traces {
		var value struct {
			Category string `json:"category"`
			Count    int    `json:"count"`
		}
		if len(trace.Deliveries) != 1 {
			t.Fatal("reader deliveries", trace)
		}
		if err := json.Unmarshal(trace.Deliveries[0].Actual, &value); err != nil || value.Category != "ONE_READ" || value.Count != 1 {
			t.Fatal("Gooo source explanation", value, err)
		}
	}
}
