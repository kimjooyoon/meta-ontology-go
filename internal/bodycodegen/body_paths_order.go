package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/orderjudge"
	"github.com/kimjooyoon/gooo-decision-runtime/orderprepared"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/fileopen"
)

type retainedOrderModel struct {
	runtime  orderprepared.Runtime
	prepared atomic.Pointer[orderprepared.Prepared]
}

type OrderPreparationReceipt struct {
	Schema         string  `json:"schema"`
	PlanSHA256     string  `json:"plan_sha256"`
	Reused         bool    `json:"reused"`
	CandidateCount int     `json:"candidate_count"`
	AcquireMS      float64 `json:"acquire_ms"`
}

// Each model owner holds at most one prepared plan. Concurrent misses may each
// prepare, but never wait on another caller or publish incomplete preparation.
func (model *retainedOrderModel) acquire(ctx context.Context, plan pathplan.Plan, boundSHA string) (*orderprepared.Prepared, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	if p := model.prepared.Load(); p != nil && p.PlanSHA256() == boundSHA {
		return p, true, nil
	}
	p, err := model.runtime.Prepare(ctx, plan)
	if err != nil {
		return nil, false, err
	}
	if p.PlanSHA256() != boundSHA {
		return nil, false, fmt.Errorf("prepared order plan differs from source-bound plan")
	}
	model.prepared.Store(p)
	return p, false, nil
}

func loadOrderJudge(name string, metadata []byte) (*retainedOrderModel, error) {
	if len(metadata) > 4096 {
		return nil, fmt.Errorf("whole-candidate metadata exceeds 4096 bytes")
	}
	f, err := fileopen.ReadOnly(filepath.Join(filepath.Dir(name), "weights.bin"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	const size = orderjudge.ParameterCount * 4
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() != size {
		return nil, fmt.Errorf("whole-candidate weights require a regular %d-byte file", size)
	}
	weights, err := io.ReadAll(io.LimitReader(f, size+1))
	if err != nil {
		return nil, err
	}
	loaded, err := orderjudge.Load(metadata, weights)
	if err != nil {
		return nil, err
	}
	runtime, err := orderprepared.NewRuntime(loaded)
	if err != nil {
		return nil, err
	}
	return &retainedOrderModel{runtime: runtime}, nil
}

// The caller has already bound the complete original plan to the source. The
// SDK sees original intent and actual operations, with no old context wrapper.
func generateOrderTypedPath(ctx context.Context, filename string, source []byte, activity string,
	document pathplan.Document, prepared *pathplan.PreparedPlan, bound typedPathSource,
	model *retainedOrderModel, diagnosis *PathDiagnosisOptions, receipt *BodyPathReceipt, started time.Time) (Result, error) {
	fail := func(err error) (Result, error) {
		receipt.Timing.TotalMS = elapsedMS(started)
		return Result{}, &BodyPathError{Receipt: receipt, Cause: err}
	}
	budget := min(8, document.MaxAttempts)
	receipt.SearchConfig, _ = json.Marshal(struct {
		Ranker      string `json:"ranker"`
		Deduplicate bool   `json:"deduplicate_equal_descriptors"`
		Budget      int    `json:"effective_attempt_budget"`
	}{orderjudge.Schema, true, budget})
	receipt.SearchConfigSHA256 = digest(receipt.SearchConfig)
	receipt.Timing.ExecutionModel = "source_bind_then_whole_candidate_rank_then_finite_tdd_then_native_emit"
	receipt.Timing.DecisionStage = "one_local_prediction_before_candidate_tests_and_final_native_emission"
	if receipt.Observation != nil {
		receipt.Timing.ExecutionModel = "source_bind_then_oracle_observations_then_whole_candidate_rank_finite_tdd_then_native_emit"
		receipt.Timing.DecisionStage = "one_local_prediction_after_recorded_oracle_observations_before_candidate_tests; original_model_input_preserved"
	}
	searchStarted := time.Now()
	receipt.SearchStarted = true
	acquireStarted := time.Now()
	candidates, reused, err := model.acquire(ctx, document.Plan, prepared.PlanSHA256())
	if err != nil {
		return fail(fmt.Errorf("prepare whole-candidate search: %w", err))
	}
	receipt.OrderPreparation = &OrderPreparationReceipt{Schema: "gooo/prepared-order-candidates/v1",
		PlanSHA256: prepared.PlanSHA256(), Reused: reused, CandidateCount: 8, AcquireMS: elapsedMS(acquireStarted)}
	search, selected, ranking, err := candidates.Search(ctx, prepared.PlanSHA256(), document.TestCases, budget, true)
	receipt.Search, receipt.OrderJudgment = search, ranking
	receipt.Timing.BoundedSearchMS = elapsedMS(searchStarted)
	if err != nil {
		return fail(fmt.Errorf("whole-candidate search: %w", err))
	}
	if err := diagnoseSelectedPath(ctx, prepared, document, diagnosis, receipt); err != nil {
		return fail(err)
	}
	return emitSelectedTypedPath(ctx, filename, source, activity, bound, selected, document.TestCases, receipt, started)
}
