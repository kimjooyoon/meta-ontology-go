package fuzz

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const contractManifestPath = "testdata/contract.json"

type contractManifest struct {
	SchemaVersion      int               `json:"schema_version"`
	MaxSourceBytes     int               `json:"max_source_bytes"`
	OperationTimeoutMS int               `json:"operation_timeout_ms"`
	Fixtures           []contractFixture `json:"fixtures"`
}

type contractFixture struct {
	Name            string                  `json:"name"`
	File            string                  `json:"file"`
	Kind            string                  `json:"kind"`
	WantDiagnostics int                     `json:"want_diagnostics"`
	MinDiagnostics  int                     `json:"min_diagnostics"`
	Declarations    int                     `json:"declarations"`
	RequiredCodes   []syntax.DiagnosticCode `json:"required_codes"`
}

func TestContractFixtures(t *testing.T) {
	manifest := loadContractManifest(t)
	if manifest.SchemaVersion != 1 || manifest.MaxSourceBytes != maxSourceBytes || manifest.OperationTimeoutMS != int(operationLimit.Milliseconds()) {
		t.Fatalf("manifest limits do not match harness: %#v", manifest)
	}
	for _, fixture := range manifest.Fixtures {
		fixture := fixture
		t.Run(fixture.Name, func(t *testing.T) {
			assertContractFixture(t, fixture, manifest.MaxSourceBytes)
		})
	}
}

func loadContractManifest(t testing.TB) contractManifest {
	t.Helper()
	source, err := os.ReadFile(contractManifestPath)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := decodeContractManifest(source)
	if err != nil {
		t.Fatalf("decode contract manifest: %v", err)
	}
	if len(manifest.Fixtures) == 0 {
		t.Fatal("contract manifest has no fixtures")
	}
	return manifest
}

func decodeContractManifest(source []byte) (contractManifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.DisallowUnknownFields()
	var manifest contractManifest
	if err := decoder.Decode(&manifest); err != nil {
		return contractManifest{}, err
	}
	return manifest, nil
}

func assertContractFixture(t *testing.T, fixture contractFixture, maxBytes int) {
	t.Helper()
	if fixture.Kind != "valid" && fixture.Kind != "negative" {
		t.Fatalf("unknown fixture kind %q", fixture.Kind)
	}
	if filepath.Base(fixture.File) != fixture.File || filepath.Ext(fixture.File) != ".gooo" {
		t.Fatalf("fixture path escapes testdata or is not .gooo: %q", fixture.File)
	}
	source, err := os.ReadFile(filepath.Join("testdata", fixture.File))
	if err != nil {
		t.Fatal(err)
	}
	if len(source) > maxBytes {
		t.Fatalf("fixture is %d bytes, limit is %d", len(source), maxBytes)
	}
	first := parseContractSource(t, string(source))
	second := parseContractSource(t, string(source))
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("fixture parse was not deterministic: first=%#v second=%#v", first, second)
	}
	assertDiagnostics(t, string(source), first.Diagnostics)
	assertFile(t, string(source), first.File)
	assertFixtureExpectations(t, fixture, first)
}

func parseContractSource(t *testing.T, source string) parseResult {
	t.Helper()
	return runWithLimit(t, "fixture parse", func() parseResult {
		file, diagnostics := syntax.ParseFile(fuzzFilename, source)
		return parseResult{File: file, Diagnostics: diagnostics}
	})
}

func assertFixtureExpectations(t *testing.T, fixture contractFixture, result parseResult) {
	t.Helper()
	if fixture.WantDiagnostics >= 0 && len(result.Diagnostics) != fixture.WantDiagnostics {
		t.Fatalf("diagnostics=%d, want %d", len(result.Diagnostics), fixture.WantDiagnostics)
	}
	if len(result.Diagnostics) < fixture.MinDiagnostics {
		t.Fatalf("diagnostics=%d, want at least %d", len(result.Diagnostics), fixture.MinDiagnostics)
	}
	if fixture.Declarations >= 0 && len(result.File.Declarations) != fixture.Declarations {
		t.Fatalf("declarations=%d, want %d", len(result.File.Declarations), fixture.Declarations)
	}
	for _, required := range fixture.RequiredCodes {
		if !hasDiagnosticCode(result.Diagnostics, required) {
			t.Fatalf("missing required diagnostic %q: %v", required, result.Diagnostics)
		}
	}
}

func hasDiagnosticCode(diagnostics syntax.Diagnostics, want syntax.DiagnosticCode) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == want {
			return true
		}
	}
	return false
}
