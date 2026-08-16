//go:build detector_bridge

package coupling

import (
	"reflect"
	"testing"

	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
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

func TestProductionBridgeExpectedOnlyMutationIsolation(t *testing.T) {
	row := testCorpus()[0]
	authority := detectorAuthorityForCorpus(0)
	producerExpected := literalProductionCorpusExpectations()
	oracleExpected := literalOracleBridgeExpectations()
	producerBeforeInput := detectorInputFromCanonical(row.Input)
	producerBefore := productionVectorFromResult(production.Evaluate(cloneProductionInput(producerBeforeInput), authority), producerBeforeInput, authority)
	oracleBefore := oracleBridgeVectorFromResult(Evaluate(cloneInput(row.Input)))
	mutatedProducer := producerExpected[row.Name]
	mutatedProducer.Decision = DecisionFailClosed
	mutatedProducer.Reasons = []production.Reason{{Code: production.ReasonCode("EXPECTED_ONLY"), Detail: "fixture mutation"}}
	mutatedProducer.Observation.CPU.Value = 999
	producerExpected[row.Name] = mutatedProducer
	mutatedOracle := oracleExpected[row.Name]
	mutatedOracle.Decision = DecisionFailClosed
	mutatedOracle.Reason = ReasonInputAmbiguous
	mutatedOracle.CanonicalOutputDigest = bridgeHash("expected-only-output")
	oracleExpected[row.Name] = mutatedOracle
	producerAfterInput := detectorInputFromCanonical(row.Input)
	producerAfter := productionVectorFromResult(production.Evaluate(cloneProductionInput(producerAfterInput), authority), producerAfterInput, authority)
	oracleAfter := oracleBridgeVectorFromResult(Evaluate(cloneInput(row.Input)))
	if !reflect.DeepEqual(producerBefore, producerAfter) || !reflect.DeepEqual(oracleBefore, oracleAfter) {
		t.Fatalf("expected-only fixture mutation affected subject output: before=%+v after=%+v", producerBefore, producerAfter)
	}
	if reflect.DeepEqual(producerExpected[row.Name], producerBefore) || reflect.DeepEqual(oracleExpected[row.Name], oracleBefore) {
		t.Fatal("expected-only mutation did not mutate fixture expectation")
	}
	presentation := cloneInput(row.Input)
	presentation.FixtureID = "fixture-label/bridge-mutated"
	presentation.Registry[0].PackageLabel = "renamed-package"
	presentation.Registry[0].FileLabel = "renamed.go"
	presentation.Registry[0].SourceSpan = "99:1-99:2"
	presentationInput := detectorInputFromCanonical(presentation)
	presentationVector := productionVectorFromResult(production.Evaluate(presentationInput, authority), presentationInput, authority)
	if !reflect.DeepEqual(presentationVector, producerBefore) {
		t.Fatalf("presentation-only producer packet mutation changed authoritative vector: got=%+v want=%+v", presentationVector, producerBefore)
	}
}

func TestProductionBridgeSubjectsAreReadOnlyAndCrossChecked(t *testing.T) {
	oracleExpected := literalOracleBridgeExpectations()
	producerExpected := literalProductionCorpusExpectations()
	for index, row := range testCorpus() {
		canonical := cloneInput(row.Input)
		before := cloneInput(canonical)
		oracle := Evaluate(canonical)
		if !reflect.DeepEqual(canonical, before) {
			t.Fatalf("%s independent oracle mutated canonical input", row.Name)
		}
		oracleVector := oracleBridgeVectorFromResult(oracle)
		oracleWant, ok := oracleExpected[row.Name]
		if !ok {
			t.Fatalf("missing literal independent oracle expectation for %s", row.Name)
		}
		if !reflect.DeepEqual(oracleVector, oracleWant) {
			t.Fatalf("%s independent oracle complete vector=%+v want=%+v", row.Name, oracleVector, oracleWant)
		}
		input := detectorInputFromCanonical(row.Input)
		authority := detectorAuthorityForCorpus(index)
		producerBefore := productionInputSnapshot(input)
		producerResult := production.Evaluate(input, authority)
		if productionInputSnapshot(input) != producerBefore {
			t.Fatalf("%s producer mutated its input packet", row.Name)
		}
		producerVector := productionVectorFromResult(producerResult, input, authority)
		producerWant, ok := producerExpected[row.Name]
		if !ok {
			t.Fatalf("missing literal producer expectation for %s", row.Name)
		}
		pair := bridgeSubjectVector{Oracle: oracleVector, Producer: producerVector}
		wantPair := bridgeSubjectVector{Oracle: oracleWant, Producer: producerWant}
		if !reflect.DeepEqual(pair, wantPair) {
			t.Fatalf("%s complete oracle/producer pair=%+v want=%+v", row.Name, pair, wantPair)
		}
		if row.Name == "positive-no-write" {
			if len(producerResult.AcceptedSurfaceIDs) != 0 || producerResult.Observation.ChangedSurfaces.Value != 0 || producerResult.Observation.Receipts.Value != 0 {
				t.Fatalf("%s producer wrote a non-empty result for explicit zero-change manifest: %+v", row.Name, producerResult)
			}
		}
	}
}

// The projection below is run once, before any producer-only mutation. It
// never receives a mutated producer packet and therefore cannot repair one.
