package resourceenvelope

import (
	"testing"
)

type contractCorpus struct {
	Cases []contractCase `json:"cases"`
}
type contractCase struct {
	Name               string           `json:"name"`
	Samples            []contractSample `json:"samples"`
	WallSampleCount    uint64           `json:"wall_sample_count"`
	CPUSampleCount     uint64           `json:"cpu_sample_count"`
	WallBaselineNS     uint64           `json:"wall_baseline_ns"`
	CPUBaselineNS      uint64           `json:"cpu_baseline_ns"`
	WallMedianLimitPPM uint64           `json:"wall_median_limit_ppm"`
	WallMaxLimitPPM    uint64           `json:"wall_max_limit_ppm"`
	CPUMedianLimitPPM  uint64           `json:"cpu_median_limit_ppm"`
	CPUMaxLimitPPM     uint64           `json:"cpu_max_limit_ppm"`
	Expected           contractExpected `json:"expected"`
}
type contractSample struct {
	WallNS uint64 `json:"wall_ns"`
	CPUNS  uint64 `json:"cpu_ns"`
}
type contractExpected struct {
	Decision      string `json:"decision"`
	Reason        string `json:"reason"`
	WallMedianNS  uint64 `json:"wall_median_ns"`
	WallMaxNS     uint64 `json:"wall_max_ns"`
	CPUMedianNS   uint64 `json:"cpu_median_ns"`
	CPUMaxNS      uint64 `json:"cpu_max_ns"`
	WallMedianPPM uint64 `json:"wall_median_ppm"`
	WallMaxPPM    uint64 `json:"wall_max_ppm"`
	CPUMedianPPM  uint64 `json:"cpu_median_ppm"`
	CPUMaxPPM     uint64 `json:"cpu_max_ppm"`
}
type contractObservation struct {
	Decision      string `json:"decision"`
	Reason        string `json:"reason"`
	WallMedianNS  uint64 `json:"wall_median_ns,omitempty"`
	WallMaxNS     uint64 `json:"wall_max_ns,omitempty"`
	CPUMedianNS   uint64 `json:"cpu_median_ns,omitempty"`
	CPUMaxNS      uint64 `json:"cpu_max_ns,omitempty"`
	WallMedianPPM uint64 `json:"wall_median_ppm,omitempty"`
	WallMaxPPM    uint64 `json:"wall_max_ppm,omitempty"`
	CPUMedianPPM  uint64 `json:"cpu_median_ppm,omitempty"`
	CPUMaxPPM     uint64 `json:"cpu_max_ppm,omitempty"`
}

func TestResourceEnvelopeContractCasesUseIndependentMedianAndPPM(t *testing.T) {
	corpus := loadContractCorpus(t)
	for _, testCase := range corpus.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := independentObservation(testCase)
			want := testCase.Expected
			if got.Decision != want.Decision || got.Reason != want.Reason {
				t.Fatalf("decision = %s/%s, want %s/%s", got.Decision, got.Reason, want.Decision, want.Reason)
			}
			if got.WallMedianNS != want.WallMedianNS || got.WallMaxNS != want.WallMaxNS ||
				got.CPUMedianNS != want.CPUMedianNS || got.CPUMaxNS != want.CPUMaxNS {
				t.Fatalf("order statistics = %#v, want wall=%d/%d cpu=%d/%d", got, want.WallMedianNS, want.WallMaxNS, want.CPUMedianNS, want.CPUMaxNS)
			}
			if got.WallMedianPPM != want.WallMedianPPM || got.WallMaxPPM != want.WallMaxPPM ||
				got.CPUMedianPPM != want.CPUMedianPPM || got.CPUMaxPPM != want.CPUMaxPPM {
				t.Fatalf("ppm = %#v, want wall=%d/%d cpu=%d/%d", got, want.WallMedianPPM, want.WallMaxPPM, want.CPUMedianPPM, want.CPUMaxPPM)
			}
		})
	}
}
