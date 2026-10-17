package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
)

func TestComparedReportReadBudget(t *testing.T) {
	const head = "0000000000000000000000000000000000000000"
	report := metaff.Evaluate(metaff.Input{ExpectedHeadSHA: head})
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, extra := range []bool{false, true} {
		name, data := "exact", raw
		if extra {
			name, data = "excess", append(append([]byte(nil), raw...), ' ')
		}
		t.Run(name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "report.json")
			if err := os.WriteFile(filename, data, 0600); err != nil {
				t.Fatal(err)
			}
			decoded, err := readComparedReportLimit(filename, len(raw))
			if (err != nil) != extra {
				t.Fatal("read budget did not use exact input byte length", extra, err)
			}
			if !extra && decoded.ReportDigest != report.ReportDigest {
				t.Fatal("boundary-sized report changed while reading")
			}
		})
	}
}
