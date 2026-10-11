package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/types"
)

const localCallSchema = "gooo/local-call-inputs/v1"
const localCallRecordLimit = 8192
const localCallByteLimit = 1 << 20

// LocalCallObservation records entries reached by one original evaluator pass.
// RECORDED describes capture coverage, not successful scoring or input independence.
// A nil observation means that the producer did not record this pass.
type LocalCallObservation struct {
	Schema              string           `json:"schema"`
	GeneratedSHA256     string           `json:"generated_sha256"`
	ActivityID          string           `json:"activity_id"`
	SuiteSHA256         string           `json:"suite_sha256"`
	Phase               string           `json:"phase,omitempty"`
	CandidateID         string           `json:"candidate_id,omitempty"`
	Status              string           `json:"status"`
	Reason              string           `json:"reason,omitempty"`
	EvaluationCompleted bool             `json:"evaluation_completed"`
	EvaluationError     string           `json:"evaluation_error,omitempty"`
	CasesStarted        int              `json:"cases_started"`
	CasesCompleted      int              `json:"cases_completed"`
	Calls               []LocalCallInput `json:"calls"`
}

type LocalCallInput struct {
	CaseIndex     int             `json:"case_index"`
	ActivityID    string          `json:"activity_id"`
	ProgramSHA256 string          `json:"program_sha256"`
	Inputs        json.RawMessage `json:"inputs"`
}

type localCallRecorder struct {
	observation *LocalCallObservation
	identities  map[string]PureCallActivity
	records     []RecordType
	bytes       int
	maxRecords  int
	maxBytes    int
}

func newLocalCallRecorder(source []byte, report Report, cases any) *localCallRecorder {
	raw, err := json.Marshal(cases)
	r := &localCallRecorder{observation: &LocalCallObservation{
		Schema: localCallSchema, GeneratedSHA256: digest(source), ActivityID: report.ActivityID,
		SuiteSHA256: digest(raw), Status: "RECORDED", Calls: []LocalCallInput{}},
		identities: make(map[string]PureCallActivity), records: report.RecordTypes,
		maxRecords: localCallRecordLimit, maxBytes: localCallByteLimit}
	if err != nil || report.ActivityID == "" {
		r.unavailable("EVALUATION_BINDING_UNAVAILABLE")
	}
	if report.CallClosure != nil {
		for _, activity := range report.CallClosure.Activities {
			if _, duplicate := r.identities[activity.Name]; duplicate || activity.ActivityID == "" || activity.ProgramSHA256 == "" {
				r.unavailable("CALL_IDENTITY_UNAVAILABLE")
			}
			r.identities[activity.Name] = activity
		}
	}
	return r
}

func (r *localCallRecorder) unavailable(reason string) {
	if r.observation.Status == "RECORDED" {
		r.observation.Status, r.observation.Reason = "UNKNOWN", reason
	}
}

func (r *localCallRecorder) finish(ctx context.Context, err error) *LocalCallObservation {
	r.observation.EvaluationCompleted = err == nil
	if err != nil {
		r.observation.EvaluationError = err.Error()
	}
	if ctx.Err() != nil {
		r.unavailable("CALLER_CANCELLED")
	}
	return r.observation
}

func (r *localCallRecorder) enter(caseIndex int, function *ast.FuncDecl, frame *integerBodyEvaluator) {
	if r == nil || r.observation.Status != "RECORDED" {
		return
	}
	if len(r.observation.Calls) >= r.maxRecords {
		r.unavailable("CALL_RECORD_LIMIT")
		return
	}
	identity, ok := r.identities[function.Name.Name]
	if !ok {
		r.unavailable("CALL_IDENTITY_UNAVAILABLE")
		return
	}
	values := make([]any, 0, len(function.Type.Params.List))
	for _, field := range function.Type.Params.List {
		for _, name := range field.Names {
			parameter := frame.information.Defs[name]
			value, err := localCallValue(frame.environment[parameter], parameter.Type(), r.records)
			if err != nil {
				r.unavailable("CALL_VALUE_UNAVAILABLE")
				return
			}
			values = append(values, value)
		}
	}
	raw, err := json.Marshal(values)
	if err != nil {
		r.unavailable("CALL_VALUE_UNAVAILABLE")
		return
	}
	if len(raw) > r.maxBytes-r.bytes {
		r.unavailable("CALL_INPUT_BYTE_LIMIT")
		return
	}
	r.bytes += len(raw)
	r.observation.Calls = append(r.observation.Calls, LocalCallInput{
		CaseIndex: caseIndex, ActivityID: identity.ActivityID, ProgramSHA256: identity.ProgramSHA256, Inputs: raw})
}

// Read already coerced values. No expression is evaluated for observation.
func localCallValue(value any, valueType types.Type, records []RecordType) (any, error) {
	if record, ok := value.(recordBodyValue); ok {
		named, ok := valueType.(*types.Named)
		if !ok || named.Obj().Name() != record.Type {
			return nil, fmt.Errorf("local call record type differs")
		}
		layout := recordTypeByName(records, record.Type)
		if layout == nil || len(layout.Fields) != record.Count {
			return nil, fmt.Errorf("local call record layout is unavailable")
		}
		object := make(map[string]any, record.Count)
		for i, field := range layout.Fields {
			if record.Values[i].Present {
				object[field.Name] = recordScalarJSONValue(record.Values[i])
			} else {
				object[field.Name] = nil
			}
		}
		return object, nil
	}
	return coerceBodyValue(value, valueType)
}

func evaluateIntegerCasesObserved(ctx context.Context, source []byte, report Report,
	cases []IRBodyFillTestCase) ([]IRBodyFillCaseResult, int, *LocalCallObservation, error) {
	r := newLocalCallRecorder(source, report, cases)
	results, passed, err := evaluateIntegerCasesRecorded(ctx, source, report.Activity, cases, r)
	return results, passed, r.finish(ctx, err), err
}
