package bodycodegen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"slices"
)

const irBodyFillBehavioralProbeSchema = "gooo/ir-body-fill-behavioral-probe/v1"
const irBodyFillBehavioralProbeCap = 128

// measureIRBodyFillBehavioralProbes compares candidate behavior on a bounded,
// deterministic set derived only from training inputs. It runs after candidate
// selection and never treats probe outputs as expected answers.
func measureIRBodyFillBehavioralProbes(activity string, candidateSources map[string][]byte,
	training []IRBodyFillTestCase,
) (*IRBodyFillBehavioralProbeReceipt, error) {
	allInputs := deriveIRBodyFillBehavioralProbeInputs(training)
	inputs := retainIRBodyFillBehavioralProbes(allInputs, irBodyFillBehavioralProbeCap)
	encodedInputs, err := json.Marshal(inputs)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(encodedInputs)
	receipt := &IRBodyFillBehavioralProbeReceipt{
		Schema:             irBodyFillBehavioralProbeSchema,
		ProbeInputs:        inputs,
		ProbeInputsSHA256:  "sha256:" + hex.EncodeToString(digest[:]),
		ProbeInputsTotal:   len(allInputs),
		ProbeInputsOmitted: len(allInputs) - len(inputs),
		CandidateCount:     len(candidateSources),
		Scope:              "synthetic integer inputs derived only from training inputs and evaluated after selection; outputs are compared across declared candidates without an expected-value oracle, and do not measure intent or correctness",
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
	return receipt, nil
}

func deriveIRBodyFillBehavioralProbeInputs(training []IRBodyFillTestCase) []int64 {
	inputs := make([]int64, len(training))
	trainingInputs := make(map[int64]struct{}, len(training))
	for index, testCase := range training {
		inputs[index] = testCase.Input
		trainingInputs[testCase.Input] = struct{}{}
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
