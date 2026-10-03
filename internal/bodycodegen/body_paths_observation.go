package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"go/token"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// OracleActivity names a separate, caller-declared pure activity in the same
// authoritative source. Empty means suggest an input without inventing a label.
type PathObservationOptions struct {
	Inputs         []int64 `json:"inputs"`
	MaxCandidates  int     `json:"max_candidates"`
	MaxRounds      int     `json:"max_rounds"`
	OracleActivity string  `json:"oracle_activity,omitempty"`
}

type PathObservationRound struct {
	Ranking     pathplan.ProbeRanking `json:"ranking"`
	Observation *pathplan.TestCase    `json:"oracle_observation,omitempty"`
}

type PathObservationReceipt struct {
	Schema             string                 `json:"schema"`
	Status             string                 `json:"status"`
	Options            PathObservationOptions `json:"options"`
	OptionsSHA256      string                 `json:"options_sha256"`
	SourceSHA256       string                 `json:"source_sha256"`
	PreparedPlanSHA256 string                 `json:"prepared_plan_sha256"`
	InitialCases       []pathplan.TestCase    `json:"initial_cases"`
	EffectiveCasesSHA  string                 `json:"effective_cases_sha256"`
	EffectiveCases     int                    `json:"effective_cases"`
	OracleActivityID   string                 `json:"oracle_activity_id,omitempty"`
	OracleGoSHA256     string                 `json:"oracle_go_sha256,omitempty"`
	OracleInitialCases []IRBodyFillCaseResult `json:"oracle_initial_cases,omitempty"`
	OracleEvaluations  int                    `json:"oracle_evaluations"`
	Rounds             []PathObservationRound `json:"rounds"`
	Scope              string                 `json:"scope"`
}

func copyPathObservation(options *PathObservationOptions) (*PathObservationOptions, error) {
	if options == nil {
		return nil, nil
	}
	if len(options.Inputs) == 0 || len(options.Inputs) > 32 || options.MaxCandidates < 1 ||
		options.MaxCandidates > 64 || options.MaxRounds < 1 || options.MaxRounds > 8 {
		return nil, fmt.Errorf("path observation requires 1..32 inputs, 1..64 candidates and 1..8 rounds")
	}
	if options.OracleActivity != "" && (len(options.OracleActivity) > 128 || !token.IsIdentifier(options.OracleActivity)) {
		return nil, fmt.Errorf("path observation oracle must be an activity identifier")
	}
	owned := *options
	owned.Inputs = append([]int64(nil), options.Inputs...)
	return &owned, nil
}

func pathCasesDigest(cases []pathplan.TestCase) string {
	raw, _ := json.Marshal(cases)
	return digest(raw)
}

// The finite suite is extended before model ranking. Original source, document
// and initial suite remain bound separately. Neither candidates nor the model
// supply their own expected answers, and no subprocess/network is invoked.
func observeTypedPaths(ctx context.Context, filename string, source []byte, activity string,
	prepared *pathplan.PreparedPlan, original []pathplan.TestCase, options *PathObservationOptions,
	receipt *BodyPathReceipt) (cases []pathplan.TestCase, failure error) {
	if options == nil {
		return original, nil
	}
	started := time.Now()
	cases = append([]pathplan.TestCase(nil), original...)
	raw, _ := json.Marshal(options)
	r := &PathObservationReceipt{Schema: "gooo/typed-path-observation/v1", Status: "STARTED",
		Options: *options, OptionsSHA256: digest(raw), SourceSHA256: digest(source),
		PreparedPlanSHA256: prepared.PlanSHA256(), InitialCases: append([]pathplan.TestCase(nil), original...),
		Scope: "bounded typed candidates and caller-declared pure Gooo oracle; Go AST evaluation; no native process or all-input proof"}
	receipt.Observation = r
	defer func() {
		r.EffectiveCasesSHA, r.EffectiveCases = pathCasesDigest(cases), len(cases)
		receipt.Timing.ObservationMS = elapsedMS(started)
		if failure != nil {
			r.Status = "FAILED"
		}
	}()
	oracle, err := preparePathOracle(ctx, filename, source, activity, original, r)
	if err != nil {
		return cases, err
	}
	for round := 0; ; round++ {
		ranking, err := prepared.RankProbes(ctx, cases, options.Inputs, options.MaxCandidates)
		if ranking.Schema != "" {
			r.Rounds = append(r.Rounds, PathObservationRound{Ranking: ranking})
		}
		if err != nil {
			return cases, err
		}
		if ranking.RecommendedIndex == nil {
			r.Status = "NO_DISTINGUISHING_INPUT"
			if len(ranking.SurvivingMasks) == 0 {
				r.Status = "NO_SURVIVING_CANDIDATE"
			}
			if len(ranking.SurvivingMasks) == 1 {
				r.Status = "ONE_SURVIVING_CANDIDATE"
			}
			if ranking.Unobserved > 0 {
				r.Status = "CANDIDATE_BUDGET"
			}
			return cases, nil
		}
		if options.OracleActivity == "" {
			r.Status = "ORACLE_UNAVAILABLE"
			return cases, nil
		}
		if round == options.MaxRounds {
			r.Status = "ROUND_BUDGET"
			return cases, nil
		}
		if len(cases) == 128 {
			r.Status = "CASE_BUDGET"
			return cases, nil
		}
		input := ranking.Probes[*ranking.RecommendedIndex].Input
		values, _, err := evaluateIntegerCasesContext(ctx, []byte(oracle.Source), options.OracleActivity,
			[]IRBodyFillTestCase{{Input: input}})
		if err != nil {
			return cases, fmt.Errorf("observe declared Gooo oracle: %w", err)
		}
		r.OracleEvaluations++
		observation := pathplan.TestCase{Input: input, Expected: values[0].Actual}
		r.Rounds[len(r.Rounds)-1].Observation = &observation
		cases = append(cases, observation)
	}
}

func preparePathOracle(ctx context.Context, filename string, source []byte, activity string,
	cases []pathplan.TestCase, receipt *PathObservationReceipt) (Result, error) {
	name := receipt.Options.OracleActivity
	if name == "" {
		return Result{}, nil
	}
	if name == activity {
		return Result{}, fmt.Errorf("path observation oracle must name a separate declared activity")
	}
	oracle, err := GenerateWithPlanner(ctx, filename, source, name, "", "")
	if err != nil {
		return Result{}, fmt.Errorf("prepare declared Gooo oracle: %w", err)
	}
	if oracle.Report.InputType != "int64" || oracle.Report.OutputType != "int64" {
		return Result{}, fmt.Errorf("path observation oracle must be Integer -> Integer")
	}
	receipt.OracleActivityID, receipt.OracleGoSHA256 = oracle.Report.ActivityID, digest([]byte(oracle.Source))
	initial := make([]IRBodyFillTestCase, len(cases))
	for i, c := range cases {
		initial[i] = IRBodyFillTestCase{Input: c.Input, Expected: c.Expected}
	}
	results, passed, err := evaluateIntegerCasesContext(ctx, []byte(oracle.Source), name, initial)
	if err != nil {
		return Result{}, err
	}
	receipt.OracleInitialCases, receipt.OracleEvaluations = results, len(results)
	if passed != len(cases) {
		return Result{}, fmt.Errorf("declared Gooo oracle contradicts the original finite cases")
	}
	return oracle, nil
}
