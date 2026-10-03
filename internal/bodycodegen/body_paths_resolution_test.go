package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func TestUniquePathResolutionProjectsWithoutSearchOrModelLoad(t *testing.T) {
	for _, reuse := range []bool{false, true} {
		for _, model := range []string{"", "unused-model-does-not-exist.json"} {
			source, doc, options := observationFixture(t)
			options.Observation.ReuseProbeOutputs, options.Observation.ResolveUniqueCandidate = reuse, true
			options.Diagnosis = &PathDiagnosisOptions{Inputs: []int64{0, 1}, MaxCandidates: 2}
			result, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", doc, model, options)
			if err != nil {
				t.Fatal(err)
			}
			p := result.Report.BodyPaths
			if !pathResolved(p) || p.SearchStarted || !reflect.DeepEqual(p.Search, pathplan.SearchResult{}) ||
				p.Resolution.ModelLoadSkipped != (model != "") || p.Resolution.ModelRankingSkipped != (model != "") ||
				bodyPathSelection(p).ModelCalls != 0 || p.Resolution.ObservedCasesTotal != 2 || p.Diagnosis == nil ||
				!strings.Contains(result.Source, "return (2 - input)") || p.Timing.ResolutionMS <= 0 {
				t.Fatal("unique resolution executed or invented search", p)
			}
			if d := completenessDimensionByID(t, result.Report.CompletenessReceipt, "typed_path_observation_resolution"); d.Status != "PASS" {
				t.Fatal(d)
			}
			if d := completenessDimensionByID(t, result.Report.CompletenessReceipt, "typed_path_candidate_observation"); d.Numerator != 0 || d.Denominator != 0 {
				t.Fatal("reused evidence became fresh search", d)
			}
			if err := VerifyTypedPathProjection(context.Background(), "fixture.gooo", source, doc, result); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestUniquePathResolutionSkipsRetainedModelSeedAndFeedbackExplicitly(t *testing.T) {
	source, doc := threeNativeFixture(t)
	source = append(source, []byte("activity Expected(Integer) -> Integer computes \"return (7 - input)\"\n")...)
	doc.TestCases = []pathplan.TestCase{{Input: 10, Expected: -3}}
	doc.Seed = "requested-seed"
	g, err := NewTypedPathGenerator(writeThreeContractModel(t))
	if err != nil {
		t.Fatal(err)
	}
	options := TypedPathOptions{StepAttempts: 1, FeedbackRounds: 7, FeedbackUnfixed: true,
		Observation: &PathObservationOptions{Inputs: []int64{0, 3, -1}, MaxCandidates: 8, MaxRounds: 2,
			OracleActivity: "Expected", ReuseProbeOutputs: true}}
	fresh, err := g.Generate(context.Background(), "fixture.gooo", source, "ThreeTest", doc, options)
	if err != nil {
		t.Fatal(err)
	}
	options.Observation.ResolveUniqueCandidate = true
	resolved, err := g.Generate(context.Background(), "fixture.gooo", source, "ThreeTest", doc, options)
	if err != nil {
		t.Fatal(err)
	}
	p := resolved.Report.BodyPaths
	if !pathResolved(p) || p.Resolution.ModelLoadSkipped || !p.Resolution.ModelRankingSkipped ||
		!p.Resolution.SeedSkipped || !p.Resolution.FeedbackSkipped || !p.ModelRetention.Loaded ||
		fresh.Report.BodyPaths.Search.Selection.ModelCalls < 1 || len(p.Progress) != 0 || len(p.Feedback) != 0 ||
		p.ModelContext != nil || fresh.Source != resolved.Source || !reflect.DeepEqual(fresh.Report.BodyPaths.NativeCases, p.NativeCases) {
		t.Fatal("retained model work or finite meaning differs", p)
	}
	if err := VerifyTypedPathProjection(context.Background(), "fixture.gooo", source, doc, resolved); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := g.Generate(ctx, "fixture.gooo", source, "ThreeTest", doc, options); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestUniquePathResolutionContinuesSearchForPartialEmptyAndAmbiguousSets(t *testing.T) {
	for _, mode := range []string{"partial", "empty", "ambiguous"} {
		t.Run(mode, func(t *testing.T) {
			source, doc, options := observationFixture(t)
			options.Observation.ResolveUniqueCandidate = true
			want := "PARTIAL_ENUMERATION"
			switch mode {
			case "partial":
				options.Observation.MaxCandidates = 1
			case "empty":
				want = "NO_SURVIVOR"
				options.Observation.OracleActivity = ""
				doc.TestCases[0].Expected = 99
			case "ambiguous":
				want = "AMBIGUOUS"
				options.Observation.Inputs = []int64{2, 2}
			}
			result, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", doc, "", options)
			if err != nil {
				t.Fatal(err)
			}
			p := result.Report.BodyPaths
			if p.Resolution.Status != want || pathResolved(p) || !p.SearchStarted || len(p.Search.Attempts) == 0 ||
				p.Resolution.SearchSkipped || !pathResolutionBound(p) {
				t.Fatal(p)
			}
			if err := VerifyTypedPathProjection(context.Background(), "fixture.gooo", source, doc, result); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUniquePathResolutionReplaysChoicesAndSkippedWork(t *testing.T) {
	source, doc, options := observationFixture(t)
	options.Observation.ResolveUniqueCandidate = true
	result, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", doc, "", options)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(result.Report.BodyPaths)
	for _, mutate := range []func(*BodyPathReceipt){
		func(p *BodyPathReceipt) { p.Resolution = nil },
		func(p *BodyPathReceipt) { *p.Resolution.SelectedMask = 0 },
		func(p *BodyPathReceipt) { p.Resolution.Selection.Choices[doc.Plan.Decisions[0].ID] = "layout_forward" },
		func(p *BodyPathReceipt) { p.Resolution.Selection.Receipts[0].Mode = "model" },
		func(p *BodyPathReceipt) { p.Resolution.Selection.ModelCalls = 1 },
		func(p *BodyPathReceipt) { p.Resolution.SeedSkipped = true },
		func(p *BodyPathReceipt) { p.Resolution.FeedbackSkipped = true },
		func(p *BodyPathReceipt) { p.Resolution.ModelLoadSkipped = true },
		func(p *BodyPathReceipt) { p.Resolution.ObservedCasesPassed-- },
		func(p *BodyPathReceipt) { p.Resolution.RankingSHA256 = "changed" },
		func(p *BodyPathReceipt) { p.SearchStarted = true },
		func(p *BodyPathReceipt) { p.Search.Attempts = append(p.Search.Attempts, pathplan.SearchAttempt{}) },
	} {
		var p BodyPathReceipt
		if err := json.Unmarshal(raw, &p); err != nil {
			t.Fatal(err)
		}
		mutate(&p)
		if err := replayPathResolution(doc, &p); err == nil {
			t.Fatal("changed resolution accepted")
		}
	}
}

func TestUniquePathResolutionUsesOriginalCasesAndHonorsOracleFailure(t *testing.T) {
	source, doc, options := observationFixture(t)
	options.Observation.ResolveUniqueCandidate = true
	options.Observation.OracleActivity = ""
	doc.TestCases = []pathplan.TestCase{{Input: 3, Expected: -1}}
	result, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", doc, "", options)
	if err != nil {
		t.Fatal(err)
	}
	if !pathResolved(result.Report.BodyPaths) || result.Report.BodyPaths.Observation.OracleEvaluations != 0 {
		t.Fatal("already unique original suite invoked an oracle")
	}
	if err := VerifyTypedPathProjection(context.Background(), "fixture.gooo", source, doc, result); err != nil {
		t.Fatal(err)
	}
	options.Observation.OracleActivity = "Expected"
	doc.TestCases[0].Expected = 99
	_, err = GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", doc, "missing-model.json", options)
	var failure *BodyPathError
	if !errors.As(err, &failure) || failure.Receipt.Resolution != nil || failure.Receipt.SearchStarted ||
		!strings.Contains(err.Error(), "contradicts") {
		t.Fatal("resolution bypassed oracle contradiction", err)
	}
}
