package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func replayPathObservation(ctx context.Context, filename string, source []byte, activity string,
	prepared *pathplan.PreparedPlan, original []pathplan.TestCase, prior *BodyPathReceipt) ([]pathplan.TestCase, error) {
	if prior.Observation == nil {
		return original, nil
	}
	if _, _, bad := observedPathContract(prior); bad {
		return nil, fmt.Errorf("typed-path observation record is unbound")
	}
	options, err := copyPathObservation(&prior.Observation.Options)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, bodyPathBudget)
	defer cancel()
	observed := &BodyPathReceipt{}
	cases, err := observeTypedPaths(ctx, filename, source, activity, prepared, original, options, observed)
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(observed.Observation, prior.Observation) {
		return nil, fmt.Errorf("typed-path candidate observations or oracle labels do not replay")
	}
	return cases, nil
}

// Reconstruct the append-only suite from original cases and source-bound oracle
// observations. Original document/test identities are never overwritten.
func observedPathContract(p *BodyPathReceipt) (string, int, bool) {
	r := p.Observation
	if r == nil {
		return p.TestSuiteSHA256, p.DeclaredTestCases, false
	}
	options, err := copyPathObservation(&r.Options)
	if err != nil {
		return "", 0, true
	}
	raw, _ := json.Marshal(options)
	cases := append([]pathplan.TestCase(nil), r.InitialCases...)
	bad := r.Schema != "gooo/typed-path-observation/v1" || digest(raw) != r.OptionsSHA256 ||
		r.SourceSHA256 != p.OriginalSourceSHA256 || !validDigest(r.SourceSHA256) ||
		!validDigest("sha256:"+r.PreparedPlanSHA256) || len(cases) != p.DeclaredTestCases ||
		len(cases) == 0 || len(cases) > 128 || pathCasesDigest(cases) != p.TestSuiteSHA256 ||
		len(r.Rounds) > options.MaxRounds+1 || !pathProbeReuseBound(r)
	probeRaw, _ := json.Marshal(options.Inputs)
	observed := 0
	for i, round := range r.Rounds {
		q := round.Ranking
		bad = bad || "sha256:"+q.CasesSHA256 != pathCasesDigest(cases) ||
			q.PlanSHA256 != r.PreparedPlanSHA256 || "sha256:"+q.ProbesSHA256 != digest(probeRaw) ||
			q.ModelPredictions != 0 || q.Observed < 0 || q.Observed > options.MaxCandidates ||
			q.Declared != q.Observed+q.Unobserved || q.Unobserved < 0
		if round.Observation == nil {
			bad = bad || i != len(r.Rounds)-1
			continue
		}
		if q.RecommendedIndex == nil || *q.RecommendedIndex < 0 || *q.RecommendedIndex >= len(q.Probes) {
			return "", 0, true
		}
		probe := q.Probes[*q.RecommendedIndex]
		bad = bad || options.OracleActivity == "" || probe.Input != round.Observation.Input ||
			probe.AlreadyTested || probe.SeparatedPairs <= 0 || observed >= options.MaxRounds
		for _, c := range cases {
			bad = bad || c.Input == round.Observation.Input
		}
		cases = append(cases, *round.Observation)
		observed++
	}
	if options.OracleActivity == "" {
		bad = bad || r.OracleEvaluations != 0 || len(r.OracleInitialCases) != 0 ||
			r.OracleActivityID != "" || r.OracleGoSHA256 != ""
	} else if r.Status != "FAILED" {
		bad = bad || !validDigest(r.OracleGoSHA256) || r.OracleActivityID == "" ||
			len(r.OracleInitialCases) != len(r.InitialCases) || r.OracleEvaluations != len(r.InitialCases)+observed
		for i, c := range r.OracleInitialCases {
			if i >= len(r.InitialCases) {
				return "", 0, true
			}
			initial := r.InitialCases[i]
			bad = bad || c.Input != initial.Input || c.Expected != initial.Expected ||
				c.Actual != c.Expected || !c.Passed
		}
	}
	sha := pathCasesDigest(cases)
	bad = bad || len(cases) > 128 || sha != r.EffectiveCasesSHA || len(cases) != r.EffectiveCases
	return sha, len(cases), bad
}

func pathObservationDimension(p *BodyPathReceipt) CompletenessDimension {
	_, _, bad := observedPathContract(p)
	r := p.Observation
	return completenessDimension("typed_path_observation_binding", boolCount(!bad && r.Status != "FAILED"), 1,
		"source-bound original cases, probe recommendations and append-only oracle observations",
		"Binding describes the observation record; unresolved probes and finite ambiguity retain their status. Oracle evaluation uses the bounded Go AST interpreter.",
		[]string{"body_paths.observation", "observation_status:" + r.Status,
			"oracle_source_sha256:" + r.SourceSHA256, "effective_cases_sha256:" + r.EffectiveCasesSHA}, bad)
}
