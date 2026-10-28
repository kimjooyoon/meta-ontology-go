package bidir

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestModelValidationUsesBoundRecordRegistryWithoutMutation(t *testing.T) {
	support, _ := syntax.EntityFieldsSupportForProfileID("gooo.entityfields.go-projection.v7")
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("records.gooo", recordReferenceSource, support)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	document, err := DocumentFromSyntaxWithEntityFieldsSupport(file, support)
	if err != nil {
		t.Fatal(err)
	}
	model, err := GetWithEntityFieldsSupport(document, support)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(model)
	if err := model.ValidateWithEntityFieldsSupport(support); err != nil {
		t.Fatal("source-bound nominal model was rejected", err)
	}
	if err := model.ValidateWithEntityFieldsSupport(EntityFieldsV6Support()); err == nil {
		t.Fatal("record-bearing model was accepted under the older profile")
	}
	support.Profile.Digest = "sha256:unbound"
	if err := model.ValidateWithEntityFieldsSupport(support); err == nil {
		t.Fatal("unbound profile was accepted")
	}
	after, _ := json.Marshal(model)
	if !bytes.Equal(before, after) {
		t.Fatal("profile validation changed the model")
	}
}
