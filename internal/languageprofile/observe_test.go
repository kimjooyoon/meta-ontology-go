package languageprofile

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

type fixedMeasurer struct{ sequence int }

func (value *fixedMeasurer) Measure(run func() sourceexecution.Receipt) (sourceexecution.Receipt, Measurement) {
	value.sequence++
	return run(), Measurement{WallNanoseconds: int64(100 + value.sequence), TotalAllocBytes: uint64(1000 + value.sequence)}
}

const profileFixture = "package billing\nnamespace billing\nentity Order id \"billing://order\"\nentity Receipt id \"billing://receipt\"\nactivity PayOrder(Order) -> Receipt\n"

func TestObserveProfilesDeterministicExecutionWithRunnerMeasurements(t *testing.T) {
	receipt := Observe(Request{Filename: "billing.gooo", Source: profileFixture, Entry: "PayOrder", Samples: 3}, &fixedMeasurer{})
	if err := Validate(receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "PASS" || receipt.Resolution != RunnerScopedResolution ||
		receipt.Summary.SamplesObserved != 3 || receipt.Summary.ExecutionDigestVariants != 1 ||
		receipt.Summary.WallObservations != 3 || receipt.Summary.AllocationObservations != 3 {
		t.Fatalf("receipt=%#v", receipt)
	}
}

func TestObserveRejectsUnknownEntry(t *testing.T) {
	receipt := Observe(Request{Filename: "billing.gooo", Source: profileFixture, Entry: "Missing", Samples: 3}, &fixedMeasurer{})
	if err := Validate(receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "FAIL_CLOSED" || receipt.Resolution != "EXACT" || receipt.Reason != "SOURCE_ENTRY_UNKNOWN" {
		t.Fatalf("receipt=%#v", receipt)
	}
}
