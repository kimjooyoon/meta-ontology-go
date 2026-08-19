package resourceenvelope

import (
	"math"
	"sort"
	"testing"
)

func independentObservation(testCase contractCase) contractObservation {
	if len(testCase.Samples) != 5 {
		return contractObservation{Decision: "UNKNOWN", Reason: "sample_count"}
	}
	if testCase.WallSampleCount == 0 || testCase.CPUSampleCount == 0 {
		return contractObservation{Decision: "UNKNOWN", Reason: "zero_sample_count"}
	}
	wall := make([]uint64, 0, len(testCase.Samples))
	cpu := make([]uint64, 0, len(testCase.Samples))
	for _, sample := range testCase.Samples {
		wall = append(wall, sample.WallNS)
		cpu = append(cpu, sample.CPUNS)
	}
	wallMedian, wallMax := medianAndMax(wall)
	cpuMedian, cpuMax := medianAndMax(cpu)
	wallMedianPPM, ok := independentPPM(wallMedian, testCase.WallBaselineNS)
	if !ok {
		return contractObservation{Decision: "UNKNOWN", Reason: "ppm_overflow"}
	}
	wallMaxPPM, ok := independentPPM(wallMax, testCase.WallBaselineNS)
	if !ok {
		return contractObservation{Decision: "UNKNOWN", Reason: "ppm_overflow"}
	}
	cpuMedianPPM, ok := independentPPM(cpuMedian, testCase.CPUBaselineNS)
	if !ok {
		return contractObservation{Decision: "UNKNOWN", Reason: "ppm_overflow"}
	}
	cpuMaxPPM, ok := independentPPM(cpuMax, testCase.CPUBaselineNS)
	if !ok {
		return contractObservation{Decision: "UNKNOWN", Reason: "ppm_overflow"}
	}
	decision := "PASS"
	if wallMedianPPM > testCase.WallMedianLimitPPM || wallMaxPPM > testCase.WallMaxLimitPPM ||
		cpuMedianPPM > testCase.CPUMedianLimitPPM || cpuMaxPPM > testCase.CPUMaxLimitPPM {
		decision = "FAIL_CLOSED"
	}
	return contractObservation{Decision: decision, Reason: "", WallMedianNS: wallMedian, WallMaxNS: wallMax,
		CPUMedianNS: cpuMedian, CPUMaxNS: cpuMax, WallMedianPPM: wallMedianPPM, WallMaxPPM: wallMaxPPM,
		CPUMedianPPM: cpuMedianPPM, CPUMaxPPM: cpuMaxPPM}
}
func medianAndMax(values []uint64) (uint64, uint64) {
	ordered := append([]uint64(nil), values...)
	sort.Slice(ordered, func(left, right int) bool { return ordered[left] < ordered[right] })
	return ordered[len(ordered)/2], ordered[len(ordered)-1]
}
func independentPPM(value, baseline uint64) (uint64, bool) {
	if baseline == 0 || value > math.MaxUint64/1_000_000 {
		return 0, false
	}
	return value * 1_000_000 / baseline, true
}
func loadContractCorpus(t *testing.T) contractCorpus {
	t.Helper()
	var corpus contractCorpus
	if err := decodeStrict(readCorpus(t), &corpus); err != nil {
		t.Fatal(err)
	}
	return corpus
}
