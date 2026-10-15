package replay

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestOrderedCollectionSyntax8429(t *testing.T) {
	results := map[string]Result{}
	for _, scalar := range []string{"integer", "string", "boolean"} {
		for _, presence := range []string{"required", "optional"} {
			name := scalar + "_" + presence
			source := fmt.Sprintf(`package syntax8429
namespace syntax8429_%s
entity Readings id "urn:syntax8429:%s:readings" fields {
 field values id "urn:syntax8429:%s:values" type %s %s many
}
activity Receive(Readings) -> Readings id "urn:syntax8429:%s:receive" computes "return input"
activity Deliver(Readings) -> Readings id "urn:syntax8429:%s:deliver" computes "return input"
bind Receive.result -> Deliver.input
`, name, name, name, scalar, presence, name, name)
			files := fstest.MapFS{"readings.gooo": &fstest.MapFile{Data: []byte(source)}}
			result := ExecuteWithOrderedCollections(files, "readings.gooo", "VALID", "")
			if result.ObservedDecision != DecisionPass || !result.ASTReplayed || !result.ByteReplayed ||
				!result.SemanticReplayed || !result.GetPut || !result.PutGet {
				t.Fatalf("%s V5 syntax and BX: %+v", name, result)
			}
			if legacy := ExecuteWithEntityFieldsSupport(files, "readings.gooo", "VALID", ""); legacy.ObservedDecision == DecisionPass {
				t.Fatalf("%s silently entered the V1 contract", name)
			}
			results[name] = result
		}
	}
	if output := os.Getenv("GOOO_HANDOFF_OBSERVATION"); output != "" {
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(output, "syntax8429.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("8429: six fresh ordered-field syntax programs preserve AST/format/semantics/GetPut/PutGet under V5; V1 refuses all six; no native/model execution")
}
