package bodycodegen

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

type fillCaptureKey struct{}

// A candidate can introduce a helper absent from the original hole skeleton.
// Bind each evaluator pass to its actual generated closure, never to that skeleton.
type fillCallCapture struct {
	history *LocalCallHistory
	reports map[string]Report
	calls   int
	bytes   int
}

func captureFillCalls(ctx context.Context) (context.Context, *fillCallCapture) {
	c := &fillCallCapture{history: &LocalCallHistory{Schema: localCallSchema, Status: "RECORDED",
		Evaluations: []*LocalCallObservation{}}, reports: make(map[string]Report)}
	return context.WithValue(ctx, fillCaptureKey{}, c), c
}

func (c *fillCallCapture) bind(source []byte, report Report) {
	c.reports[digest(source)] = report
}

func (c *fillCallCapture) begin(ctx context.Context, source []byte, activity string, cases any) *localCallRecorder {
	if len(c.history.Evaluations) >= 512 {
		c.history.Status, c.history.Reason = "UNKNOWN", "EVALUATION_RECORD_LIMIT"
		return nil
	}
	report, bound := c.reports[digest(source)]
	r := newLocalCallRecorder(source, report, cases)
	r.maxRecords, r.maxBytes = localCallRecordLimit-c.calls, localCallByteLimit-c.bytes
	phase, _ := ctx.Value(localPhaseKey{}).(localEvaluationPhase)
	r.observation.Phase, r.observation.CandidateID = phase.phase, phase.candidate
	if !bound || report.Activity != activity || phase.phase == "" || phase.candidate == "" {
		r.unavailable("EVALUATION_BINDING_UNAVAILABLE")
	}
	return r
}

func (c *fillCallCapture) finish(ctx context.Context, r *localCallRecorder, err error) {
	if r == nil {
		return
	}
	observation := r.finish(ctx, err)
	c.history.Evaluations = append(c.history.Evaluations, observation)
	c.calls, c.bytes = c.calls+len(observation.Calls), c.bytes+r.bytes
	if observation.Status != "RECORDED" && c.history.Status == "RECORDED" {
		c.history.Status, c.history.Reason = "UNKNOWN", observation.Reason
	}
}

func (c *fillCallCapture) integer(ctx context.Context, source []byte, activity string,
	cases []IRBodyFillTestCase) ([]IRBodyFillCaseResult, int, error) {
	r := c.begin(ctx, source, activity, cases)
	rows, passed, err := evaluateIntegerCasesRecorded(ctx, source, activity, cases, r)
	c.finish(ctx, r, err)
	return rows, passed, err
}

func (c *fillCallCapture) record(ctx context.Context, source []byte, activity string, records []RecordType,
	cases []assemblyspec.ValueCase) ([]RecordAssemblyCase, error) {
	r := c.begin(ctx, source, activity, cases)
	rows, err := evaluateRecordAssemblyRecorded(ctx, source, activity, records, cases, r)
	c.finish(ctx, r, err)
	return rows, err
}

// Legacy observations keep their original absence. New typed histories are
// compared after compact JSON encoding, preserving int64 values and tuple order.
func EqualFillCandidateHistory(prior, actual []FillCandidate) bool {
	if len(prior) != len(actual) || (prior == nil) != (actual == nil) {
		return false
	}
	for i, observed := range actual {
		if prior[i].LocalCalls == nil {
			observed.LocalCalls = nil
		}
		if !equalLocalCallEvidence(prior[i], observed) {
			return false
		}
	}
	return true
}
