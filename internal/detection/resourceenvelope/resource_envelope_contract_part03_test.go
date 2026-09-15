package resourceenvelope

import (
	"testing"
)

func TestResourceEnvelopeImplementationContract(t *testing.T) {
	envelope := Envelope{
		SchemaVersion: SchemaVersion, RunnerImageDigest: "runner-digest",
		AllocatedCPUCount: 2, WarmupCount: 1, SampleCount: ExpectedSampleCount,
		Limits: Limits{CPUCoreNS: 30, PeakRSSBytes: 500, ReadBytes: 500, WriteBytes: 500},
		Samples: []Sample{
			{CPUCoreNS: 999, WallNS: 999, PeakRSSBytes: 999, ReadBytes: 999, WriteBytes: 999},
			{CPUCoreNS: 30, WallNS: 300, PeakRSSBytes: 300, ReadBytes: 300, WriteBytes: 300},
			{CPUCoreNS: 10, WallNS: 100, PeakRSSBytes: 100, ReadBytes: 100, WriteBytes: 100},
			{CPUCoreNS: 50, WallNS: 500, PeakRSSBytes: 500, ReadBytes: 500, WriteBytes: 500},
			{CPUCoreNS: 20, WallNS: 200, PeakRSSBytes: 200, ReadBytes: 200, WriteBytes: 200},
			{CPUCoreNS: 40, WallNS: 400, PeakRSSBytes: 400, ReadBytes: 400, WriteBytes: 400},
		},
	}
	pass := Evaluate(envelope)
	if pass.Status != PASS || pass.FullSuiteRequired || pass.CPUCoreNS != 30 ||
		pass.CPUUtilizationPPM != 50000 || pass.PeakRSSBytes != 500 ||
		pass.ReadBytes != 500 || pass.WriteBytes != 500 {
		t.Fatalf("pass result = %#v", pass)
	}

	overrun := envelope
	overrun.Limits.CPUCoreNS = 29
	if result := Evaluate(overrun); result.Status != FAIL_CLOSED || result.FullSuiteRequired {
		t.Fatalf("exact CPU budget overrun result = %#v", result)
	}

	unknowns := []struct {
		name   string
		result Result
	}{
		{name: "malformed", result: EvaluateJSON([]byte(`{"unexpected":true}`))},
		{name: "missing", result: EvaluateJSON([]byte(`{"schema_version":"gooo/resource-envelope/v1"}`))},
		{name: "mismatched-count", result: Evaluate(Envelope{SchemaVersion: SchemaVersion, RunnerImageDigest: "runner-digest", AllocatedCPUCount: 1, WarmupCount: 1, SampleCount: 4, Samples: envelope.Samples})},
	}
	zeroDenominator := envelope
	zeroDenominator.Samples[1].WallNS = 0
	unknowns = append(unknowns, struct {
		name   string
		result Result
	}{name: "zero-denominator", result: Evaluate(zeroDenominator)})
	for _, testCase := range unknowns {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.result.Status != UNKNOWN || !testCase.result.FullSuiteRequired {
				t.Fatalf("result = %#v", testCase.result)
			}
		})
	}
}
