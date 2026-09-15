package symbolicinvocationusecase

import (
	"encoding/json"
	"testing"
)

const readerObservationTestSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func readerObservationFixture() ReaderRequestResultInput {
	return ReaderRequestResultInput{
		Schema:     SymbolicReaderRequestResultSchema,
		SubjectSHA: readerObservationTestSHA,
		MetricID:   SymbolicReaderRequestResultMetric,
		Decision:   "PASS",
		Resolution: "GOOO_REQUEST_BOUND_ONLY",
		Reason:     "CANONICAL_GOOO_READER_REQUEST_BOUND",
		Request: ReaderRequestSelectionInput{
			Audience:           "USER",
			ExpectedResolution: "DECISION_AND_COUNTS_ONLY",
		},
		View: ReaderRequestViewInput{
			Audience:            "USER",
			SourceResolution:    "USER_VISIBLE",
			EffectiveResolution: "DECISION_AND_COUNTS_ONLY",
			IndicatorIDs:        []string{"compiler.defense-only-default-policies", "guardrail.reachable-default-policies"},
			Coordinates:         ReaderObservationCoordinates{Satisfied: 2, Total: 2, BasisPoints: 10000},
		},
		Coordinates:        ReaderObservationCoordinates{Satisfied: 12, Total: 12, BasisPoints: 10000},
		Effects:            ReaderObservationEffects{},
		PromotionCreditBPS: 0,
		Digest:             "sha256:8a7ba0f81bececfbbced1e9daa945c497649fea470e22efb95176358b28698ff",
	}
}

func readerObservationJSON(t *testing.T, input ReaderRequestResultInput) []byte {
	t.Helper()
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
