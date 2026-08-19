package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"reflect"
	"testing"
)

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
