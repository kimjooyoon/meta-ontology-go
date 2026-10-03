package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/orderjudge"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func loadOrderJudge(name string, metadata []byte) (*orderjudge.Model, error) {
	if len(metadata) > 4096 {
		return nil, fmt.Errorf("whole-candidate metadata exceeds 4096 bytes")
	}
	f, err := os.Open(filepath.Join(filepath.Dir(name), "weights.bin"))
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
	return orderjudge.Load(metadata, weights)
}

// The caller has already bound the complete original plan to the source. The
// SDK sees original intent and actual operations, with no old context wrapper.
func generateOrderTypedPath(ctx context.Context, filename string, source []byte, activity string,
	document pathplan.Document, prepared *pathplan.PreparedPlan, bound typedPathSource,
	model *orderjudge.Model, diagnosis *PathDiagnosisOptions, receipt *BodyPathReceipt, started time.Time) (Result, error) {
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
	search, selected, ranking, err := orderjudge.Search(ctx, document.Plan, model, document.TestCases, budget, true)
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
