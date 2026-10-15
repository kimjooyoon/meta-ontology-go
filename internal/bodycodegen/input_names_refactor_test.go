package bodycodegen

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestNamedExampleRenameKeepsSignatureLocationsFirst(t *testing.T) {
	source, err := os.ReadFile("../../examples/named-inputs/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	r, err := RenameActivitySource(context.Background(), "main.gooo", source, "Difference", "차액")
	if err != nil || r.CallSites != 1 || !strings.Contains(r.Source, "차액(원금, 공제)") {
		t.Fatal(r, err)
	}
	view, err := InspectTypedValueOrigins("renamed.gooo", []byte(r.Source), "차액")
	if err != nil || view.Status != "RESOLVED" || len(view.Inputs) != 2 {
		t.Fatal(view, err)
	}
	for i, name := range []string{"minuend", "subtrahend"} {
		d := view.Inputs[i].Declaration
		if d == nil || d.Name != name || d.SourceSHA256 != r.OutputSHA256 ||
			d.ActivityID != "urn:gooo:example:named-inputs:difference" || r.Source[d.NameStart:d.NameEnd] != name {
			t.Fatal("current source binding lost after Unicode rename", d)
		}
	}
}
