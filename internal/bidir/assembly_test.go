package bidir

import (
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func assemblyDocument(t *testing.T, source []byte) Document {
	t.Helper()
	file, diagnostics := syntax.Parse(string(source))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func TestAssemblySurvivesBXLawsAndCoreSemanticIdentity(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	testAssemblySurvivesBXLawsAndCoreSemanticIdentity(t, source)
	checkpoint := strings.Replace(string(source), "assembling {", `assembling {
    baseline "return input"
    picked "base-local" -> "reference_second"
    picked "condition-branches" -> "layout_forward"
    picked "independent-declarations" -> "schedule_reverse"`, 1)
	testAssemblySurvivesBXLawsAndCoreSemanticIdentity(t, []byte(checkpoint))
}

func TestSourceIRSearchSurvivesSemanticLoweringAndBindsHoldout(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/ir-search-source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	document := assemblyDocument(t, source)
	core, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	var activityNode semantic.Node
	foundActivity := false
	for _, node := range core.Graph.Nodes() {
		if node.Name == "ClampNegativeToZero" {
			activityNode = node
			foundActivity = true
			break
		}
	}
	if !foundActivity || activityNode.Assembly == nil || activityNode.Assembly.Search == nil ||
		activityNode.Assembly.Search.Grammar != "integer-offset-constant/v1" ||
		len(activityNode.Assembly.Cases) != 5 || len(activityNode.Assembly.HoldoutCases) != 2 {
		t.Fatalf("typed source search contract did not survive semantic lowering: %#v", activityNode)
	}
	changed := assemblyDocument(t, []byte(strings.Replace(string(source),
		`holdout_case "-9223372036854775808" -> "0"`,
		`holdout_case "-9223372036854775808" -> "1"`, 1)))
	changedCore, err := LowerDocument(changed)
	if err != nil || core.StableHash() == changedCore.StableHash() {
		t.Fatal("source-declared holdout edit did not change semantic identity", err)
	}
	activityNode.Assembly.Search.Intent = "detached edit"
	activityNode.Assembly.HoldoutCases[0].Expected = 99
	fresh, found := core.Graph.Node(activityNode.ID)
	if !found || fresh.Assembly.Search.Intent == "detached edit" || fresh.Assembly.HoldoutCases[0].Expected == 99 {
		t.Fatal("graph lookup exposed mutable source-search contract storage")
	}
}

func testAssemblySurvivesBXLawsAndCoreSemanticIdentity(t *testing.T, source []byte) {
	t.Helper()
	document := assemblyDocument(t, source)
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	roundtrip, err := Put(document, model)
	if err != nil {
		t.Fatal(err)
	}
	getPut, err := Get(roundtrip)
	if err != nil || !SemanticEquivalent(model, getPut) {
		t.Fatal("assembly violated Get-Put", err)
	}
	core, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []struct{ from, to string }{
		{`case "5" -> "24"`, `case "5" -> "25"`},
		{`attempts "8"`, `attempts "7"`},
		{`alternative "second"`, `alternative "first"`},
		{`두 번째 지역값`, `첫 번째 지역값`},
	} {
		changed := assemblyDocument(t, []byte(strings.Replace(string(source), mutation.from, mutation.to, 1)))
		changedModel, err := Get(changed)
		if err != nil {
			t.Fatal(err)
		}
		changedCore, err := LowerDocument(changed)
		if err != nil || core.StableHash() == changedCore.StableHash() ||
			SemanticFingerprint(model) == SemanticFingerprint(changedModel) {
			t.Fatal("assembly contract edit lost semantic identity", mutation, err)
		}
		putGet, err := Put(document, changedModel)
		if err != nil {
			t.Fatal(err)
		}
		restored, err := Get(putGet)
		if err != nil || !SemanticEquivalent(changedModel, restored) {
			t.Fatal("assembly violated Put-Get", err)
		}
	}
	nodes := core.Graph.Nodes()
	for _, node := range nodes {
		if node.Name == "Qualified" {
			if node.Assembly == nil || len(node.Assembly.Choices) != 3 || len(node.Assembly.Cases) != 5 {
				t.Fatal("core IR dropped the typed assembly", node)
			}
			node.Assembly.Choices[0].Intent = "changed"
			if len(node.Assembly.Picked) != 0 {
				node.Assembly.Picked[0].Label = "reference_first"
				fresh, _ := core.Graph.Node(node.ID)
				if fresh.Assembly.Picked[0].Label == "reference_first" {
					t.Fatal("graph lookup exposes mutable checkpoint storage")
				}
			}
			fresh, _ := core.Graph.Node(node.ID)
			if fresh.Assembly.Choices[0].Intent == "changed" {
				t.Fatal("graph lookup exposes mutable assembly storage")
			}
		}
	}
}
