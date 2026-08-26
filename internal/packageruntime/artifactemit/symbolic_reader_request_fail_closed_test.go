package artifactemit

import "testing"

func TestCompileSymbolicReaderRequestFailsClosed(t *testing.T) {
	tests := []struct {
		name       string
		audience   string
		resolution string
		mutate     func(*SymbolicValueReaderProjection)
		reason     string
	}{
		{"unknown audience", "OBSERVER", "DECISION_AND_COUNTS_ONLY", nil, "GOOO_READER_REQUEST_AUDIENCE_UNKNOWN"},
		{"resolution mismatch", "USER", "SOURCE_BOUND_RECEIPT_ONLY", nil, "GOOO_READER_REQUEST_RESOLUTION_MISMATCH"},
		{"unknown source decision", "USER", "DECISION_AND_COUNTS_ONLY", func(value *SymbolicValueReaderProjection) { value.Decision = "UNKNOWN" }, "GOOO_READER_REQUEST_SOURCE_INVALID"},
		{"count mismatch", "USER", "DECISION_AND_COUNTS_ONLY", func(value *SymbolicValueReaderProjection) { value.Readers[0].Coordinates.Total-- }, "GOOO_READER_REQUEST_COUNT_MISMATCH"},
		{"mutation authority", "USER", "DECISION_AND_COUNTS_ONLY", func(value *SymbolicValueReaderProjection) { value.Effects.MutationAuthority = true }, "GOOO_READER_REQUEST_SOURCE_NOT_READ_ONLY"},
		{"duplicate indicators", "USER", "DECISION_AND_COUNTS_ONLY", duplicateSymbolicReaderRequestIndicator, "GOOO_READER_REQUEST_INDICATORS_NOT_UNIQUE"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			projection := symbolicReaderRequestProjectionFixture()
			if test.mutate != nil {
				test.mutate(&projection)
			}
			result, err := CompileSymbolicReaderRequest(
				symbolicReaderRequestSourceFixture(test.audience, test.resolution),
				encodeSymbolicReaderRequestProjection(t, projection), "fixture-sha",
			)
			if err != nil {
				t.Fatal(err)
			}
			if result.Decision != "FAIL_CLOSED" || result.Resolution != "INVARIANT_ONLY" ||
				result.Reason != test.reason || len(result.View.IndicatorIDs) != 0 {
				t.Fatalf("result=%+v", result)
			}
		})
	}
}

func duplicateSymbolicReaderRequestIndicator(value *SymbolicValueReaderProjection) {
	reader := &value.Readers[0]
	reader.IndicatorIDs = append(reader.IndicatorIDs, reader.IndicatorIDs[0])
	reader.Coordinates.Satisfied++
	reader.Coordinates.Total++
}
