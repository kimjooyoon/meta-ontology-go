package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestEntityFieldsProjectionRejectsMalformedPartitionsAndUnknownState(t *testing.T) {
	baseIR, baseModel := cliEntityFieldsFixture(t)
	supported := syntax.CurrentEntityFieldsSupport()
	supported.State = syntax.EntityFieldsSupported
	for _, testCase := range entityFieldsProjectionMalformedCases() {
		t.Run(testCase.name, func(t *testing.T) {
			ir, model := baseIR, baseModel
			normalized, err := ir.Normalized()
			if err != nil {
				t.Fatal(err)
			}
			ir = normalized
			model = model.Clone()
			support := supported
			testCase.edit(&ir, &model, &support)
			projected, err := projectionIRFromBidirModelWithSupport(ir, model, support)
			if err == nil || !strings.Contains(strings.ToUpper(err.Error()), strings.ToUpper(testCase.want)) || !reflect.DeepEqual(projected, generator.SemanticIR{}) {
				t.Fatalf("partition result = %#v err=%v, want error containing %q and empty projection", projected, err, testCase.want)
			}
		})
	}
}
