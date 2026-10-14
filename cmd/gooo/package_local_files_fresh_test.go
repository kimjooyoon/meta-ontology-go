package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestOriginalLocalFiles12421(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := t.TempDir()
	body := `let 기준=Extract(Result{value:ops.Adjust(input)});let copied=기준;return Result{value:copied}`
	quoted := strings.ReplaceAll(strconv.Quote(body), "기준", `\uAE30\uC900`)
	files := map[string]string{
		"types.gooo": "package localfiles\nnamespace localfiles12421\nentity Integer id \"urn:gooo:type:integer\"\nentity Result id \"urn:localfiles12421:result\" fields {field value id \"urn:localfiles12421:value\" type integer required one}\n",
		"build.gooo": "package localfiles\nnamespace localfiles12421\nimport ops \"origin/math12421\"\nactivity Extract(Result) -> Integer id \"urn:localfiles12421:extract\" computes `return input.value`\nactivity Build(Integer) -> Result id \"urn:localfiles12421:build\" computes " + quoted + "\n",
		"ops.gooo":   "package localops\nnamespace localops12421\nentity Integer id \"urn:gooo:type:integer\"\nactivity Adjust(Integer) -> Integer id \"urn:localops12421:adjust\" computes `let same=input+709;return same`\n",
		"w.json":     `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"origin/app12421","activity":"Build"},"packages":[{"path":"origin/app12421","name":"localfiles","imports":["origin/math12421"],"sources":["types.gooo","build.gooo"]},{"path":"origin/math12421","name":"localops","imports":[],"sources":["ops.gooo"]}]}`,
		"cases.json": `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/app12421:Build":719},"expected":{"origin/app12421:Build":{"value":1428}}},{"inputs":{"origin/app12421:Build":-727},"expected":{"origin/app12421:Build":{"value":-18}}}]}`,
	}
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var output, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "w.json")}, &output, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), output.String())
	}
	saveLocalFiles12421(t, "producer12421.json", output.Bytes())
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.Result.Runtime.FinitePassed != 2 {
		t.Fatal("new workspace", err)
	}
	runtime := receipt.Result.Runtime
	checkOriginalLocalFiles12421(t, runtime, files)
	rows, summary, err := bodyexecution.SelectRecordedNativeOrigins(runtime, bodyexecution.NativeOriginQuery{Kind: "local-input-locations", Limit: 32})
	if err != nil || len(rows) != 24 || summary.TotalRows != 24 {
		t.Fatal("original-file population", err, summary, len(rows))
	}
	checkLocalFileReader12421(t, root, output.Bytes(), rows)
	t.Log("new multi-package producer2/2;three static writes/three use sites;24 original-file facts;call and record renaming reversed;escaped Korean declaration/use ranges retained;Gooo reader24outputs matched;original producer/model calls during read0")
}

func checkOriginalLocalFiles12421(t *testing.T, runtime bodyexecution.CompositionRuntime, files map[string]string) {
	t.Helper()
	if runtime.LocalFileSources == nil || len(runtime.LocalFileSources.Bindings) != 6 {
		t.Fatal("static annotations", runtime.LocalFileSources)
	}
	unicode, nested := false, false
	for _, binding := range runtime.LocalFileSources.Bindings {
		for _, loc := range []bodyexecution.CompositionLocalFileLocation{binding.Declaration, binding.Value} {
			original, ok := files[loc.Filename]
			digest := sha256.Sum256([]byte(original))
			if !ok || loc.Status != "MAPPED" || loc.SourceSHA256 != "sha256:"+hex.EncodeToString(digest[:]) || loc.Start < 0 || loc.End <= loc.Start || loc.End > len(original) {
				t.Fatal("original file anchor", loc)
			}
			if loc.Expression == "기준" {
				unicode = true
				if original[loc.Start:loc.End] != `\uAE30\uC900` {
					t.Fatal("escaped token range", loc)
				}
			}
			if loc.Expression == `Extract(Result{value:ops.Adjust(input)})` {
				nested = true
			}
		}
	}
	if !unicode || !nested {
		t.Fatal("original Unicode and nested call/constructor expressions missing", unicode, nested)
	}
}

func checkLocalFileReader12421(t *testing.T, root string, producer []byte, rows []json.RawMessage) {
	t.Helper()
	path := filepath.Join(root, "saved.json")
	if err := os.WriteFile(path, producer, 0600); err != nil {
		t.Fatal(err)
	}
	sourceRoot := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if sourceRoot == "" {
		sourceRoot = "../.."
	}
	var output, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--origin-receipt", path, "--origin-kind", "local-input-locations", "--origin-limit", "32", filepath.Join(sourceRoot, "examples/native-origin-reader/local-files/reader.workspace.json")}, &output, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), output.String())
	}
	saveLocalFiles12421(t, "consumer12421.json", output.Bytes())
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.OriginInput == nil {
		t.Fatal(err)
	}
	r := receipt.Result.Runtime
	if len(r.Traces) != len(rows) || r.FiniteTotal != 0 || r.ModelCalls != 0 || receipt.OriginInput.ProducerExecutions != 0 {
		t.Fatal("reader population")
	}
	for index, row := range rows {
		var fact bodyexecution.CompositionLocalFileFact
		var got string
		if err := json.Unmarshal(row, &fact); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(r.Traces[index].Deliveries[0].Actual, &got); err != nil {
			t.Fatal(err)
		}
		if got != fact.Filename+":"+fact.EventKind+":"+fact.Category {
			t.Fatal(index, got, fact)
		}
	}
}

func saveLocalFiles12421(t *testing.T, name string, raw []byte) {
	t.Helper()
	if root := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); root != "" {
		if err := os.WriteFile(filepath.Join(root, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
