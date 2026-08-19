package coupling

import (
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"reflect"
	"testing"
)

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
