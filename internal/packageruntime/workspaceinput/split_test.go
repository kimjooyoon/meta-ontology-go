package workspaceinput_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func nativeSplitFixture() (workspaceinput.Definition, fstest.MapFS, workspaceinput.SplitRequest) {
	definition := workspaceinput.Definition{Schema: workspaceinput.Schema,
		Entry:    packageruntime.EntrySpec{PackagePath: "native/p", Activity: "Run"},
		Packages: []workspaceinput.Package{{Path: "native/p", Sources: []string{"main.gooo"}}}}
	files := fstest.MapFS{"main.gooo": {Data: []byte("package p\nnamespace p\n" +
		"entity Integer id \"urn:gooo:type:integer\"\n" +
		"activity Keep(Integer) -> Integer id \"urn:native:keep\" computes \"return input+13\"\n" +
		"activity Run(Integer) -> Integer id \"urn:native:run\" computes \"return input-4\"\n" +
		"bind Keep.result -> Run.input\n")}}
	return definition, files, workspaceinput.SplitRequest{
		PackagePath: "native/p", Source: "main.gooo", BeforeID: "urn:native:run", Into: "run.gooo"}
}

func TestNativeSplitProposalDoesNotShareAuthoredSourceStorage(t *testing.T) {
	definition, files, request := nativeSplitFixture()
	backing := []string{"main.gooo", "retained-capacity"}
	definition.Packages[0].Sources = backing[:1]
	proposal, err := workspaceinput.PrepareSplit(files, "workspace.json", definition, request)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || string(files["main.gooo"].Data) == proposal.Result.Files[0].Source {
		t.Fatal("read-only preparation installed a source change")
	}
	if backing[1] != "retained-capacity" {
		t.Fatal("append changed caller-owned backing storage")
	}
	proposal.Definition.Packages[0].Sources[0] = "caller-edit.gooo"
	if definition.Packages[0].Sources[0] != "main.gooo" {
		t.Fatal("returned file list aliases authored input")
	}
}

func TestNativeSplitRefusalReturnsNoInstallablePartialProposal(t *testing.T) {
	for _, mode := range []string{"missing input", "first declaration", "occupied source", "source count"} {
		t.Run(mode, func(t *testing.T) {
			definition, files, request := nativeSplitFixture()
			switch mode {
			case "missing input":
				definition.Packages[0].Sources = append(definition.Packages[0].Sources, "missing.gooo")
			case "first declaration":
				request.BeforeID = "urn:gooo:type:integer"
			case "occupied source":
				request.Into = "main.gooo"
			case "source count":
				for i := 1; i < workspaceinput.MaxSourceCount; i++ {
					name := fmt.Sprintf("part%03d.gooo", i)
					definition.Packages[0].Sources = append(definition.Packages[0].Sources, name)
					files[name] = &fstest.MapFile{Data: []byte("package p\nnamespace p\n")}
				}
			}
			proposal, err := workspaceinput.PrepareSplit(files, "workspace.json", definition, request)
			if err == nil || !reflect.DeepEqual(proposal, workspaceinput.SplitProposal{}) {
				t.Fatal(proposal, err)
			}
			if mode == "source count" && !strings.Contains(err.Error(), "split output exceeds") {
				t.Fatal("fixture did not reach the output source-set obligation", err)
			}
		})
	}
}
