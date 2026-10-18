package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestPackageFilterExampleAndSavedReplay(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	manifest := filepath.Join(root, "examples/collection-filter/gooo.workspace.json")
	cases := filepath.Join(root, "examples/collection-filter/cases.json")
	var output, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", cases, manifest}, &output, &diagnostics); code != exitOK {
		t.Fatal("example execution", code, diagnostics.String())
	}
	var result packageExecutionReceipt
	if err := bodyexecution.DecodeExecutionReceipt(output.Bytes(), &result); err != nil || result.Result == nil || result.Decision != "PASS" || result.Result.Runtime.FinitePassed != 3 {
		t.Fatal("source cases", err, result.Error)
	}
	if len(result.Result.Runtime.ReadSites) != 1 || result.Result.Runtime.ReadSites[0].Source.Status != "MAPPED" ||
		len(result.Result.Runtime.ConstructionSites) != 1 || result.Result.Runtime.ConstructionSites[0].PredicateActivityID != "urn:collection-filter:keep-nonempty" || result.Result.Runtime.ModelCalls != 0 {
		t.Fatal("source-owned filter boundary", result.Result.Runtime.ReadSites, result.Result.Runtime.ConstructionSites)
	}
	path := filepath.Join(t.TempDir(), "filter-receipt.json")
	if err := os.WriteFile(path, output.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	output.Reset()
	diagnostics.Reset()
	if code := run([]string{"package", "replay", "--json", "--receipt", path, "--cases", cases, manifest}, &output, &diagnostics); code != exitOK {
		t.Fatal("saved filter replay", code, diagnostics.String())
	}
	var replay packageExecutionReceipt
	if err := bodyexecution.DecodeExecutionReceipt(output.Bytes(), &replay); err != nil || replay.Result == nil || replay.Decision != "PASS" || replay.Result.Runtime.FinitePassed != 3 || replay.Result.Replay.ModelCalls != 0 {
		t.Fatal("deterministic replay", err, replay.Error)
	}
}
