package symbolicinvocationusecase

import "testing"

func TestSymbolicReaderObservationFailsClosed(t *testing.T) {
	tests := []struct {
		name        string
		reason      string
		expectedSHA string
		mutate      func(*ReaderRequestResultInput)
	}{
		{"unknown decision", "READER_REQUEST_DECISION_NOT_EXPLICIT_PASS", readerObservationTestSHA, func(input *ReaderRequestResultInput) {
			input.Decision = "UNKNOWN"
		}},
		{"unsafe write", "READER_REQUEST_UNSAFE_EFFECTS", readerObservationTestSHA, func(input *ReaderRequestResultInput) {
			input.Effects.RepositoryWrites = 1
		}},
		{"resolution mismatch", "READER_REQUEST_RESOLUTION_MISMATCH", readerObservationTestSHA, func(input *ReaderRequestResultInput) {
			input.View.EffectiveResolution = "FULL_DETAIL"
		}},
		{"subject mismatch", "READER_REQUEST_SUBJECT_MISMATCH", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", func(*ReaderRequestResultInput) {}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := readerObservationFixture()
			test.mutate(&input)
			report := EvaluateSymbolicReaderRequest(test.expectedSHA, readerObservationJSON(t, input))
			if report.Decision != "FAIL_CLOSED" || report.Resolution != "FAIL_CLOSED" || report.Reason != test.reason {
				t.Fatalf("decision=%s resolution=%s reason=%s", report.Decision, report.Resolution, report.Reason)
			}
			if report.Coordinates.Satisfied != 9 || report.Coordinates.Total != SymbolicReaderObservationTotal {
				t.Fatalf("coordinates=%+v", report.Coordinates)
			}
		})
	}
}
