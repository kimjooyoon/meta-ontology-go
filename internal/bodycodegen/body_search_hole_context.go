package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

type IRBodyHoleProbe struct {
	Input    int64   `json:"input"`
	Expected int64   `json:"expected"`
	Zero     *int64  `json:"zero_output,omitempty"`
	One      *int64  `json:"one_output,omitempty"`
	Derived  *int64  `json:"derived_hole_value,omitempty"`
	Status   string  `json:"status"`
	MinusOne *int64  `json:"minus_one_output,omitempty"`
	Roots    []int64 `json:"quadratic_roots,omitempty"`
}

type IRBodyHoleContext struct {
	Schema              string            `json:"schema"`
	BodySHA256          string            `json:"body_sha256"`
	TrainingSuiteSHA256 string            `json:"training_suite_sha256"`
	EvaluationCalls     int               `json:"evaluation_calls"`
	Probes              []IRBodyHoleProbe `json:"probes"`
	Scope               string            `json:"scope"`
}

func generateIRBodySearchCandidatesForSource(ctx context.Context, filename string, source []byte, activity string,
	plan *IRBodySearchPlan) (*IRBodySearchCandidateGenerationReceipt, error) {
	if plan.CandidateGeneration == nil || (plan.CandidateGeneration.Grammar != bodySearchHoleResidualGrammar && !isQuadraticHoleGrammar(plan.CandidateGeneration.Grammar)) {
		return generateIRBodySearchCandidates(plan)
	}
	if ctx == nil {
		return nil, fmt.Errorf("hole candidate construction requires a context")
	}
	if err := validateIRBodySearchCandidateGeneration(*plan.CandidateGeneration); err != nil {
		return nil, err
	}
	if len(plan.TestCases) < 1 || len(plan.TestCases) > 128 {
		return nil, fmt.Errorf("hole residual grammar requires 1..128 training cases")
	}
	for _, test := range plan.TestCases {
		if len(test.Inputs) != 0 {
			return nil, fmt.Errorf("hole residual grammar uses the scalar input field")
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, _, id, body, err := prepareBodySearch(filename, source, activity, plan.HoleID)
	if err != nil {
		return nil, err
	}
	programs := holeProbePrograms(file.Package.Name, activity, id, body, plan.HoleID)
	training, _ := json.Marshal(plan.TestCases)
	observation := &IRBodyHoleContext{Schema: "gooo/integer-hole-context/v1", BodySHA256: digest([]byte(body)),
		TrainingSuiteSHA256: digest(training), Probes: make([]IRBodyHoleProbe, 0, len(plan.TestCases)),
		Scope: "pure typed-body outputs with the hole set to zero and one on training inputs; exact integer residual proposals; nonlinear and wrapped behavior still requires full candidate scoring"}
	var negative []byte
	quadratic := isQuadraticHoleGrammar(plan.CandidateGeneration.Grammar)
	if quadratic {
		negative = holeProbeProgram(file.Package.Name, activity, id, body, plan.HoleID, -1)
		observation.Schema = "gooo/integer-hole-context/v2"
		observation.Scope = "training-only minus-one/zero/one outputs; exact integer roots of a three-point quadratic fit are proposals; branches, higher-degree expressions and wrapped arithmetic require whole-body scoring"
	}
	for _, test := range plan.TestCases {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		probe := observeHoleProbe(ctx, activity, programs, test, &observation.EvaluationCalls)
		if quadratic {
			observeQuadraticHoleProbe(ctx, activity, negative, test, &observation.EvaluationCalls, &probe)
		}
		observation.Probes = append(observation.Probes, probe)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	expressions := holeContextExpressions(observation.Probes, bodySearchAffineExpressions(plan.TestCases))
	if quadratic {
		expressions = quadraticHoleExpressions(observation.Probes, expressions)
	}
	if plan.CandidateGeneration.Grammar == bodySearchHoleQuadraticFitGrammar {
		expressions = orderQuadraticHoleExpressions(observation.Probes, expressions)
	}
	return retainIRBodySearchCandidates(plan, expressions, observation)
}

func holeProbePrograms(packageName, activity, id, body, hole string) [2][]byte {
	var programs [2][]byte
	for value := range programs {
		programs[value] = holeProbeProgram(packageName, activity, id, body, hole, value)
	}
	return programs
}

func holeProbeProgram(packageName, activity, id, body, hole string, value int) []byte {
	filled, err := replaceIdentifier(body, bodyFillHoleToken(hole), strconv.Itoa(value))
	if err != nil {
		return nil
	}
	generated, err := generateRoute(packageName, activity, id, "int64", "int64", filled, preserveRoute)
	if err != nil {
		return nil
	}
	return generated.source
}

func observeHoleProbe(ctx context.Context, activity string, programs [2][]byte, test IRBodyFillTestCase, calls *int) IRBodyHoleProbe {
	probe := IRBodyHoleProbe{Input: test.Input, Expected: test.Expected, Status: "PROBE_UNAVAILABLE"}
	outputs := [2]**int64{&probe.Zero, &probe.One}
	for i, program := range programs {
		if len(program) == 0 || ctx.Err() != nil {
			continue
		}
		*calls += 1
		results, _, err := evaluateIntegerCasesContext(ctx, program, activity, []IRBodyFillTestCase{test})
		if err == nil && len(results) == 1 {
			value := results[0].Actual
			*outputs[i] = &value
		}
	}
	if probe.Zero != nil && probe.One != nil {
		probe.Derived, probe.Status = exactHoleResidual(*probe.Zero, *probe.One, test.Expected)
	}
	return probe
}
