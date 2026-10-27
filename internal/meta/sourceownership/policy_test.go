package sourceownership_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodylibrary"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourceownership"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func TestSourceOwnershipPolicyAndRegeneration(t *testing.T) {
	for _, row := range []struct {
		generated, mayRewrite bool
		reason                string
	}{
		{false, true, "HANDWRITTEN_GO_SOURCE"}, {true, false, "GENERATED_SOURCE_OWNER_REQUIRED"},
	} {
		got := sourceownership.Evaluate(row.generated)
		if got.F_may_rewrite_go != row.mayRewrite || got.F_reason != row.reason || got.F_rule_id != "gooo://source-ownership/resolve" {
			t.Fatal(got)
		}
		if row.generated && got.F_next_operation != "regenerate-source-owner" {
			t.Fatal(got)
		}
	}
	raw, err := os.ReadFile("gooo.workspace.json")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	definition, err := workspaceinput.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	inputs, err := workspaceinput.Load(fstest.MapFS{"policy.gooo": &fstest.MapFile{Data: source}}, "gooo.workspace.json", definition)
	if err != nil {
		t.Fatal(err)
	}
	program, err := workspaceexecution.Prepare(inputs)
	if err != nil {
		t.Fatal(err)
	}
	generated, err := bodylibrary.Generate(context.Background(), "workspace.gooo", []byte(program.Source), program.Entry.LoweredName, "sourceownership")
	if err != nil {
		t.Fatal(err)
	}
	stored, err := os.ReadFile("semantic.gooo.go")
	if err != nil || !strings.HasSuffix(string(stored), generated) {
		t.Fatal("regenerate Gooo source ownership policy", err)
	}
}
