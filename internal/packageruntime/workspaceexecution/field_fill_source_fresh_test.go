package workspaceexecution

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestFieldCandidateLocations16541(t *testing.T) {
	const source = `package fieldfill
namespace fieldfill16541
entity Integer id "urn:gooo:type:integer"
entity Box id "urn:fieldfill16541:box" fields {
 field value id "urn:fieldfill16541:value" type integer required one
}
activity Change(Integer)->Integer id "urn:fieldfill16541:change" computes "let box=Box{value:input+__GOOO_BODY_HOLE_delta__};box.value=box.value+3;return box.value" assembling {
 source_fill intent "필드에 선택한 증가량을 반영한 뒤 3을 더한다" {
  hole "delta"
  candidate "delta17" {fill "delta" "17"}
  candidate "delta41" {fill "delta" "41"}
 }
 case "29" -> "49"
}
`
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "origin/fieldfill16541", Activity: "Change"}, Packages: []packageruntime.PackageSpec{{Path: "origin/fieldfill16541", Name: "fieldfill", Sources: []packageruntime.Source{{Filename: "필드후보16541.gooo", Content: source}}}}}
	cases, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/fieldfill16541:Change":107},"expected":{"origin/fieldfill16541:Change":127}},{"inputs":{"origin/fieldfill16541:Change":-113},"expected":{"origin/fieldfill16541:Change":-93}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, cases, "", "")
	if err != nil || result.Runtime.FinitePassed != 2 {
		t.Fatal("new field fill", err, result.Runtime.FinitePassed)
	}
	candidates, unmappedValues, fieldRelations := 0, 0, 0
	for _, b := range result.Runtime.FieldFileSources.Bindings {
		site := result.Runtime.FieldSites[0]
		if b.Kind == "write" {
			site = result.Runtime.FieldSites[b.SiteIndex]
		} else {
			site = result.Runtime.FieldUseSites[b.SiteIndex]
		}
		if site.Kind == "record_constructor" && b.Value.Status == "UNMAPPED" && b.Value.Reason == "ASSEMBLED_RANGE_HAS_NO_TEMPLATE_LOCATION" {
			unmappedValues++
		}
		if b.Construction == nil {
			continue
		}
		for _, fragment := range b.Construction.Fragments {
			if fragment.Relation == "field_expression" {
				fieldRelations++
			}
		}
	}
	rows, report, err := bodyexecution.SelectRecordedNativeOrigins(result.Runtime, bodyexecution.NativeOriginQuery{Kind: "field-input-locations", Limit: 128})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		var fact bodyexecution.CompositionFieldFileFact
		if err := json.Unmarshal(row, &fact); err != nil {
			t.Fatal(err)
		}
		if fact.Category == "declared_candidate" {
			if fact.CandidateID != "delta17" || fact.HoleID != "delta" || fact.Mapping != "MAPPED" || fact.Filename != "필드후보16541.gooo" || source[fact.SourceStart:fact.SourceEnd] != "17" {
				t.Fatal("original candidate anchor", fact)
			}
			candidates++
		}
	}
	if candidates < 2 || unmappedValues != 1 || fieldRelations == 0 || report.ProducerExecutions != 0 || result.Runtime.ModelCalls != 0 {
		t.Fatal("field fill populations", candidates, unmappedValues, fieldRelations)
	}
	if root := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); root != "" {
		f, e := os.Create(filepath.Join(root, "field-fill16541-runtime.json.gz"))
		if e != nil {
			t.Fatal(e)
		}
		z := gzip.NewWriter(f)
		if e = json.NewEncoder(z).Encode(result.Runtime); e != nil {
			t.Fatal(e)
		}
		if e = z.Close(); e != nil {
			t.Fatal(e)
		}
		if e = f.Close(); e != nil {
			t.Fatal(e)
		}
	}
	t.Logf("fresh native2/2;candidate anchors%d;cross-fragment RHS explicitly unmapped%d;field-expression relations%d;selection and native denominators separate;model0;saved-read producer0", candidates, unmappedValues, fieldRelations)
}
