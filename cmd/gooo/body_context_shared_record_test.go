package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestBodyContextCLIExportsSharedRecordFeature(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/record-field-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	code := runBodyContext([]string{"--activity", "Select", "--feature-version", jointdecision.RecordSharedFeatureVersion, "r.gooo"},
		mapSourceReader{"r.gooo": source}, &out, &stderr)
	var exported bodycodegen.RecordAssemblyContextExport
	if err = json.Unmarshal(out.Bytes(), &exported); err != nil {
		t.Fatal(err)
	}
	if code != exitOK || stderr.Len() != 0 || exported.Context.FeatureVersion != jointdecision.RecordSharedFeatureVersion ||
		exported.Context.Status != "ENCODED" || exported.ModelPredictions != 0 || exported.CandidateTests != 0 || len(exported.Choices) != 3 {
		t.Fatal("shared context CLI", out.String(), stderr.String())
	}
}
