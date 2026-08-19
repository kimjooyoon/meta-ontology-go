package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"reflect"
	"testing"
)

func TestProductionBridgePreparationHead(t *testing.T) {
	expected := literalProductionCorpusExpectations()
	if len(expected) != len(testCorpus()) {
		t.Fatalf("literal producer corpus has %d rows, want %d", len(expected), len(testCorpus()))
	}
	for index, row := range testCorpus() {
		want, ok := expected[row.Name]
		if !ok {
			t.Fatalf("missing literal producer expectation for %s", row.Name)
		}
		input := detectorInputFromCanonical(row.Input)
		authority := detectorAuthorityForCorpus(index)
		got := productionVectorFromResult(production.Evaluate(input, authority), input, authority)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s producer vector=%+v want=%+v detector_head=%s", row.Name, got, want, detectorAuthorityHead)
		}
	}
}
func TestProductionBridgeRejectsProducerOnlyResultMutations(t *testing.T) {
	row := testCorpus()[0]
	input := detectorInputFromCanonical(row.Input)
	authority := detectorAuthorityForCorpus(0)
	base := production.Evaluate(input, authority)
	for _, mutation := range literalResultMutations(base, input, authority) {
		gotResult := base
		mutation.mutate(&gotResult)
		got := productionVectorFromResult(gotResult, input, authority)
		if !reflect.DeepEqual(got, mutation.observed) {
			t.Errorf("producer-only %s exact observed vector=%+v want=%+v", mutation.name, got, mutation.observed)
		}
		if reflect.DeepEqual(got, mutation.truth) {
			t.Errorf("producer-only %s mutation was not rejected: vector=%+v", mutation.name, got)
		}
	}
}
func TestProductionBridgeRejectsProducerInputMutations(t *testing.T) {
	row := testCorpus()[0]
	authority := detectorAuthorityForCorpus(0)
	for _, mutation := range literalInputMutations(detectorInputFromCanonical(row.Input), authority) {
		input := cloneProductionInput(mutation.input)
		got := production.Evaluate(input, authority)
		vector := productionVectorFromResult(got, input, authority)
		if !reflect.DeepEqual(vector, mutation.want) {
			t.Errorf("producer input %s vector=%+v want=%+v", mutation.name, vector, mutation.want)
		}
		if !reflect.DeepEqual(authority, mutation.authorityBefore) {
			t.Errorf("producer input %s changed evaluator authority", mutation.name)
		}
	}
}
