package workspacesplit_test

import (
	_ "embed"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

//go:embed testdata/shared.gooo
var sharedSource string

func sharedFixture() (fstest.MapFS, workspaceinput.Definition, workspaceinput.SplitRequest) {

	definition := workspaceinput.Definition{Schema: workspaceinput.Schema,
		Entry:    packageruntime.EntrySpec{PackagePath: "shared/app", Activity: "Run"},
		Packages: []workspaceinput.Package{{Path: "shared/app", Sources: []string{"shared.gooo"}}}}
	request := workspaceinput.SplitRequest{PackagePath: "shared/app", Source: "shared.gooo",
		BeforeID: "urn:shared:run", Into: "run.gooo"}
	return fstest.MapFS{"shared.gooo": {Data: []byte(sharedSource)}}, definition, request
}

func jointCases(t *testing.T, entry string) (bodyexecution.CompositionCases, int) {
	t.Helper()
	raw, total := `{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"shared/app:Credit":2},"expected":{"shared/app:Credit":9,"shared/app:Run":6}},
 {"inputs":{"shared/app:Credit":-10},"expected":{"shared/app:Credit":-3,"shared/app:Run":-6}}
]}`, 4
	if entry == "Credit" {
		raw, total = `{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"shared/app:Credit":2},"expected":{"shared/app:Credit":9}},
 {"inputs":{"shared/app:Credit":-10},"expected":{"shared/app:Credit":-3}}
]}`, 2
	}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	return suite, total
}

func sharedOwners() (fstest.MapFS, []workspacesplit.Owner) {
	files, definition, request := sharedFixture()
	other := definition
	other.Entry.Activity = "Credit"
	return files, []workspacesplit.Owner{
		{ManifestPath: "first.workspace.json", Definition: definition, Request: request},
		{ManifestPath: "second.workspace.json", Definition: other, Request: request},
	}
}
