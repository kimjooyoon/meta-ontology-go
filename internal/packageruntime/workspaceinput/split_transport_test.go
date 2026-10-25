package workspaceinput_test

import (
	"encoding/json"
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
