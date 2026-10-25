package workspaceinput_test

import (
	"context"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func TestNativeSplitPreservesAuthoredPathWhileResolvingItsSource(t *testing.T) {
	definition, files, request := nativeSplitFixture()
	definition.Packages[0].Sources[0] = "./main.gooo"
	request.Source = "nested/../main.gooo"
	proposal, err := workspaceinput.PrepareSplit(files, "workspace.json", definition, request)
	if err != nil {
		t.Fatal(err)
	}
	if proposal.Result.InputFilename != "main.gooo" ||
		!reflect.DeepEqual(proposal.Definition.Packages[0].Sources, []string{"./main.gooo", "run.gooo"}) ||
		!reflect.DeepEqual(definition.Packages[0].Sources, []string{"./main.gooo"}) {
		t.Fatal("source resolution changed the authored path or retained a noncanonical identity")
	}
}

func TestNativeSplitRejectsOutputDefinitionThatCannotBeReloaded(t *testing.T) {
	definition, files, request := nativeSplitFixture()
	// A virtual reader permits long names; the transport has its own byte bound.
	request.Into = strings.Repeat("x", workspaceinput.MaxDefinitionBytes) + ".gooo"
	proposal, err := workspaceinput.PrepareSplit(files, "workspace.json", definition, request)
	if err == nil {
		raw, _ := json.Marshal(proposal.Definition)
		_, reloadErr := workspaceinput.Decode(raw)
		t.Fatalf("accepted an unreloadable proposal: %d definition bytes, reload error: %v", len(raw), reloadErr)
	}
	if !reflect.DeepEqual(proposal, workspaceinput.SplitProposal{}) {
		t.Fatal("output transport refusal returned an installable partial proposal")
	}
	if !strings.Contains(err.Error(), "encode workspace manifest") {
		t.Fatal("fixture did not reach the output definition's transport limit", err)
	}
}

func TestSharedSourceReloadPreservesTheOtherPackageEntry(t *testing.T) {
	definition, files, _ := loadImportedNativeFixture(t)
	other := definition.Packages[1]
	other.Path, other.Sources = "shared/other", []string{"./app.gooo"}
	definition.Packages = append(definition.Packages, other)
	after := reloadNativeSplitFixture(t, files, definition)
	if len(after.Packages[2].Sources) != 2 {
		t.Fatal("shared package source list was not updated")
	}
	after.Entry.PackagePath = "shared/other"
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"shared/other:Credit":5},"expected":{"shared/other:Credit":12,"shared/other:Run":41}},
 {"inputs":{"shared/other:Credit":-7},"expected":{"shared/other:Credit":0,"shared/other:Run":29}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := workspaceexecution.ExecuteWorkspace(context.Background(), after, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Runtime
	if r.FinitePassed != 4 || r.FiniteTotal != 4 || !r.RuntimeReplayed || r.ModelCalls != 0 {
		t.Fatal("other shared package lost observable behavior", r)
	}
}
