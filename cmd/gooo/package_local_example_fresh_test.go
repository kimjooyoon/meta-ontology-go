package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestWorkspaceLocalExample12753(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	sourceRoot := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if sourceRoot == "" {
		sourceRoot = "../.."
	}
	root := filepath.Join(sourceRoot, "examples/workspace-local-origins")
	var output, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "gooo.workspace.json")}, &output, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), output.String())
	}
	saveLocalFiles12421(t, "producer12753.json", output.Bytes())
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.Result.Runtime.FinitePassed != 2 {
		t.Fatal("published example", err)
	}
	runtime := receipt.Result.Runtime
	rows, _, err := bodyexecution.SelectRecordedNativeOrigins(runtime, bodyexecution.NativeOriginQuery{Kind: "local-input-locations", Limit: 32})
	if err != nil || len(rows) != 28 {
		t.Fatal("example origin rows", err, len(rows))
	}
	for _, raw := range rows {
		var fact bodyexecution.CompositionLocalFileFact
		if err := json.Unmarshal(raw, &fact); err != nil || fact.Mapping != "MAPPED" || fact.Filename != "main.gooo" {
			t.Fatal("example source anchor", err, fact)
		}
	}
	checkAssemblyFileReader12503(t, t.TempDir(), "example12753", output.Bytes(), rows)
	t.Log("actual new Gooo example2/2;28 file origins and28Gooo reader outputs matched;branch-local names and nested record/helper source retained;original producer/model calls during reads0")
}
