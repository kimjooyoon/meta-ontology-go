package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
)

func TestReportComparisonRejectsUnsupportedJSON(t *testing.T) {
	const head = "0000000000000000000000000000000000000000"
	report := metaff.Evaluate(metaff.Input{ExpectedHeadSHA: head})
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	variants := []struct {
		name string
		raw  []byte
	}{
		{"unknown", append([]byte(`{"unknown":true,`), raw[1:]...)},
		{"duplicate", append([]byte(`{"schema":"shadow",`), raw[1:]...)},
		{"trailing", append(append([]byte(nil), raw...), []byte(`{}`)...)},
		{"utf8", append([]byte{255}, raw...)},
		{"empty", nil},
	}
	root := t.TempDir()
	good := filepath.Join(root, "good.json")
	if err := os.WriteFile(good, raw, 0600); err != nil {
		t.Fatal(err)
	}
	for _, variant := range variants {
		t.Run(variant.name, func(t *testing.T) {
			bad := filepath.Join(root, variant.name+".json")
			if err := os.WriteFile(bad, variant.raw, 0600); err != nil {
				t.Fatal(err)
			}
			if err := run(config{compare: bad, check: good, head: head}); err == nil {
				t.Fatal("saved comparison discarded unsupported input", variant.name)
			}
			if err := writeOrCheck("", bad, report); err == nil {
				t.Fatal("fresh comparison discarded unsupported input", variant.name)
			}
		})
	}
}
