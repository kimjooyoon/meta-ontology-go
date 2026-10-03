package bodycodegen

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestPathProbeReusePreservesBodiesAndFiniteObservationMeaning(t *testing.T) {
	for _, variant := range []string{"resolved", "missing", "partial", "agreement", "model-feedback"} {
		t.Run(variant, func(t *testing.T) {
			source, doc, options := observationFixture(t)
			model := ""
			switch variant {
			case "missing":
				options.Observation.OracleActivity = ""
			case "partial":
				options.Observation.MaxCandidates = 1
			case "agreement":
				options.Observation.Inputs = []int64{2, 2}
			case "model-feedback":
				source, doc = threeNativeFixture(t)
				source = append(source, []byte("activity Expected(Integer) -> Integer computes \"return (7 - input)\"\n")...)
				doc.TestCases = doc.TestCases[:1]
				doc.TestCases[0].Input, doc.TestCases[0].Expected = 10, -3
				options.StepAttempts, options.FeedbackRounds = 1, 7
				options.Observation.Inputs, options.Observation.MaxCandidates = []int64{0, 3, -1}, 8
				model = writeThreeContractModel(t)
			}
			activity := doc.Plan.Base.Name
			fresh, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, activity, doc, model, options)
			if err != nil {
				t.Fatal(err)
			}
			options.Observation.ReuseProbeOutputs = true
			reused, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, activity, doc, model, options)
			if err != nil {
				t.Fatal(err)
			}
			a, b := fresh.Report.BodyPaths, reused.Report.BodyPaths
			if fresh.Source != reused.Source || !reflect.DeepEqual(a.NativeCases, b.NativeCases) ||
				a.Search.Selection.ModelCalls != b.Search.Selection.ModelCalls ||
				!reflect.DeepEqual(a.Search.Selection.Choices, b.Search.Selection.Choices) ||
				a.Observation.Status != b.Observation.Status ||
				a.Observation.EffectiveCasesSHA != b.Observation.EffectiveCasesSHA ||
				len(a.Observation.Rounds) != len(b.Observation.Rounds) {
				t.Fatal("reuse changed selected behavior or finite evidence")
			}
			for i, round := range a.Observation.Rounds {
				x, y := round.Ranking, b.Observation.Rounds[i].Ranking
				x.EvaluationAttempts, y.EvaluationAttempts = 0, 0
				if !reflect.DeepEqual(x, y) || !reflect.DeepEqual(round.Observation, b.Observation.Rounds[i].Observation) {
					t.Fatal("reuse changed candidate partitions or oracle observation", i)
				}
			}
			if _, _, bad := observedPathContract(b); bad {
				t.Fatal("invalid reuse binding")
			}
			for _, result := range []Result{fresh, reused} {
				if err := VerifyTypedPathProjection(context.Background(), "fixture.gooo", source, doc, result); err != nil {
					t.Fatal("old or reused observation did not replay", err)
				}
			}
		})
	}
}

func TestPathProbeReuseBindsAccountingAndReplaysOutputs(t *testing.T) {
	source, doc, options := observationFixture(t)
	options.Observation.ReuseProbeOutputs = true
	result, err := GenerateWithTypedPathOptions(context.Background(), "fixture.gooo", source, "Probe", doc, "", options)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(result.Report.BodyPaths)
	for _, mutate := range []func(*PathObservationReceipt){
		func(r *PathObservationReceipt) { r.Rounds[1].Reuse.Revision++ },
		func(r *PathObservationReceipt) { r.Rounds[1].Reuse.InitialRankingSHA256 = "changed" },
		func(r *PathObservationReceipt) { r.Rounds[1].Reuse.ReusedProbeValues++ },
		func(r *PathObservationReceipt) { r.Rounds[1].Reuse.CachedComparisons++ },
		func(r *PathObservationReceipt) { r.Rounds[1].Reuse.TotalCachedComparisons++ },
		func(r *PathObservationReceipt) { r.Rounds[1].Reuse.TotalEvaluationAttempts++ },
		func(r *PathObservationReceipt) { r.Rounds[1].Ranking.EvaluationAttempts++ },
		func(r *PathObservationReceipt) { r.Rounds[1].Reuse = nil },
		func(r *PathObservationReceipt) { r.Options.ReuseProbeOutputs = false },
	} {
		var p BodyPathReceipt
		if err := json.Unmarshal(raw, &p); err != nil {
			t.Fatal(err)
		}
		mutate(p.Observation)
		if _, _, bad := observedPathContract(&p); !bad {
			t.Fatal("changed reuse accounting accepted")
		}
	}
	prepared, err := doc.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	result.Report.BodyPaths.Observation.Rounds[1].Ranking.Probes[1].Outputs[0]++
	if _, err := replayPathObservation(context.Background(), "fixture.gooo", source, "Probe", prepared, doc.TestCases, result.Report.BodyPaths); err == nil {
		t.Fatal("imported cache value trusted during replay")
	}
}

func TestPathProbeReuseOmittedOptionPreservesV1Bytes(t *testing.T) {
	_, _, options := observationFixture(t)
	raw, err := json.Marshal(options.Observation)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"inputs":[2,3,0],"max_candidates":2,"max_rounds":2,"oracle_activity":"Expected"}`
	if string(raw) != want {
		t.Fatal("old receipt option hash changed", string(raw))
	}
	if !pathProbeReuseBound(&PathObservationReceipt{Rounds: []PathObservationRound{{}}}) ||
		pathProbeReuseBound(&PathObservationReceipt{Rounds: []PathObservationRound{{Reuse: &PathProbeReuse{}}}}) {
		t.Fatal("reuse mode no longer explicit")
	}
}
