package semanticbinding

import (
	"go/parser"
	"go/token"
	"testing"
)

type fixturePosition struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}
type fixtureLocation struct {
	Filename string          `json:"filename"`
	Start    fixturePosition `json:"start"`
	End      fixturePosition `json:"end"`
}
type fixtureRecord struct {
	Directive string           `json:"directive"`
	Name      string           `json:"name"`
	ID        string           `json:"id"`
	Subject   string           `json:"subject,omitempty"`
	Pressure  string           `json:"pressure,omitempty"`
	Location  *fixtureLocation `json:"location,omitempty"`
}
type fixtureExpectation struct {
	Status     string          `json:"status"`
	Canonical  string          `json:"canonical"`
	Diagnostic string          `json:"diagnostic,omitempty"`
	Records    []fixtureRecord `json:"records"`
}

var fixtureNames = []string{
	"valid_bind", "valid_obligation", "rename_before", "rename_after",
	"same_name_without_directive", "detached_comment", "unknown_field",
	"duplicate_field", "duplicate_id", "invalid_uri", "multi_name_var",
	"exact_spans", "canonical_permutation_a", "canonical_permutation_b",
	"filename_identity_a", "filename_identity_b",
}

func TestFixtureCorpusIsLiteralAndParseable(t *testing.T) {
	for _, name := range fixtureNames {
		t.Run(name, func(t *testing.T) {
			source, want := loadFixture(t, name)
			if _, diagnostics := parser.ParseFile(token.NewFileSet(), name+".go", source, parser.ParseComments); diagnostics != nil {
				t.Fatalf("fixture is not valid Go: %v", diagnostics)
			}
			if want.Status != "accepted" && want.Status != "rejected" {
				t.Fatalf("status = %q, want accepted or rejected", want.Status)
			}
			if want.Records == nil {
				t.Fatal("expected records must be present as a literal array")
			}
			if want.Status == "rejected" && len(want.Records) != 0 {
				t.Fatalf("rejected fixture has records: %#v", want.Records)
			}
			for index, record := range want.Records {
				if record.Directive == "" || record.Name == "" || record.ID == "" {
					t.Fatalf("record[%d] is incomplete: %#v", index, record)
				}
			}
		})
	}
}
