package executioncontinuation

import (
	"fmt"
	"io/fs"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation/policy"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

type CaseResult struct {
	ID       string                        `json:"id"`
	Input    int64                         `json:"input"`
	Expected int64                         `json:"expected_output"`
	Actual   int64                         `json:"actual_output"`
	Passed   bool                          `json:"passed"`
	Elapsed  int64                         `json:"elapsed_nanoseconds"`
	Origin   valueexecution.ResultEvidence `json:"origin"`
	Error    string                        `json:"error,omitempty"`
}

type Consumption struct {
	Schema                 string       `json:"schema"`
	Stage                  string       `json:"stage"`
	Metadata               Metadata     `json:"metadata"`
	GrantID                string       `json:"grant_id"`
	GrantRequestDigest     string       `json:"grant_request_digest"`
	ReservationDigest      string       `json:"reservation_digest"`
	ExecutionInputDigest   string       `json:"execution_input_digest"`
	PolicyDigest           string       `json:"policy_digest"`
	ExecutionCount         int          `json:"execution_count"`
	Invocations            int          `json:"evaluator_invocations"`
	ConsumedUses           int          `json:"consumed_uses"`
	RemainingUses          int          `json:"remaining_uses"`
	OneUseEnforced         bool         `json:"one_use_enforced"`
	EnforcementScope       string       `json:"enforcement_scope"`
	RepositoryWrites       int          `json:"repository_writes"`
	ExternalEffects        []string     `json:"external_effects"`
	CaseCount              int          `json:"case_count"`
	PassedCases            int          `json:"passed_cases"`
	Cases                  []CaseResult `json:"cases"`
	CompileElapsed         int64        `json:"compile_elapsed_nanoseconds"`
	EvaluationElapsed      int64        `json:"evaluation_elapsed_nanoseconds"`
	PerformanceImprovement string       `json:"performance_improvement"`
	Digest                 string       `json:"digest"`
}

// consume executes only through the lease in execution.go.
func consume(root fs.FS, p Prepared, reservation Reservation) (Consumption, error) {
	in, err := validatePrepared(root, p)
	if err != nil {
		return Consumption{}, err
	}
	if err := validateReservation(p, reservation); err != nil {
		return Consumption{}, err
	}
	r := Consumption{Schema: ConsumptionSchema, Stage: policy.ObserveStage(true, true, false), Metadata: p.Metadata,
		GrantID: reservation.GrantID, GrantRequestDigest: reservation.RequestDigest, ReservationDigest: reservation.Digest,
		ExecutionInputDigest: in.Digest, PolicyDigest: policy.SourceDigest(), ConsumedUses: 1, OneUseEnforced: true,
		EnforcementScope: ReservationScope, ExternalEffects: []string{}, CaseCount: len(in.Corpus), Cases: []CaseResult{}, PerformanceImprovement: "UNKNOWN"}
	started := time.Now()
	program, err := valueexecution.Compile(in.Source.Path, []byte(in.Source.Bytes), in.Activity.Name)
	r.CompileElapsed = time.Since(started).Nanoseconds()
	if err != nil {
		return sealConsumption(r), err
	}
	if program.SourceDigest != in.Source.Digest || program.SemanticFingerprint != in.Activity.SemanticFingerprint {
		return sealConsumption(r), refuted("EXECUTOR_SOURCE_MISMATCH")
	}
	r.ExecutionCount = 1
	for _, test := range in.Corpus {
		item := executeCase(program, test.ID, test.Input, test.ExpectedOutput)
		r.Cases = append(r.Cases, item)
		r.Invocations++
		r.EvaluationElapsed += item.Elapsed
		if item.Passed {
			r.PassedCases++
		}
	}
	r.Stage = policy.ObserveStage(true, true, r.Invocations > 0)
	r = sealConsumption(r)
	if r.PassedCases != r.CaseCount {
		return r, fmt.Errorf("bounded Gooo execution incomplete: %d/%d", r.PassedCases, r.CaseCount)
	}
	return r, nil
}

func executeCase(program valueexecution.Program, id string, input, expected int64) CaseResult {
	item := CaseResult{ID: id, Input: input, Expected: expected}
	started := time.Now()
	result, err := program.ExecuteResult([]int64{input})
	item.Elapsed = time.Since(started).Nanoseconds()
	if err != nil {
		item.Error = err.Error()
		return item
	}
	if err := program.ValidateProducedResult(result); err != nil {
		item.Error = err.Error()
		return item
	}
	value, err := result.Integer()
	if err != nil {
		item.Error = err.Error()
		return item
	}
	item.Actual = int64(value)
	item.Origin = result.Evidence()
	item.Passed = item.Actual == expected
	if !item.Passed {
		item.Error = "value differs from source-declared corpus"
	}
	return item
}

func sealConsumption(r Consumption) Consumption { r.Digest = ""; r.Digest = digestValue(r); return r }
