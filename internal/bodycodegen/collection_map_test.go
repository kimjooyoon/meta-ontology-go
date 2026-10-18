package bodycodegen

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

const collectionMapFixture = `package mapping
namespace mapping
entity Text id "urn:gooo:type:string"
entity Paths id "urn:mapping:paths" fields {field values id "urn:mapping:values" type string required many}
activity Quote(Text) -> Text id "urn:mapping:quote" computes ` + "`return \"[\" + input + \"]\"`" + `
activity Project(Paths) -> Paths id "urn:mapping:project" computes ` + "`return Paths{values: map_values(input.values, Quote)}`" + "\n"

func TestCollectionMapGenerateAndInterpret(t *testing.T) {
	for _, name := range []string{"mapping", "map_prototype"} {
		t.Run(name, func(t *testing.T) {
			source := strings.ReplaceAll(collectionMapFixture, "mapping", name)
			testCollectionMapGeneration(t, source)
		})
	}
}

func testCollectionMapGeneration(t *testing.T, source string) {
	t.Helper()
	generated, err := Generate("mapping.gooo", []byte(source), "Project")
	if err != nil {
		t.Fatal(err)
	}
	if generated.Report.CallClosure == nil || generated.Report.CallClosure.MaxCallsPerInvocation != 256 ||
		len(generated.Report.CallClosure.Activities) != 1 || !strings.HasSuffix(generated.Report.CallClosure.Activities[0].ActivityID, ":quote") {
		t.Fatal("mapper must retain its source identity and 256-call bound", generated.Report.CallClosure)
	}
	for _, values := range [][]string{{}, {"same", "same", "한글"}, make([]string, 256)} {
		want := make([]string, len(values))
		for i, value := range values {
			want[i] = "[" + value + "]"
		}
		inputs, _ := json.Marshal([]map[string][]string{{"values": values}})
		expected, _ := json.Marshal(map[string][]string{"values": want})
		observed, err := evaluateRecordAssembly(context.Background(), []byte(generated.Source), "Project", generated.Report.RecordTypes,
			[]assemblyspec.ValueCase{{Inputs: string(inputs), Expected: string(expected)}})
		if err != nil || len(observed) != 1 || !observed[0].Passed {
			t.Fatal("mapped result differs from independent expectation", err, observed)
		}
	}
}
