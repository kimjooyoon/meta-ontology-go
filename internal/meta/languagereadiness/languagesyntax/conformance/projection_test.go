package languagesyntax_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

type projectionOverlay struct {
	fs.FS
	path string
	data []byte
}

func (o projectionOverlay) Open(name string) (fs.File, error) {
	if name != o.path {
		return o.FS.Open(name)
	}
	if o.data == nil {
		return nil, fs.ErrNotExist
	}
	return fstest.MapFS{name: &fstest.MapFile{Data: o.data}}.Open(name)
}

func TestProjectionCorpusBindsActualSourceAndBothArtifacts(t *testing.T) {
	repository, raw := fixture(t)
	baseline := languagesyntax.Evaluate(repository, testHead, raw, languageconcept.BuildArtifact(repository))
	if baseline.Decision != languagesyntax.DecisionPass || len(baseline.ProjectionUnits) != 1 {
		t.Fatalf("projection corpus not proven: %s %v", baseline.Decision, baseline.ProjectionUnits)
	}
	if baseline.Summary.GetPutLaws != 78 || baseline.Summary.PutGetLaws != 78 {
		t.Fatal("structure profile changed the existing general BX law denominator")
	}
	for _, sample := range []struct {
		name, path, old, replacement string
		missing                      bool
	}{
		{"changed-source-type", "internal/completeness/receipt.gooo", "type Count required one", "type Text required one", false},
		{"changed-generated-go", "internal/completeness/receipt.generated.go", "Numerator   int", "Numerator   string", false},
		{"changed-json-schema", "internal/completeness/receipt.schema.json", "\"minimum\": 0", "\"minimum\": -1", false},
		{"missing-generated-go", "internal/completeness/receipt.generated.go", "", "", true},
		{"missing-json-schema", "internal/completeness/receipt.schema.json", "", "", true},
	} {
		t.Run(sample.name, func(t *testing.T) {
			var altered []byte
			if !sample.missing {
				original, err := fs.ReadFile(repository, sample.path)
				if err != nil {
					t.Fatal(err)
				}
				altered = bytes.Replace(original, []byte(sample.old), []byte(sample.replacement), 1)
				if bytes.Equal(altered, original) {
					t.Fatal("counterexample did not alter the actual input")
				}
			}
			overlay := projectionOverlay{repository, sample.path, altered}
			report := languagesyntax.Evaluate(overlay, testHead, raw, languageconcept.BuildArtifact(overlay))
			if err := languagesyntax.Validate(report, testHead); err != nil {
				t.Fatal(err)
			}
			if report.Decision != languagesyntax.DecisionClosed || report.Summary.ProjectionSatisfied != 0 ||
				report.Summary.ProjectionTotal != 1 || report.Summary.Satisfied != 81 {
				t.Fatalf("projection failure was hidden or changed ordinary case coverage: %#v", report.Summary)
			}
			if sample.missing && (report.Resolution != languagesyntax.ResolutionLower || report.Summary.ProjectionUnresolved != 1) {
				t.Fatal("missing artifact was not preserved as unresolved")
			}
		})
	}
	baseline.ProjectionUnits = nil
	if languagesyntax.Validate(baseline, testHead) == nil {
		t.Fatal("consumer accepted an omitted projection obligation")
	}
}

func TestProjectionRegistryCannotDropOrRedirectObligations(t *testing.T) {
	repository, raw := fixture(t)
	for _, mutate := range []func(*languagesyntax.Registry){
		func(r *languagesyntax.Registry) { r.ProjectionUnits = nil },
		func(r *languagesyntax.Registry) { r.ProjectionUnits[0].Profile = "invented/profile" },
		func(r *languagesyntax.Registry) { r.ProjectionUnits[0].GoPath = "unrelated.go" },
	} {
		var registry languagesyntax.Registry
		if err := json.Unmarshal(raw, &registry); err != nil {
			t.Fatal(err)
		}
		mutate(&registry)
		changed, err := json.Marshal(registry)
		if err != nil {
			t.Fatal(err)
		}
		report := languagesyntax.Evaluate(repository, testHead, changed, languageconcept.BuildArtifact(repository))
		if err := languagesyntax.Validate(report, testHead); err != nil {
			t.Fatal(err)
		}
		if report.Decision != languagesyntax.DecisionClosed || report.Summary.Unresolved != 81 ||
			report.Summary.ProjectionUnresolved != 1 || report.Summary.ProjectionTotal != 1 {
			t.Fatal("registry drift reduced the fixed proof obligations")
		}
	}
}
