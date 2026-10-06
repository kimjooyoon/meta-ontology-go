package bodycodegen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"math/big"
	"slices"
	"sort"
)

const irBodyFillBehavioralProbeSchema = "gooo/ir-body-fill-behavioral-probe/v1"
const irBodyFillBehavioralProbeCap = 128

// measureIRBodyFillBehavioralProbes compares candidate behavior on a bounded,
// deterministic set derived only from training inputs. It never treats probe
// outputs as expected answers and does not consume holdout data.
func measureIRBodyFillBehavioralProbes(activity string, candidateSources map[string][]byte,
	training, holdout []IRBodyFillTestCase,
) (*IRBodyFillBehavioralProbeReceipt, error) {
	if len(training) > 0 && len(training[0].inputValues()) > 1 {
		return measureIRBodyFillBehavioralProbeVectors(activity, candidateSources, training, holdout)
	}
	holdoutInputs := make([]int64, len(holdout))
	for index, testCase := range holdout {
		holdoutInputs[index] = testCase.Input
	}
	allInputs := deriveIRBodyFillBehavioralProbeInputs(training, holdoutInputs...)
	inputs := retainIRBodyFillBehavioralProbes(allInputs, irBodyFillBehavioralProbeCap)
	receipt := &IRBodyFillBehavioralProbeReceipt{
		Schema:             irBodyFillBehavioralProbeSchema,
		ProbeInputs:        inputs,
		ProbeInputsTotal:   len(allInputs),
		ProbeInputsOmitted: len(allInputs) - len(inputs),
		CandidateCount:     len(candidateSources),
		Scope:              "synthetic integer inputs derived only from training inputs with declared holdout inputs excluded; candidate outputs have no expected-value oracle and measure behavioral distinctions, not intent or correctness",
	}
	if len(inputs) == 0 {
		return receipt, nil
	}
	probeCases := make([]IRBodyFillTestCase, len(inputs))
	for index, input := range inputs {
		probeCases[index] = IRBodyFillTestCase{Input: input}
	}
	candidateIDs := make([]string, 0, len(candidateSources))
	for id := range candidateSources {
		candidateIDs = append(candidateIDs, id)
	}
	slices.Sort(candidateIDs)
	outputs := make(map[string][]int64, len(candidateIDs))
	for _, id := range candidateIDs {
		results, _, evalErr := evaluateIntegerCases(candidateSources[id], activity, probeCases)
		if evalErr != nil || len(results) != len(inputs) {
			receipt.CandidateRunsFailed++
			continue
		}
		values := make([]int64, len(results))
		for index, result := range results {
			values[index] = result.Actual
		}
		outputs[id] = values
		receipt.CandidateProfiles = append(receipt.CandidateProfiles, IRBodyFillCandidateProbeProfile{
			CandidateID: id, Outputs: slices.Clone(values),
		})
		receipt.CandidateRunsCompleted++
	}
	for probeIndex := range inputs {
		values := make(map[int64]struct{}, len(outputs))
		for _, candidateOutputs := range outputs {
			values[candidateOutputs[probeIndex]] = struct{}{}
		}
		if len(values) > 1 {
			receipt.ProbeInputsWithDisagreement++
		}
	}
	receipt.CandidatePairsTotal = len(candidateIDs) * (len(candidateIDs) - 1) / 2
	for left := 0; left < len(candidateIDs); left++ {
		leftOutputs, leftOK := outputs[candidateIDs[left]]
		if !leftOK {
			continue
		}
		for right := left + 1; right < len(candidateIDs); right++ {
			rightOutputs, rightOK := outputs[candidateIDs[right]]
			if !rightOK {
				continue
			}
			receipt.CandidatePairsEvaluated++
			for index := range inputs {
				if leftOutputs[index] != rightOutputs[index] {
					receipt.CandidatePairsDistinguished++
					break
				}
			}
		}
	}
	if receipt.CandidatePairsEvaluated > 0 {
		receipt.CandidatePairDistinguishabilityPercent = float64(receipt.CandidatePairsDistinguished) * 100 /
			float64(receipt.CandidatePairsEvaluated)
	}
	encodedProfile, err := json.Marshal(struct {
		Inputs     []int64                           `json:"inputs"`
		Candidates []IRBodyFillCandidateProbeProfile `json:"candidates"`
	}{Inputs: inputs, Candidates: receipt.CandidateProfiles})
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(encodedProfile)
	receipt.ProbeProfileSHA256 = "sha256:" + hex.EncodeToString(digest[:])
	return receipt, nil
}

func measureIRBodyFillBehavioralProbeVectors(activity string, candidateSources map[string][]byte,
	training, holdout []IRBodyFillTestCase,
) (*IRBodyFillBehavioralProbeReceipt, error) {
	allVectors := deriveIRBodyFillBehavioralProbeVectors(training, holdout)
	vectors := retainIRBodyFillBehavioralProbeVectors(allVectors, irBodyFillBehavioralProbeCap)
	receipt := &IRBodyFillBehavioralProbeReceipt{
		Schema:             "gooo/ir-body-fill-behavioral-probe/v2",
		ProbeVectors:       vectors,
		ProbeInputsTotal:   len(allVectors),
		ProbeInputsOmitted: len(allVectors) - len(vectors),
		CandidateCount:     len(candidateSources),
		Scope:              "synthetic integer input vectors perturb one declared input at a time from training cases; declared holdout vectors are excluded; outputs have no expected-value oracle and measure distinctions, not intent or correctness",
	}
	if len(vectors) == 0 {
		return receipt, nil
	}
	probeCases := make([]IRBodyFillTestCase, len(vectors))
	for index, vector := range vectors {
		probeCases[index] = IRBodyFillTestCase{Input: vector[0], Inputs: slices.Clone(vector)}
	}
	candidateIDs := make([]string, 0, len(candidateSources))
	for id := range candidateSources {
		candidateIDs = append(candidateIDs, id)
	}
	slices.Sort(candidateIDs)
	outputs := make(map[string][]int64, len(candidateIDs))
	for _, id := range candidateIDs {
		results, _, evalErr := evaluateIntegerCases(candidateSources[id], activity, probeCases)
		if evalErr != nil || len(results) != len(vectors) {
			receipt.CandidateRunsFailed++
			continue
		}
		values := make([]int64, len(results))
		for index, result := range results {
			values[index] = result.Actual
		}
		outputs[id] = values
		receipt.CandidateProfiles = append(receipt.CandidateProfiles, IRBodyFillCandidateProbeProfile{
			CandidateID: id, Outputs: slices.Clone(values),
		})
		receipt.CandidateRunsCompleted++
	}
	for probeIndex := range vectors {
		values := make(map[int64]struct{}, len(outputs))
		for _, candidateOutputs := range outputs {
			values[candidateOutputs[probeIndex]] = struct{}{}
		}
		if len(values) > 1 {
			receipt.ProbeInputsWithDisagreement++
		}
	}
	receipt.CandidatePairsTotal = len(candidateIDs) * (len(candidateIDs) - 1) / 2
	for left := 0; left < len(candidateIDs); left++ {
		leftOutputs, leftOK := outputs[candidateIDs[left]]
		if !leftOK {
			continue
		}
		for right := left + 1; right < len(candidateIDs); right++ {
			rightOutputs, rightOK := outputs[candidateIDs[right]]
			if !rightOK {
				continue
			}
			receipt.CandidatePairsEvaluated++
			for index := range vectors {
				if leftOutputs[index] != rightOutputs[index] {
					receipt.CandidatePairsDistinguished++
					break
				}
			}
		}
	}
	if receipt.CandidatePairsEvaluated > 0 {
		receipt.CandidatePairDistinguishabilityPercent = float64(receipt.CandidatePairsDistinguished) * 100 /
			float64(receipt.CandidatePairsEvaluated)
	}
	encodedProfile, err := json.Marshal(struct {
		Vectors    [][]int64                         `json:"vectors"`
		Candidates []IRBodyFillCandidateProbeProfile `json:"candidates"`
	}{Vectors: vectors, Candidates: receipt.CandidateProfiles})
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(encodedProfile)
	receipt.ProbeProfileSHA256 = "sha256:" + hex.EncodeToString(digest[:])
	return receipt, nil
}

func deriveIRBodyFillBehavioralProbeVectors(training, holdout []IRBodyFillTestCase) [][]int64 {
	excluded := make(map[string]struct{}, len(training)+len(holdout))
	for _, testCase := range append(append([]IRBodyFillTestCase(nil), training...), holdout...) {
		excluded[bodyFillInputKey(testCase.inputValues())] = struct{}{}
	}
	probes := make(map[string][]int64)
	for _, testCase := range training {
		base := testCase.inputValues()
		for index, value := range base {
			for _, delta := range []int64{-1, 1} {
				if delta < 0 && value == math.MinInt64 || delta > 0 && value == math.MaxInt64 {
					continue
				}
				candidate := slices.Clone(base)
				candidate[index] += delta
				key := bodyFillInputKey(candidate)
				if _, exists := excluded[key]; !exists {
					probes[key] = candidate
				}
			}
		}
	}
	result := make([][]int64, 0, len(probes))
	for _, vector := range probes {
		result = append(result, vector)
	}
	sort.Slice(result, func(i, j int) bool { return slices.Compare(result[i], result[j]) < 0 })
	return result
}

func retainIRBodyFillBehavioralProbeVectors(vectors [][]int64, limit int) [][]int64 {
	if len(vectors) <= limit {
		return slices.Clone(vectors)
	}
	if limit <= 1 {
		return [][]int64{slices.Clone(vectors[0])}
	}
	retained := make([][]int64, limit)
	for index := range retained {
		position := index * (len(vectors) - 1) / (limit - 1)
		retained[index] = slices.Clone(vectors[position])
	}
	return retained
}

func deriveIRBodyFillBehavioralProbeInputs(training []IRBodyFillTestCase, excludedInputs ...int64) []int64 {
	inputs := make([]int64, len(training))
	trainingInputs := make(map[int64]struct{}, len(training))
	for index, testCase := range training {
		inputs[index] = testCase.Input
		trainingInputs[testCase.Input] = struct{}{}
	}
	for _, input := range excludedInputs {
		trainingInputs[input] = struct{}{}
	}
	slices.Sort(inputs)
	inputs = slices.Compact(inputs)
	probes := make(map[int64]struct{}, len(inputs)*4)
	add := func(value int64) {
		if _, isTrainingInput := trainingInputs[value]; !isTrainingInput {
			probes[value] = struct{}{}
		}
	}
	addNeighbors := func(value int64) {
		if value > -1<<63 {
			add(value - 1)
		}
		if value < 1<<63-1 {
			add(value + 1)
		}
	}
	for index := 0; index+1 < len(inputs); index++ {
		lower := big.NewInt(inputs[index])
		upper := big.NewInt(inputs[index+1])
		gap := new(big.Int).Sub(upper, lower)
		if gap.Cmp(big.NewInt(1)) <= 0 {
			continue
		}
		half, remainder := new(big.Int), new(big.Int)
		half.QuoRem(gap, big.NewInt(2), remainder)
		floor := new(big.Int).Add(lower, half).Int64()
		add(floor)
		addNeighbors(floor)
		if remainder.Sign() != 0 {
			ceiling := floor + 1
			add(ceiling)
			addNeighbors(ceiling)
		}
	}
	for _, input := range inputs {
		addNeighbors(input)
	}
	result := make([]int64, 0, len(probes))
	for probe := range probes {
		result = append(result, probe)
	}
	slices.Sort(result)
	return result
}

func retainIRBodyFillBehavioralProbes(inputs []int64, limit int) []int64 {
	if len(inputs) <= limit {
		return slices.Clone(inputs)
	}
	if limit <= 1 {
		return []int64{inputs[0]}
	}
	retained := make([]int64, limit)
	for index := range retained {
		position := index * (len(inputs) - 1) / (limit - 1)
		retained[index] = inputs[position]
	}
	return retained
}
