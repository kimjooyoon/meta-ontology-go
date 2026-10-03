package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// Resolution is separate from Search: reused finite observations are not new
// candidate tests or model predictions. Final native emission still checks them.
type PathResolutionReceipt struct {
	Schema               string              `json:"schema"`
	Status               string              `json:"status"`
	RankingSHA256        string              `json:"ranking_sha256"`
	EffectiveCasesSHA256 string              `json:"effective_cases_sha256"`
	SelectedMask         *uint16             `json:"selected_mask,omitempty"`
	Selection            *pathplan.Selection `json:"selection,omitempty"`
	ObservedCasesPassed  int                 `json:"observed_cases_passed"`
	ObservedCasesTotal   int                 `json:"observed_cases_total"`
	SearchSkipped        bool                `json:"search_skipped"`
	ModelRankingSkipped  bool                `json:"model_ranking_skipped"`
	ModelLoadSkipped     bool                `json:"model_load_skipped"`
	SeedSkipped          bool                `json:"seed_skipped"`
	FeedbackSkipped      bool                `json:"feedback_skipped"`
}

func pathResolved(p *BodyPathReceipt) bool {
	return p.Resolution != nil && p.Resolution.Status == "RESOLVED" &&
		p.Resolution.SelectedMask != nil && p.Resolution.Selection != nil
}

func bodyPathSelection(p *BodyPathReceipt) pathplan.Selection {
	if pathResolved(p) {
		return *p.Resolution.Selection
	}
	return p.Search.Selection
}

func bodyPathSelectedScore(p *BodyPathReceipt) (int, int) {
	if pathResolved(p) {
		return p.Resolution.ObservedCasesPassed, p.Resolution.ObservedCasesTotal
	}
	return p.Search.SelectedTrainingPassed, p.Search.TrainingTotal
}

func pathResolutionStatus(q pathplan.ProbeRanking) string {
	if q.Status != "COMPLETE" || q.Unobserved != 0 {
		return "PARTIAL_ENUMERATION"
	}
	if len(q.SurvivingMasks) == 0 {
		return "NO_SURVIVOR"
	}
	if len(q.SurvivingMasks) != 1 {
		return "AMBIGUOUS"
	}
	return "RESOLVED"
}

func resolutionFeedbackRequested(p *BodyPathReceipt) bool {
	var config struct {
		Rounds int `json:"feedback_rounds"`
	}
	return json.Unmarshal(p.SearchConfig, &config) == nil && config.Rounds > 0
}

func makePathResolution(document pathplan.Document, p *BodyPathReceipt) (*PathResolutionReceipt, error) {
	o := p.Observation
	if o == nil || !o.Options.ResolveUniqueCandidate {
		return nil, nil
	}
	if len(o.Rounds) == 0 {
		return nil, fmt.Errorf("resolution requires a completed observation")
	}
	q := o.Rounds[len(o.Rounds)-1].Ranking
	raw, _ := json.Marshal(q)
	r := &PathResolutionReceipt{Schema: "gooo/typed-path-observation-resolution/v1",
		Status: pathResolutionStatus(q), RankingSHA256: digest(raw), EffectiveCasesSHA256: o.EffectiveCasesSHA}
	if r.Status != "RESOLVED" {
		return r, nil
	}
	mask := q.SurvivingMasks[0]
	if int(mask) >= 1<<len(document.Plan.Decisions) {
		return nil, fmt.Errorf("resolved mask exceeds declared choices")
	}
	s := &pathplan.Selection{Schema: "gooo/typed-body-path-selection/v1", PlanSHA256: o.PreparedPlanSHA256,
		Choices: make(map[string]string, len(document.Plan.Decisions)), ExternalCallsKnown: true}
	for i, choice := range document.Plan.Decisions {
		label := choice.Options[int(mask>>i&1)].Label
		s.Choices[choice.ID] = label
		s.Receipts = append(s.Receipts, pathplan.Receipt{ID: choice.ID, Kind: choice.Kind,
			IntentSHA256: strings.TrimPrefix(digest([]byte(choice.Intent)), "sha256:"),
			Selected:     label, Mode: "unique_source_observation"})
	}
	r.SelectedMask, r.Selection = &mask, s
	r.ObservedCasesPassed, r.ObservedCasesTotal = o.EffectiveCases, o.EffectiveCases
	r.SearchSkipped, r.ModelRankingSkipped = true, p.LocalModelRequested
	r.ModelLoadSkipped = p.LocalModelRequested && p.ModelRetention == nil
	r.SeedSkipped, r.FeedbackSkipped = document.Seed != "", resolutionFeedbackRequested(p)
	return r, nil
}

func resolveObservedPath(ctx context.Context, document pathplan.Document, prepared *pathplan.PreparedPlan,
	p *BodyPathReceipt) (*bodyplan.Program, error) {
	if p.Observation == nil || !p.Observation.Options.ResolveUniqueCandidate {
		return nil, nil
	}
	started := time.Now()
	defer func() { p.Timing.ResolutionMS = elapsedMS(started) }()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p.Observation.PreparedPlanSHA256 != prepared.PlanSHA256() {
		return nil, fmt.Errorf("resolved plan differs")
	}
	r, err := makePathResolution(document, p)
	if err != nil {
		return nil, err
	}
	p.Resolution = r
	if !pathResolved(p) {
		return nil, nil
	}
	program, err := prepared.Compile(r.Selection.Choices)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p.FeedbackUnfixed = false
	p.Timing.ExecutionModel = "source_bind_then_oracle_observations_then_unique_selection_then_native_emit"
	p.Timing.DecisionStage = "complete_unique_candidate_observation; no_model_ranking_or_candidate_search"
	return program, nil
}

func pathResolutionBound(p *BodyPathReceipt) bool {
	r, o := p.Resolution, p.Observation
	if r == nil {
		return o == nil || !o.Options.ResolveUniqueCandidate
	}
	if o == nil || !o.Options.ResolveUniqueCandidate || len(o.Rounds) == 0 {
		return false
	}
	q := o.Rounds[len(o.Rounds)-1].Ranking
	raw, _ := json.Marshal(q)
	if r.Schema != "gooo/typed-path-observation-resolution/v1" || r.Status != pathResolutionStatus(q) ||
		r.RankingSHA256 != digest(raw) || r.EffectiveCasesSHA256 != o.EffectiveCasesSHA {
		return false
	}
	if r.Status != "RESOLVED" {
		return r.SelectedMask == nil && r.Selection == nil && r.ObservedCasesPassed == 0 &&
			r.ObservedCasesTotal == 0 && !r.SearchSkipped && !r.ModelRankingSkipped &&
			!r.ModelLoadSkipped && !r.SeedSkipped && !r.FeedbackSkipped
	}
	if !pathResolved(p) || *r.SelectedMask != q.SurvivingMasks[0] || p.SearchStarted ||
		!reflect.DeepEqual(p.Search, pathplan.SearchResult{}) || len(p.Progress) != 0 || len(p.Feedback) != 0 ||
		p.ModelContext != nil || p.FeedbackUnfixed || p.Timing.BoundedSearchMS != 0 ||
		p.Timing.ModelLoadMS != 0 || p.Timing.ContextPrepareMS != 0 {
		return false
	}
	s := r.Selection
	return s.Schema == "gooo/typed-body-path-selection/v1" && s.PlanSHA256 == o.PreparedPlanSHA256 &&
		s.ModelCalls == 0 && s.ExternalCalls == 0 && s.ExternalCallsKnown && s.MetadataSHA256 == "" &&
		s.WeightsSHA256 == "" && s.ModelVariant == "" && s.SeedSHA256 == "" && s.Joint == nil && s.Three == nil &&
		r.SearchSkipped && r.ModelRankingSkipped == p.LocalModelRequested &&
		r.ModelLoadSkipped == (p.LocalModelRequested && p.ModelRetention == nil) &&
		r.FeedbackSkipped == resolutionFeedbackRequested(p) &&
		r.ObservedCasesPassed == o.EffectiveCases && r.ObservedCasesTotal == o.EffectiveCases
}

func pathResolutionDimension(p *BodyPathReceipt) CompletenessDimension {
	bound := p.Resolution != nil && pathResolutionBound(p)
	status := "UNOBSERVED"
	if p.Resolution != nil {
		status = p.Resolution.Status
	}
	return completenessDimension("typed_path_observation_resolution", boolCount(bound), 1,
		"recorded use of a complete unique finite candidate or explicit continuation of search",
		"Complete enumeration and one survivor permit direct projection; final finite checks and source replay remain required.",
		[]string{"body_paths.resolution", "resolution_status:" + status}, p.Resolution != nil && !bound)
}

func replayPathResolution(document pathplan.Document, p *BodyPathReceipt) error {
	expected, err := makePathResolution(document, p)
	if err != nil {
		return err
	}
	if !pathResolutionBound(p) || !reflect.DeepEqual(expected, p.Resolution) {
		return fmt.Errorf("typed-path resolution or skipped-work accounting does not replay")
	}
	return nil
}
