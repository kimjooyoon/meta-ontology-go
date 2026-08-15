package resourceenvelope

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
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

func TestResourceEnvelopeContractInputPermutationInvariance(t *testing.T) {
	corpus := loadContractCorpus(t)
	base := contractByName(t, corpus, "five-sample-boundary")
	want := independentObservation(base)
	for left, right := 0, len(base.Samples)-1; left < right; left, right = left+1, right-1 {
		base.Samples[left], base.Samples[right] = base.Samples[right], base.Samples[left]
	}
	if got := independentObservation(base); got != want {
		t.Fatalf("permuted samples changed observation: got=%#v want=%#v", got, want)
	}
}

func TestResourceEnvelopeContractCanonicalReplayEquality(t *testing.T) {
	corpus := loadContractCorpus(t)
	base := contractByName(t, corpus, "five-sample-boundary")
	first, err := json.Marshal(independentObservation(base))
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(independentObservation(base))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("canonical replay changed: first=%s second=%s", first, second)
	}
}

func TestResourceEnvelopeContractRejectsUnknownJSONField(t *testing.T) {
	raw := readCorpus(t)
	var document map[string]json.RawMessage
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	document["unregistered"] = json.RawMessage(`true`)
	mutated, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := decodeStrict(mutated, &contractCorpus{}); err == nil {
		t.Fatal("unknown JSON field was accepted")
	}
}

func TestResourceEnvelopeContractRejectsTrailingJSON(t *testing.T) {
	raw := append(readCorpus(t), []byte(`
{"trailing":true}`)...)
	if err := decodeStrict(raw, &contractCorpus{}); err == nil {
		t.Fatal("trailing JSON was accepted")
	}
}

func TestResourceEnvelopeImplementationDependencyLocalNotRun(t *testing.T) {
	t.Skip("dependency-local NOT_RUN: resourceenvelope production implementation and entry point are absent from authority leaf")
}

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

func readCorpus(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "resource-envelope-cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func decodeStrict(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return io.ErrUnexpectedEOF
		}
		return err
	}
	return nil
}

func contractByName(t *testing.T, corpus contractCorpus, name string) contractCase {
	t.Helper()
	for _, testCase := range corpus.Cases {
		if testCase.Name == name {
			return testCase
		}
	}
	t.Fatalf("missing contract case %q", name)
	return contractCase{}
}
