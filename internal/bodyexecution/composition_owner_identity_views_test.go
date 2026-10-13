package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestSourceOwnerIdentityViewsAndNative3001(t *testing.T) {
	source := []byte(`package owner3001
namespace owner3001
entity Ticket id "OWNER3001://Entities/Quantity" fields {
 field units id "owner3001://fields/Units" type "urn:gooo:type:integer" required one
 field active id "urn:owner3001:active" type "urn:gooo:type:boolean" required one
}
entity Integer id "urn:gooo:type:integer"
activity Price(Ticket) -> Integer id "OWNER3001://Activities/Price" computes "if input.active && input.units < 101 { return input.units + 7 }; return input.units - 7"
`)
	before := append([]byte(nil), source...)
	file, diagnostics := bodycodegen.ParseBodyFile("owner3001.gooo", source)
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
		t.Fatal("source Get-Put", err)
	}
	if err := bidir.CheckPutGetWithEntityFieldsSupport(document, model, support); err != nil {
		t.Fatal("semantic Put-Get", err)
	}
	wrong := model
	wrong.Nodes = append([]bidir.Node(nil), model.Nodes...)
	for i := range wrong.Nodes {
		if len(wrong.Nodes[i].Fields) != 0 {
			wrong.Nodes[i].Fields = append([]bidir.Field(nil), wrong.Nodes[i].Fields...)
			wrong.Nodes[i].Fields[0].Parent = "owner3001://entities/Other"
		}
	}
	if _, err := bidir.PutWithEntityFieldsSupport(document, wrong, support); err == nil {
		t.Fatal("different URI parent accepted")
	}
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Price":{"units":13,"active":true}},"expected":{"Price":20}},{"inputs":{"Price":{"units":109,"active":false}},"expected":{"Price":102}}]}`)
	prior, err := GenerateComposition(context.Background(), "owner3001.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(context.Background(), "owner3001.gooo", source, prior, suite, nativeTool())
	if output := os.Getenv("GOOO_OWNER_ORIGIN_OUTPUT"); output != "" {
		data, _ := json.MarshalIndent(map[string]any{"source": string(source), "runtime": run, "get_put": "PASS", "put_get": "PASS", "wrong_parent_declined": true}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "owner3001.json"), append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err != nil || run.FinitePassed != 2 || len(run.ConditionSites) != 1 {
		t.Fatal("fresh owner execution", err, run.FinitePassed)
	}
	anchor := run.ConditionSites[0].Source
	if anchor == nil || anchor.Status != "MAPPED" || len(anchor.Fields) != 2 || anchor.Declaration == nil || !strings.Contains(prior.GoooSource[anchor.Declaration.Start:anchor.Declaration.End], "OWNER3001://Activities/Price") {
		t.Fatal("activity source spelling lost", anchor)
	}
	for _, field := range anchor.Fields {
		if field.Declaration == nil || field.Declaration.StableID != field.FieldID || !strings.Contains(prior.GoooSource[field.Declaration.Start:field.Declaration.End], field.FieldID) {
			t.Fatal("field declaration origin lost", field)
		}
		id, err := semantic.ParseIdentity(field.EntityID)
		if err != nil || id.String() != "owner3001://entities/Quantity" {
			t.Fatal("owner URI identity differs", id, err)
		}
	}
	if !bytes.Equal(source, before) || run.ModelCalls != 0 {
		t.Fatal("source or model behavior changed")
	}
	t.Log("owner3001: source spelling retained; canonical URI ownership; different parent declined; Get-Put/Put-Get; field origins2/2; native2/2; model0")
}
