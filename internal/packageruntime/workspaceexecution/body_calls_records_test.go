package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestWorkspacePureCallsRecordNamespacesAndValueCopies(t *testing.T) {
	m := pureCallWorkspace()
	m.Packages[0].Sources = m.Packages[0].Sources[:1]
	m.Packages[0].Sources[0].Content = `package core
namespace core
entity Integer id "calls://integer"
entity Result id "calls://core/result" fields { field value id "calls://core/value" type integer required one }
activity Bump(Result) -> Result computes "let next = input; next.value = input.value + 1; return next"
activity Make(Integer) -> Result computes "let original = Result{value: input}; let changed = Bump(original); return Result{value: changed.value + original.value}"
`
	m.Packages[1].Sources[0].Content = `package app
namespace app
import core "example/core"
entity Result id "calls://app/result" fields { field value id "calls://app/value" type integer required one }
activity Main(Integer) -> Result computes "let foreign = core.Make(input); return Result{value: foreign.value + 2}"
`
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"example/app:Main":3},"expected":{"example/app:Main":{"value":9}}},
{"inputs":{"example/app:Main":-3},"expected":{"example/app:Main":{"value":-3}}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), m, suite, "", "")
	if err != nil || result.Runtime.FinitePassed != 2 || len(result.Program.EntityAliases) != 2 {
		t.Fatal("called record types lost package identity or value semantics", err, result.Runtime)
	}
}

func TestWorkspacePureCallsInDeclaredFieldAlternative(t *testing.T) {
	m := pureCallWorkspace()
	m.Packages[1].Sources[0].Content = `package app
namespace app
import core "example/core"
entity Result id "calls://result" fields { field value id "calls://value" type integer required one }
activity Main(Integer) -> Result computes "return Result{value: input}" assembling {
    choice "value" field_value at "0" alternative "core.Scale(input)" intent "Apply the imported scale rule."
    value_case "[1]" -> "{\"value\":2}"
    value_case "[2]" -> "{\"value\":5}"
    attempts "2"
}
`
	suite := pureCallCases(t)
	suite.Cases = suite.Cases[:1]
	suite.Cases[0].Expected["example/app:Main"] = json.RawMessage(`{"value":11}`)
	result, err := ExecuteWorkspace(context.Background(), m, suite, "", "")
	if err != nil || result.Runtime.FinitePassed != 1 {
		t.Fatal("declared alternative could not call the imported helper", err, result.Runtime)
	}
	if result.Composition.Steps[0].Generation.Report.RecordAssembly.SelectedMask != 1 ||
		result.Program.PureCalls.Sites[0].Surface != "choice:value" {
		t.Fatal("choice call identity missing", result.Program.PureCalls)
	}
	if _, err := ReplayWorkspace(context.Background(), m, result, suite, ""); err != nil {
		t.Fatal("field-call selection is not replayable", err)
	}
}

func TestWorkspacePureCallsAliasSubstitutionKeepsTargetIdentity(t *testing.T) {
	m := pureCallWorkspace()
	before, err := Prepare(m)
	if err != nil {
		t.Fatal(err)
	}
	source := &m.Packages[1].Sources[0].Content
	*source = strings.Replace(*source, "import core ", "import rules ", 1)
	*source = strings.Replace(*source, "return core.Scale", "return rules.Scale", 1)
	after, err := Prepare(m)
	if err != nil || before.Source != after.Source {
		t.Fatal("import alias changed lowered behavior", err)
	}
	for i, activity := range before.PureCalls.Activities {
		other := after.PureCalls.Activities[i]
		if activity.ActivityID != other.ActivityID || activity.LoweredID != other.LoweredID {
			t.Fatal("import alias changed semantic identity", activity, other)
		}
	}
}
