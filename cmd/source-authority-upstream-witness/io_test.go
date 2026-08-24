package main

import (
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthorityupstream"
)

func TestWriteArtifactsRefusesExistingDirectory(t *testing.T) {
	output := filepath.Join(t.TempDir(), "evidence")
	suite := sourceauthorityupstream.Suite{
		Schema: sourceauthorityupstream.SuiteSchema,
		Cases: []sourceauthorityupstream.CaseResult{{
			ID: sourceauthorityupstream.CaseExact,
			Receipt: sourceauthorityupstream.Receipt{Schema: sourceauthorityupstream.ReceiptSchema},
		}},
	}
	if err := writeArtifacts(output, suite); err != nil {
		t.Fatal(err)
	}
	if err := writeArtifacts(output, suite); err == nil {
		t.Fatal("existing evidence directory was overwritten")
	}
}
