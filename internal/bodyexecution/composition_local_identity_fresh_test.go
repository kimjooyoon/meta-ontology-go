package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestLocalIdentityContradictions10201(t *testing.T) {
	const source = `package locals10201
namespace locals10201
entity Integer id "urn:gooo:type:integer"
activity Transform(Integer) -> Integer id "urn:locals10201:transform" computes "let initial=input+137;let shifted=initial*139;if input>149 {shifted=shifted-151};return shifted"
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{
		{Inputs: map[string]json.RawMessage{"Transform": json.RawMessage(`163`)}, Expected: map[string]json.RawMessage{"Transform": json.RawMessage(`41549`)}},
	}}
	composition, err := GenerateCompositionWithOptions(context.Background(), "identity10201.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Transform"})
	if err != nil {
		t.Fatal(err)
	}
	sites, err := compositionLocalSites(composition)
	if err != nil || len(sites) != 3 {
		t.Fatal("fresh site binding", err, sites)
	}
	changes := map[string]func([]bodycodegen.BodyLocalOrigin) []bodycodegen.BodyLocalOrigin{
		"reassignment-to-other-binding": func(o []bodycodegen.BodyLocalOrigin) []bodycodegen.BodyLocalOrigin {
			o[2].VariableOrdinal = o[0].VariableOrdinal
			return o
		},
		"two-declarations-one-identity": func(o []bodycodegen.BodyLocalOrigin) []bodycodegen.BodyLocalOrigin {
			o[1].VariableOrdinal = o[0].VariableOrdinal
			return o
		},
		"missing-native-write":   func(o []bodycodegen.BodyLocalOrigin) []bodycodegen.BodyLocalOrigin { o[2].NativeOrdinal = 99; return o },
		"duplicate-native-write": func(o []bodycodegen.BodyLocalOrigin) []bodycodegen.BodyLocalOrigin { return append(o, o[0]) },
	}
	rejections := map[string]string{}
	for name, change := range changes {
		altered := composition
		altered.Steps = append([]CompositionStep(nil), composition.Steps...)
		altered.Steps[0].Generation.Report.BodyLocals = change(append([]bodycodegen.BodyLocalOrigin(nil), composition.Steps[0].Generation.Report.BodyLocals...))
		if _, err := compositionLocalSites(altered); err == nil {
			t.Fatal("contradiction accepted", name)
		} else {
			rejections[name] = err.Error()
		}
	}
	saveReadObservation(t, "local-identity10201.json", map[string]any{"source": source,
		"composition": composition, "sites": sites, "rejections": rejections, "native_executions": 0})
	t.Log("three fresh write sites; four lexical/projection contradictions rejected; no native execution")
}
