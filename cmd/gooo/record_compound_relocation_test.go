package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestRegisteredRecordExamplePreservesProjection(t *testing.T) {
	archive, err := os.ReadFile("../../docs/observations/record-field-updates/first-observations.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	old := readNamedInputArchive(t, archive)["sources/examples/compound-updates/main.gooo"]
	const path = "../../examples/compound-updates/fields.gooo"
	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	before, err := bodycodegen.Generate("historical-main.gooo", old, "Summarize")
	if err != nil {
		t.Fatal(err)
	}
	after, err := bodycodegen.Generate(path, current, "Summarize")
	if err != nil || before.Source != after.Source {
		t.Fatal("moving the record example changed its selected Go projection", err)
	}
	var out, diagnostics bytes.Buffer
	if code := run([]string{"roundtrip", "--json", path}, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String(), out.String())
	}
	var report jsonReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil || report.Equivalent == nil || !*report.Equivalent || report.GetPut == nil || !*report.GetPut || report.PutGet == nil || !*report.PutGet {
		t.Fatal("registered record source coherence", err)
	}
	var runtime struct {
		SelectedSourceSHA256 string `json:"selected_source_sha256"`
	}
	if err := json.Unmarshal(readCompoundRuntimeFixture(t), &runtime); err != nil || runtime.SelectedSourceSHA256 != "sha256:"+namedInputObservationSHA(old) {
		t.Fatal("historical source remains bound to its receipt", err)
	}
	t.Log("separate registered source has identical selected Go; original receipt retains its original source; native executions 0")
}
