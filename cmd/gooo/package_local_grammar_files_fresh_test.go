package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestOriginalGrammarFiles12547(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := t.TempDir()
	files := assemblyFiles12503("grammar")
	replace := strings.NewReplacer("grammar12503", "grammar12547", "797", "853", "1584", "1692", "-809", "-857", "-22", "-18", "811", "863", "1598", "1702", "-821", "-877", "-34", "-38")
	for name, content := range files {
		files[name] = replace.Replace(content)
		if err := os.WriteFile(filepath.Join(root, name), []byte(files[name]), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var output, diagnostics bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "w.json")}, &output, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), output.String())
	}
	saveLocalFiles12421(t, "producer12547.json", output.Bytes())
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(output.Bytes(), &receipt); err != nil || receipt.Result == nil || receipt.Result.Runtime.FinitePassed != 2 {
		t.Fatal(err)
	}
	rows, _, err := bodyexecution.SelectRecordedNativeOrigins(receipt.Result.Runtime, bodyexecution.NativeOriginQuery{Kind: "local-input-locations", Limit: 128})
	if err != nil {
		t.Fatal(err)
	}
	checkAssemblyFileRows12503(t, "grammar", files, rows, receipt.Result.Runtime)
	checkAssemblyFileReader12503(t, root, "grammar12547", output.Bytes(), rows)
	t.Logf("different grammar source/inputs producer2/2;original grammar/file and retained declaration/use facts%d;no invented authored expression/candidate range;Gooo reader%doutputs matched;original producer/model calls during reads0", len(rows), len(rows))
}
