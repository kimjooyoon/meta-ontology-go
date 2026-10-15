package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestReadOnlyRegistrationInventoryCLI143901(t *testing.T) {
	var output bytes.Buffer
	if err := writeRegistrationInventory("../..", &output); err != nil {
		t.Fatal(err)
	}
	var report struct {
		Schema, SemanticAdmission string
		Unregistered, Missing     []string
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(output.Bytes(), &fields); err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal(fields["schema"], &report.Schema)
	_ = json.Unmarshal(fields["semantic_admission"], &report.SemanticAdmission)
	_ = json.Unmarshal(fields["unregistered"], &report.Unregistered)
	_ = json.Unmarshal(fields["missing"], &report.Missing)
	if report.Schema != "gooo/syntax-source-inventory/v1" || report.SemanticAdmission != "UNASSESSED" || len(report.Unregistered) != 0 || len(report.Missing) != 0 || !strings.Contains(output.String(), "internal/meta/originview/roles.gooo") {
		t.Fatal(output.String())
	}
}
