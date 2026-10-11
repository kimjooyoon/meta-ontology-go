package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageComputesOuterWhitespaceExecutesPreservedText(t *testing.T) {
	root := t.TempDir()
	body := "\n\treturn \" left \" + input + \" right \"\n  "
	quoted, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	writeWorkspaceSource(t, root, "main.gooo.fixture", "package whitespace\nnamespace whitespace\n"+
		"entity Text id \"gooo://std/text\"\nactivity Main(Text) -> Text computes "+string(quoted)+"\n")
	manifest := writeWorkspaceManifest(t, root, `{"schema":"gooo/package-workspace-manifest/v1",
"entry":{"package_path":"example/whitespace","activity":"Main"},
"packages":[{"path":"example/whitespace","name":"whitespace","imports":[],"sources":["main.gooo.fixture"]}]}`)
	cases := filepath.Join(root, "cases.json")
	raw := `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"example/whitespace:Main":"한글"},"expected":{"example/whitespace:Main":" left 한글 right "}},
{"inputs":{"example/whitespace:Main":""},"expected":{"example/whitespace:Main":" left  right "}}]}`
	if err := os.WriteFile(cases, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	var output, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", cases, manifest}, &output, &diagnostics); code != exitOK {
		t.Fatal("multiline computes body rejected", code, diagnostics.String(), output.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil || receipt.Result == nil ||
		receipt.Result.Runtime.FinitePassed != 2 || !receipt.Result.Runtime.RuntimeReplayed {
		t.Fatal("literal spaces changed during execution", err)
	}
}
