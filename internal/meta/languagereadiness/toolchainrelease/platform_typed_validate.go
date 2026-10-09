package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

type typedSmokeOutput struct {
	Generated    *bool `json:"generated_now"`
	Construction typedSmokeConstruction
	Evaluation   struct {
		Runtime    jointSmokeRuntime
		Replayed   *bool                              `json:"construction_replayed"`
		Calls      *int                               `json:"new_model_calls"`
		Separation bodyexecution.JointInputSeparation `json:"input_separation"`
	}
}

type typedSmokeConstruction struct {
	bodyexecution.JointConstruction
	Attempts []struct {
		bodyexecution.JointAttempt
		Runtime jointSmokeRuntime `json:"runtime"`
	} `json:"attempts"`
}

func validateTypedSmoke(raw, source, feedback, cases []byte, contextSHA string, replay bool, selected string) (string, error) {
	var r typedSmokeOutput
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", err
	}
	c, e := r.Construction, r.Evaluation
	declared, err := json.Marshal(c.ConstructionCases)
	if err != nil {
		return "", err
	}
	if !jointSmokeBool(r.Generated, !replay) || c.Schema != "gooo/joint-construction/v7" ||
		c.Stage != "COMPLETE" || c.Failure != "" || c.Decision != "COMPLETE_FINITE" ||
		c.StopReason != "LOCAL_AND_CALLER_CASES_MATCHED" || c.ProgramBudget != 16 || c.CandidateSpace != "16" ||
		c.SelectedAttempt != 10 || len(c.Attempts) != 11 || c.SelectedSource == "" || c.Selected.GeneratedSHA256 == "" ||
		c.OriginalSourceSHA256 != scalarPreflightDigest(source) || !samePackageSourceValue(declared, feedback) ||
		!reflect.DeepEqual(c.CandidateKinds, []string{"typed_path_mask", "record_mask"}) ||
		!jointSmokeBool(e.Replayed, replay) || !jointSmokeInt(e.Calls, 0) ||
		e.Separation.UniqueInputs != 3 || e.Separation.DuplicateRows != 0 ||
		e.Separation.ConstructionInputs != 0 || e.Separation.OtherInputs != 3 {
		return "", fmt.Errorf("typed construction source, budget, history or evaluation scope differs")
	}
	if replay && (selected == "" || selected != c.Selected.GeneratedSHA256) {
		return "", fmt.Errorf("typed replay changed the selected program")
	}
	if err := validateTypedModelBinding(c.Initial, contextSHA); err != nil {
		return "", err
	}
	if err := validateTypedPredictionCounter(raw); err != nil {
		return "", err
	}
	if err := validateTypedAttempts(c, e.Runtime.SHA); err != nil {
		return "", err
	}
	if err := validateTypedRuntime(e.Runtime, cases); err != nil {
		return "", err
	}
	return c.Selected.GeneratedSHA256, nil
}

func validateTypedAttempts(c typedSmokeConstruction, runtimeSHA string) error {
	for i, a := range c.Attempts {
		passed := 0
		if i == c.SelectedAttempt {
			passed = 1
		}
		if a.LocalTotal != 2 || a.LocalPassed < 0 || a.LocalPassed > 2 || a.Rejection != nil ||
			len(a.PathCandidates) != 1 || a.Runtime.Stage != "COMPLETE" || a.Runtime.Failure != "" ||
			!jointSmokeInt(a.Runtime.Passed, passed) || !jointSmokeInt(a.Runtime.Total, 1) ||
			!jointSmokeInt(a.Runtime.Calls, 0) || !jointSmokeBool(a.Runtime.Projection, true) ||
			!jointSmokeBool(a.Runtime.Replay, true) || len(a.Runtime.Traces) != 1 {
			return fmt.Errorf("typed attempt %d lost local cases, caller observation or zero inference", i)
		}
		trace := a.Runtime.Traces[0]
		if trace.Index != 0 || len(trace.Deliveries) != 1 {
			return fmt.Errorf("typed attempt %d lost its consumed caller row", i)
		}
		d := trace.Deliveries[0]
		if d.ID != "callerpaths://activity/main" || !samePackageSourceValue(d.Input, []byte("3")) ||
			!samePackageSourceValue(d.Expected, []byte("9")) || !jointSmokeBool(d.Passed, passed == 1) ||
			samePackageSourceValue(d.Actual, d.Expected) != (passed == 1) {
			return fmt.Errorf("typed attempt %d changed its consumed caller value", i)
		}
	}
	first, last := c.Attempts[0], c.Attempts[c.SelectedAttempt]
	if first.LocalPassed != 2 || last.LocalPassed != 2 || last.Runtime.SHA != c.Selected.GeneratedSHA256 ||
		runtimeSHA != c.Selected.GeneratedSHA256 {
		return fmt.Errorf("typed local/caller result or selected program digest differs")
	}
	return nil
}
