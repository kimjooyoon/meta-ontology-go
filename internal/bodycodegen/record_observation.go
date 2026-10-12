package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func recordSelectionCases(p recordAssemblyPlan) []assemblyspec.ValueCase {
	if p.effectiveCases != nil {
		return p.effectiveCases
	}
	return p.spec.ValueCases
}

func recordCasesSHA(cases []assemblyspec.ValueCase) string {
	raw, _ := json.Marshal(cases)
	return digest(raw)
}

func observeRecordAssembly(ctx context.Context, filename string, source []byte, p *recordAssemblyPlan,
	receipt *RecordAssemblyReceipt) (failure error) {
	if p.spec.Observation == nil {
		return nil
	}
	o := *p.spec.Clone().Observation
	raw, _ := json.Marshal(o)
	r := &RecordObservationReceipt{Schema: "gooo/record-observation/v1", Status: "STARTED",
		SourceSHA256: digest(source), ContractSHA256: receipt.ContractSHA256, Options: o, OptionsSHA256: digest(raw),
		InitialCases: append([]assemblyspec.ValueCase(nil), p.spec.ValueCases...),
		Scope:        "source-declared questions and fixed pure oracle; bounded AST evaluations; held cases and model inputs excluded"}
	receipt.Observation = r
	p.effectiveCases = append([]assemblyspec.ValueCase(nil), p.spec.ValueCases...)
	defer func() {
		r.EffectiveCases, r.EffectiveCasesSHA256 = len(p.effectiveCases), recordCasesSHA(p.effectiveCases)
		if failure != nil {
			r.Status = "FAILED"
		}
	}()
	oracle, err := prepareRecordOracle(ctx, filename, source, *p, p.effectiveCases, r)
	if err != nil {
		return err
	}
	ranker := recordObservationRanker{plan: *p, receipt: r, outputs: make(map[string]RecordProbeOutput), candidates: make(map[uint16]generatedRoute)}
	for round := 0; ; round++ {
		ranked, err := ranker.rank(ctx, p.effectiveCases)
		r.Rounds = append(r.Rounds, ranked)
		if err != nil {
			return err
		}
		if ranked.RecommendedIndex == nil {
			r.Status = recordObservationStop(ranked)
			if o.ResolveUniqueCandidate && r.Status == "ONE_SURVIVING_CANDIDATE" {
				r.ResolvedMask = new(ranked.SurvivingMasks[0])
			}
			return nil
		}
		if o.OracleActivity == "" {
			r.Status = "ORACLE_UNAVAILABLE"
			return nil
		}
		if round == o.MaxRounds {
			r.Status = "ROUND_BUDGET"
			return nil
		}
		if len(p.effectiveCases)+len(p.spec.ValueHoldoutCases) >= 128 {
			r.Status = "CASE_BUDGET"
			return nil
		}
		input := ranked.Probes[*ranked.RecommendedIndex].Input
		probe := assemblyspec.ValueCase{Inputs: input, Expected: p.spec.ValueCases[0].Expected}
		values, count, err := evaluateRecordOracle(ctx, oracle, []assemblyspec.ValueCase{probe})
		r.OracleEvaluations += count
		if err != nil {
			return fmt.Errorf("observe declared record oracle: %w", err)
		}
		probe.Expected = string(values[0].Actual)
		r.Rounds[len(r.Rounds)-1].Observation = &probe
		p.effectiveCases = append(p.effectiveCases, probe)
	}
}

func recordObservationStop(r RecordObservationRound) string {
	if r.Unobserved > 0 {
		return "CANDIDATE_BUDGET"
	}
	if len(r.UnavailableMasks) > 0 {
		return "EVALUATION_UNAVAILABLE"
	}
	if len(r.SurvivingMasks) == 0 {
		return "NO_SURVIVING_CANDIDATE"
	}
	if len(r.SurvivingMasks) == 1 {
		return "ONE_SURVIVING_CANDIDATE"
	}
	return "NO_DISTINGUISHING_INPUT"
}

func prepareRecordOracle(ctx context.Context, filename string, source []byte, p recordAssemblyPlan,
	cases []assemblyspec.ValueCase, r *RecordObservationReceipt) (Result, error) {
	name := r.Options.OracleActivity
	if name == "" {
		return Result{}, nil
	}
	if name == p.body.activity.Name {
		return Result{}, fmt.Errorf("record oracle must name a separate fixed computes activity")
	}
	spec, err := SourceAssembly(ctx, filename, source, name)
	if err != nil {
		return Result{}, err
	}
	if spec != nil {
		return Result{}, fmt.Errorf("record oracle requires fixed computes")
	}
	declared, err := prepareActivityBodySource(filename, source, name)
	if err != nil {
		return Result{}, err
	}
	if declared.outputType != p.body.outputType || !reflect.DeepEqual(declared.parameters, p.body.parameters) {
		return Result{}, fmt.Errorf("record oracle requires the same typed positional signature")
	}
	oracle, err := GenerateWithPlanner(ctx, filename, source, name, "", "")
	if err != nil {
		return Result{}, fmt.Errorf("prepare record oracle: %w", err)
	}
	r.OracleActivityID, r.OracleGoSHA256 = oracle.Report.ActivityID, digest([]byte(oracle.Source))
	r.OracleInitialCases, r.OracleEvaluations, err = evaluateRecordOracle(ctx, oracle, cases)
	if err != nil {
		return Result{}, err
	}
	if scoreRecordCases(0, r.OracleInitialCases).Passed != len(cases) {
		return Result{}, fmt.Errorf("declared record oracle contradicts retained finite cases")
	}
	return oracle, nil
}

func evaluateRecordOracle(ctx context.Context, oracle Result, cases []assemblyspec.ValueCase) ([]RecordAssemblyCase, int, error) {
	recorder := newLocalCallRecorder([]byte(oracle.Source), oracle.Report, cases)
	rows, err := evaluateRecordAssemblyRecorded(ctx, []byte(oracle.Source), oracle.Report.Activity, oracle.Report.RecordTypes, cases, recorder)
	return rows, recorder.observation.CasesStarted, err
}

func resolveRecordObservationRanking(r *RecordAssemblyReceipt) bool {
	if r.Observation == nil || r.Observation.ResolvedMask == nil {
		return false
	}
	mask := *r.Observation.ResolvedMask
	r.Ranking = []uint16{mask}
	for i := range 1 << len(r.Choices) {
		if uint16(i) != mask {
			r.Ranking = append(r.Ranking, uint16(i))
		}
	}
	return true
}
