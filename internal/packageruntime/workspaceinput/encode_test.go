package workspaceinput_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func TestWorkspaceDefinitionEncodingUsesTheReadersExactByteLimit(t *testing.T) {
	definition, _, _ := nativeSplitFixture()
	base, err := workspaceinput.Encode(definition)
	if err != nil {
		t.Fatal(err)
	}
	padding := workspaceinput.MaxDefinitionBytes - len(base)
	definition.Packages[0].Sources[0] = strings.Repeat("a", len("main")+padding) + ".gooo"
	raw, err := workspaceinput.Encode(definition)
	if err != nil || len(raw) != workspaceinput.MaxDefinitionBytes {
		t.Fatal("exact-boundary definition was not retained", len(raw), err)
	}
	reloaded, err := workspaceinput.Decode(raw)
	if err != nil || !reflect.DeepEqual(reloaded, definition) {
		t.Fatal("encoded boundary did not preserve authored values", err)
	}
	definition.Packages[0].Sources[0] = "a" + definition.Packages[0].Sources[0]
	if raw, err := workspaceinput.Encode(definition); err == nil || raw != nil {
		t.Fatal("over-boundary definition returned publishable bytes")
	}
}

func TestWorkspaceDefinitionEncodingRejectsChangedAuthoredValues(t *testing.T) {
	definition, _, _ := nativeSplitFixture()
	definition.Entry.Activity = string([]byte{0xff})
	if raw, err := workspaceinput.Encode(definition); err == nil || raw != nil {
		t.Fatal("JSON UTF-8 replacement silently changed the authored activity")
	}
}
