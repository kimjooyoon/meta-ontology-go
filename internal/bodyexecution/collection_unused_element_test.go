package bodyexecution

import (
	"context"
	"testing"
)

func TestCollectionUnusedElementAndIndex7567(t *testing.T) {
	source := []byte(`package unused7567
namespace unused7567
entity Integer id "urn:gooo:type:integer"
entity Numbers id "urn:unused7567:numbers" fields {
 field items id "urn:unused7567:items" type integer required many
}
activity LastWithCount(Numbers) -> Integer id "urn:unused7567:last" computes "let n = 0; for ignored in input.items { n = n + 1 }; if n > 0 { return n + input.items[n - 1] }; return n"
`)
	suite, err := DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"LastWithCount":{"items":[4,7,9]}},"expected":{"LastWithCount":12}},
{"inputs":{"LastWithCount":{"items":[]}},"expected":{"LastWithCount":0}},
{"inputs":{"LastWithCount":{"items":[18]}},"expected":{"LastWithCount":19}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := GenerateComposition(context.Background(), "unused7567.gooo", source, suite, "")
	if err != nil {
		t.Fatal("unused iteration element generation", err)
	}
	run, err := ExecuteComposition(context.Background(), "unused7567.gooo", source, generated, suite, nativeTool())
	if err != nil || run.FinitePassed != 3 || run.FiniteTotal != 3 || run.ModelCalls != 0 {
		t.Fatal("unused iteration element native behavior", err, run)
	}
	t.Log("7567: unused element allowed without changing Gooo source; guarded collection index; native3/3 model0")
}
