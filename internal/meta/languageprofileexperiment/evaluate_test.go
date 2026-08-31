package languageprofileexperiment

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/languageprofile"
	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

type experimentMeasurer struct{ sequence int }

func (value *experimentMeasurer) Measure(run func() sourceexecution.Receipt) (sourceexecution.Receipt, languageprofile.Measurement) {
	value.sequence++
	return run(), languageprofile.Measurement{WallNanoseconds: int64(100 + value.sequence), TotalAllocBytes: uint64(1000 + value.sequence)}
}

const experimentSource = "package billing\nnamespace billing\nentity Order id \"billing://order\"\nentity Receipt id \"billing://receipt\"\nactivity PayOrder(Order) -> Receipt\n"

func validInput() Input {
	request := languageprofile.Request{Filename: "billing.gooo", Source: experimentSource, Entry: "PayOrder", Samples: 5}
	unknown := request
	unknown.Entry = "Missing"
	return Input{SubjectSHA: strings.Repeat("a", 40), ExecutableDigest: "sha256:" + strings.Repeat("b", 64),
		Contract: ExpectedContract(), First: languageprofile.Observe(request, &experimentMeasurer{}),
		Replay:       languageprofile.Observe(request, &experimentMeasurer{}),
		UnknownEntry: languageprofile.Observe(unknown, &experimentMeasurer{})}
}

func TestEvaluateClosesFixedProfileExperiment(t *testing.T) {
	report := Evaluate(validInput())
	if report.Decision != "PASS" || report.Summary.Coordinates.Satisfied != ExpectedIndicators ||
		report.Summary.Samples != 10 || report.Summary.ExecutionDigestVariants != 1 ||
		report.Summary.UnknownEntryRejections != 1 || report.Summary.Unknowns != 0 {
		t.Fatalf("report=%#v", report)
	}
}

func TestEvaluateUnknownDecisionLowersResolution(t *testing.T) {
	input := validInput()
	input.First.Decision = "UNKNOWN"
	report := Evaluate(input)
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "LOWER_RESOLUTION" ||
		report.Summary.Coordinates.Satisfied != 0 || report.Summary.Unknowns != 1 {
		t.Fatalf("report=%#v", report)
	}
}
