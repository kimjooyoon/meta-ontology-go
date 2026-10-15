package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestSourceOwnerURISpellingAndNativeOrigin2867(t *testing.T) {
	source := []byte(`package owner2867
namespace owner2867
entity Packet id "OWNER2867://Domain/Packet" fields {
 field amount id "OWNER2867://Field/Amount" type "urn:gooo:type:integer" required one
}
entity Integer id "urn:gooo:type:integer"
activity Calculate(Packet) -> Integer id "OWNER2867://Activity/Calculate" computes "let offset = 3; if input.amount < 97 { return input.amount + offset }; return input.amount - offset"
`)
	before := append([]byte(nil), source...)
	file, diagnostics := bodycodegen.ParseBodyFile("owner2867.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	support := bodycodegen.BodyEntityFieldsSupport(file)
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, support)
	if err != nil {
		t.Fatal(err)
	}
	model, err := bidir.GetWithEntityFieldsSupport(document, support)
	if err != nil {
		t.Fatal(err)
	}
	if err := bidir.CheckGetPutWithEntityFieldsSupport(document, support); err != nil {
		t.Fatal("source-preserving Get-Put", err)
	}
	if err := bidir.CheckPutGetWithEntityFieldsSupport(document, model, support); err != nil {
		t.Fatal("semantic Put-Get", err)
	}
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Calculate":{"amount":8}},"expected":{"Calculate":11}},{"inputs":{"Calculate":{"amount":105}},"expected":{"Calculate":102}}]}`)
	prior, err := GenerateComposition(context.Background(), "owner2867.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(context.Background(), "owner2867.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 2 || len(run.ConditionSites) != 1 {
		t.Fatal("fresh URI owner source execution", err, run.FinitePassed)
	}
	anchor := run.ConditionSites[0].Source
	if anchor == nil || anchor.Status != "MAPPED" || len(anchor.Fields) != 1 || anchor.Fields[0].Declaration == nil || anchor.Declaration.StableID != run.ConditionSites[0].ActivityID || anchor.Fields[0].Declaration.StableID != "OWNER2867://Field/Amount" {
		t.Fatal("authored IDs changed", anchor)
	}
	id, err := semantic.ParseIdentity(anchor.Fields[0].EntityID)
	if err != nil || id.String() != "owner2867://domain/Packet" || !bytes.Equal(source, before) {
		t.Fatal("URI identity or authored source differs", id, err)
	}
	if output := os.Getenv("GOOO_OWNER_ORIGIN_OUTPUT"); output != "" {
		data, _ := json.MarshalIndent(map[string]any{"source": string(source), "runtime": run, "get_put": "PASS", "put_get": "PASS", "source_unchanged": true, "source_facing_ids_retained": true}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "owner2867.json"), append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("owner2867: URI owner identity recognized; source spelling/origin retained; Get-Put/Put-Get; native2/2; no model")
}
