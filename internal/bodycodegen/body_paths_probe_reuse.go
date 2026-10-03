package bodycodegen

import (
	"context"
	"encoding/json"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// PathProbeReuse records work reused within this request. Each round stores its
// ranking once; this record adds only the bounded session's accounting fields.
type PathProbeReuse struct {
	Revision                int    `json:"revision"`
	InitialRankingSHA256    string `json:"initial_ranking_sha256"`
	ReusedProbeValues       int    `json:"reused_probe_values"`
	CachedComparisons       int    `json:"cached_comparisons"`
	TotalEvaluationAttempts int    `json:"total_evaluation_attempts"`
	TotalCachedComparisons  int    `json:"total_cached_comparisons"`
}

type pathObservationRanker struct {
	prepared *pathplan.PreparedPlan
	options  *PathObservationOptions
	session  *pathplan.ProbeSession
}

func (r *pathObservationRanker) rank(ctx context.Context, cases []pathplan.TestCase) (PathObservationRound, error) {
	if !r.options.ReuseProbeOutputs {
		ranking, err := r.prepared.RankProbes(ctx, cases, r.options.Inputs, r.options.MaxCandidates)
		return PathObservationRound{Ranking: ranking}, err
	}
	var snapshot pathplan.ProbeSnapshot
	var err error
	if r.session == nil {
		r.session, snapshot, err = r.prepared.StartProbeSession(ctx, cases, r.options.Inputs, r.options.MaxCandidates)
	} else {
		snapshot, err = r.session.AppendObservation(ctx, cases[len(cases)-1])
	}
	return PathObservationRound{Ranking: snapshot.Ranking, Reuse: &PathProbeReuse{
		Revision: snapshot.Revision, InitialRankingSHA256: snapshot.InitialRankingSHA256,
		ReusedProbeValues: snapshot.ReusedProbeValues, CachedComparisons: snapshot.CachedComparisons,
		TotalEvaluationAttempts: snapshot.TotalEvaluationAttempts,
		TotalCachedComparisons:  snapshot.TotalCachedComparisons}}, err
}

// The option identifies the replay algorithm. Its omitted/false form preserves
// earlier v1 receipt bytes and their original repeated-evaluation accounting.
// Receipt binding checks accounting; execution reconstructs a fresh session from
// source and replays every round rather than trusting imported cached values.
func pathProbeReuseBound(r *PathObservationReceipt) bool {
	initialSHA, evaluations, comparisons := "", 0, 0
	for i, round := range r.Rounds {
		if !r.Options.ReuseProbeOutputs {
			if round.Reuse != nil {
				return false
			}
			continue
		}
		q, reused := round.Ranking, round.Reuse
		if reused == nil || reused.Revision != i || q.EvaluationAttempts < 0 {
			return false
		}
		values, compared := 0, 0
		if i == 0 {
			raw, _ := json.Marshal(q)
			initialSHA, evaluations = digest(raw), q.EvaluationAttempts
		} else {
			values, compared = len(q.SurvivingMasks)*len(r.Options.Inputs), len(r.Rounds[i-1].Ranking.SurvivingMasks)
			if q.EvaluationAttempts != 0 {
				return false
			}
		}
		comparisons += compared
		if "sha256:"+reused.InitialRankingSHA256 != initialSHA || reused.ReusedProbeValues != values ||
			reused.CachedComparisons != compared || reused.TotalCachedComparisons != comparisons ||
			reused.TotalEvaluationAttempts != evaluations {
			return false
		}
	}
	return true
}
