package workspacesplit_test

import (
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func sharedFixture() (fstest.MapFS, workspaceinput.Definition, workspaceinput.SplitRequest) {
	source := "package shared\nnamespace shared\nentity Integer id \"urn:gooo:type:integer\"\n" +
		"activity Credit(Integer) -> Integer id \"urn:shared:credit\" computes \"return input+7\"\n" +
		"activity Run(Integer) -> Integer id \"urn:shared:run\" computes \"return input-3\"\n" +
		"bind Credit.result -> Run.input\n"
	definition := workspaceinput.Definition{Schema: workspaceinput.Schema,
		Entry:    packageruntime.EntrySpec{PackagePath: "shared/app", Activity: "Run"},
		Packages: []workspaceinput.Package{{Path: "shared/app", Sources: []string{"shared.gooo"}}}}
	request := workspaceinput.SplitRequest{PackagePath: "shared/app", Source: "shared.gooo",
		BeforeID: "urn:shared:run", Into: "run.gooo"}
	return fstest.MapFS{"shared.gooo": {Data: []byte(source)}}, definition, request
}
