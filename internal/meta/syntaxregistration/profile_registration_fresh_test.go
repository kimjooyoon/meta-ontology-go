package syntaxregistration

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestExplicitProfileNativeProjectionFresh1843(t *testing.T) {
	const path = "native-role1843.go"
	input := map[string][]byte{path: []byte(`package languagesyntax
func expectedRegistry() Registry { return Registry{Cases: []CaseDefinition{{ID: "retained1843"}}} }
`)}
	source, err := parseSourceUnits(input, "./", "languagesyntax", false)
	if err != nil {
		t.Fatal(err)
	}
	source.activity = "gooo://profile1843/generate-native"
	request := Request{Case: languagesyntax.CaseDefinition{ID: "new1843", Path: "examples/profile1843/main.gooo",
		ProofChoice: "COHERENCE", MetaOperation: "replay-language-syntax", EntityFieldsProfile: syntax.EntityFieldsV2ProfileID}}
	if err := generateRegistry(source, request); err != nil {
		t.Fatal(err)
	}
	outputs, err := source.finish()
	if err != nil || len(outputs) != 1 || !strings.Contains(string(outputs[path]), "retained1843") ||
		!strings.Contains(string(outputs[path]), `EntityFieldsProfile: "gooo.entityfields.go-projection.v2"`) {
		t.Fatalf("explicit source profile not projected: %v %v", err, outputs)
	}
}
