package resourceenvelope

import "sort"

const partsPerMillion uint64 = 1_000_000

// Evaluate evaluates only the observations in envelope. Invalid shape,
// arithmetic overflow, and an invalid CPU denominator produce UNKNOWN.
// Inclusive bound overruns produce FAIL_CLOSED.
func Evaluate(envelope Envelope) Result {
	result := Result{SchemaVersion: SchemaVersion, Status: UNKNOWN}
	if err := envelope.Validate(); err != nil {
		return sealResult(result, "invalid-envelope")
	}
	retained := append([]Sample(nil), envelope.Samples[envelope.WarmupCount:]...)
	sort.SliceStable(retained, func(left, right int) bool {
		return retained[left].CPUCoreNS < retained[right].CPUCoreNS
	})
	selected := retained[2]
	utilization, ok := cpuUtilizationPPM(selected.CPUCoreNS, selected.WallNS, envelope.AllocatedCPUCount)
	if !ok {
		return sealResult(result, "cpu-arithmetic")
	}
	result.CPUCoreNS = selected.CPUCoreNS
	result.CPUUtilizationPPM = utilization
	result.PeakRSSBytes = maxPeakRSS(retained)
	result.ReadBytes = maxReadBytes(retained)
	result.WriteBytes = maxWriteBytes(retained)
	if result.CPUCoreNS > envelope.Limits.CPUCoreNS ||
		result.PeakRSSBytes > envelope.Limits.PeakRSSBytes ||
		result.ReadBytes > envelope.Limits.ReadBytes ||
		result.WriteBytes > envelope.Limits.WriteBytes {
		return sealResult(result, "resource-overrun")
	}
	result.Status = PASS
	return sealResult(result, "")
}

// EvaluateJSON maps every decode or shape failure to a sealed UNKNOWN result.
func EvaluateJSON(data []byte) Result {
	envelope, err := DecodeJSON(data)
	if err != nil {
		return sealResult(Result{SchemaVersion: SchemaVersion, Status: UNKNOWN}, "invalid-json")
	}
	return Evaluate(envelope)
}

// EvaluateJSONWithError preserves the decode error while still returning the
// required UNKNOWN result for malformed or mismatched input.
func EvaluateJSONWithError(data []byte) (Result, error) {
	envelope, err := DecodeJSON(data)
	if err != nil {
		return sealResult(Result{SchemaVersion: SchemaVersion, Status: UNKNOWN}, "invalid-json"), err
	}
	return Evaluate(envelope), nil
}

func cpuUtilizationPPM(cpuCoreNS, wallNS, allocatedCPUCount uint64) (uint64, bool) {
	if wallNS == 0 || allocatedCPUCount == 0 || cpuCoreNS > ^uint64(0)/partsPerMillion {
		return 0, false
	}
	if wallNS > ^uint64(0)/allocatedCPUCount {
		return 0, false
	}
	return (cpuCoreNS * partsPerMillion) / (wallNS * allocatedCPUCount), true
}

func maxPeakRSS(samples []Sample) uint64 {
	var value uint64
	for _, sample := range samples {
		if sample.PeakRSSBytes > value {
			value = sample.PeakRSSBytes
		}
	}
	return value
}

func maxReadBytes(samples []Sample) uint64 {
	var value uint64
	for _, sample := range samples {
		if sample.ReadBytes > value {
			value = sample.ReadBytes
		}
	}
	return value
}

func maxWriteBytes(samples []Sample) uint64 {
	var value uint64
	for _, sample := range samples {
		if sample.WriteBytes > value {
			value = sample.WriteBytes
		}
	}
	return value
}
