package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/rollbackfixedpoint"
)

func TestWriteAppliedReportAndCheckPreserveOriginalBytes(t *testing.T) {
	path := "../../../internal/meta/languagereadiness/rollbackfixedpoint/testdata/applied-original-8f45c78e.json.gz"
	if root := os.Getenv("GOOO_ROLLBACK_TEST_DRAFT"); root != "" {
		path = filepath.Join(root, "internal/meta/languagereadiness/rollbackfixedpoint/testdata/applied-original-8f45c78e.json.gz")
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	var report rollbackfixedpoint.Report
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	cfg := config{output: filepath.Join(t.TempDir(), "observed.json")}
	if err := writeReport(cfg, report); err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(cfg.output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, actual) {
		t.Fatal("original rejected recovery was rewritten")
	}
	cfg.check = true
	if err := writeReport(cfg, report); err != nil {
		t.Fatal(err)
	}
	report.Source.Transformation.ReceiptCount--
	cfg.output = filepath.Join(t.TempDir(), "contradictory.json")
	cfg.check = false
	if err := writeReport(cfg, rollbackfixedpoint.Build(report.Source)); err == nil {
		t.Fatal("contradictory report accepted")
	}
}
