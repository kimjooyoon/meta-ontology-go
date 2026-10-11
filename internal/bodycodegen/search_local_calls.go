package bodycodegen

import (
	"bytes"
	"context"
	"encoding/json"
)

// LocalCallHistory covers the evaluator passes of this search invocation.
// Separate preflight, other construction profiles and native calls have their
// own boundaries; this history alone cannot establish construction independence.
type LocalCallHistory struct {
	Schema      string                  `json:"schema"`
	Status      string                  `json:"status"`
	Reason      string                  `json:"reason,omitempty"`
	Evaluations []*LocalCallObservation `json:"evaluations"`
}

type localCaptureKey struct{}
type localPhaseKey struct{}
type localEvaluationPhase struct{ phase, candidate string }

type searchCallCapture struct {
	report  Report
	history *LocalCallHistory
	calls   int
	bytes   int
}

func captureSearchCalls(ctx context.Context, filename string, source []byte, activity string) (context.Context, *LocalCallHistory) {
	c := &searchCallCapture{history: &LocalCallHistory{Schema: localCallSchema, Status: "RECORDED",
		Evaluations: []*LocalCallObservation{}}}
	p, err := prepareActivityBodySource(filename, source, activity)
	if err == nil {
		c.report = Report{Activity: activity, ActivityID: p.activityID, RecordTypes: p.allRecords}
		var functions []pureCallFunction
		functions, _, _, err = p.resolvePureCalls(p.body)
		c.report.CallClosure = &PureCallClosure{}
		for _, f := range functions {
			c.report.CallClosure.Activities = append(c.report.CallClosure.Activities, f.identity)
		}
	}
	if err != nil {
		c.history.Status, c.history.Reason = "UNKNOWN", "SOURCE_CALL_BINDING_UNAVAILABLE"
	}
	return context.WithValue(ctx, localCaptureKey{}, c), c.history
}

func localCallPhase(ctx context.Context, phase, candidate string) context.Context {
	return context.WithValue(ctx, localPhaseKey{}, localEvaluationPhase{phase, candidate})
}

func (c *searchCallCapture) evaluate(ctx context.Context, source []byte, activity string,
	cases []IRBodyFillTestCase) ([]IRBodyFillCaseResult, int, error) {
	if len(c.history.Evaluations) >= 512 {
		c.history.Status, c.history.Reason = "UNKNOWN", "EVALUATION_RECORD_LIMIT"
		return evaluateIntegerCasesRecorded(ctx, source, activity, cases, nil)
	}
	r := newLocalCallRecorder(source, c.report, cases)
	r.maxRecords, r.maxBytes = localCallRecordLimit-c.calls, localCallByteLimit-c.bytes
	phase, _ := ctx.Value(localPhaseKey{}).(localEvaluationPhase)
	r.observation.Phase, r.observation.CandidateID = phase.phase, phase.candidate
	if activity != c.report.Activity || phase.phase == "" {
		r.unavailable("EVALUATION_BINDING_UNAVAILABLE")
	}
	results, passed, err := evaluateIntegerCasesRecorded(ctx, source, activity, cases, r)
	observation := r.finish(ctx, err)
	c.history.Evaluations = append(c.history.Evaluations, observation)
	c.calls, c.bytes = c.calls+len(observation.Calls), c.bytes+r.bytes
	if observation.Status != "RECORDED" && c.history.Status == "RECORDED" {
		c.history.Status, c.history.Reason = "UNKNOWN", observation.Reason
	}
	return results, passed, err
}

// EqualSearchCandidateHistory preserves legacy absence without fabricating old
// observations. New histories must match the existing deterministic replay pass.
func EqualSearchCandidateHistory(prior, replayed []SearchCandidate) bool {
	if len(prior) != len(replayed) || (prior == nil) != (replayed == nil) {
		return false
	}
	for i, actual := range replayed {
		if prior[i].LocalCalls == nil {
			actual.LocalCalls = nil
		}
		if !equalLocalCallEvidence(prior[i], actual) {
			return false
		}
	}
	return true
}

// Saved pretty-printed JSON can add whitespace inside raw typed input tuples.
// Marshal compacts those values while preserving their integer precision.
func equalLocalCallEvidence(expected, actual any) bool {
	want, wantErr := json.Marshal(expected)
	got, gotErr := json.Marshal(actual)
	return wantErr == nil && gotErr == nil && bytes.Equal(want, got)
}
