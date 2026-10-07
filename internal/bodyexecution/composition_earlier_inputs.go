package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// ConstructionInputStage binds an earlier fill's known input set to its source
// and contract. Actual tuples remain in the retained body-fill receipt.
type ConstructionInputStage struct {
	ActivityID     string `json:"activity_id"`
	SourceSHA256   string `json:"source_sha256"`
	PlanSHA256     string `json:"plan_sha256"`
	InputSetSHA256 string `json:"input_set_sha256"`
	UniqueInputs   int    `json:"unique_inputs"`
}

// MeasureEarlierFillInputs extends a fresh composition observation with fills
// generated or replayed by the calling workspace execution. It makes no model
// calls and does not re-attest a saved native execution supplied by a caller.
func MeasureEarlierFillInputs(ctx context.Context, filename string, source []byte, prior Composition,
	suite CompositionCases, runtime CompositionRuntime, fills []bodycodegen.Result) CompositionInputSeparation {
	unknown := unknownCompositionInputSeparation("EARLIER_CONSTRUCTION_INPUTS_UNAVAILABLE")
	if ctx == nil || ctx.Err() != nil || runtime.Stage != "COMPLETE" || !runtime.RuntimeReplayed ||
		!runtime.ProjectionReplayed || runtime.CompositionSHA256 != compositionDigest(prior) ||
		runtime.OriginalSourceSHA256 != digest(source) || runtime.RuntimeSuiteSHA256 != compositionDigest(suite) {
		return unknown
	}
	graph, err := prepareCompositionGraphForEntry(ctx, filename, source, prior.Plan.EntryActivity)
	if err != nil {
		return unknown
	}
	seen, err := graph.selectionInputs(ctx, filename, source, prior)
	if err != nil {
		return unknown
	}
	result := unknownCompositionInputSeparation("NO_DISJOINT_INPUTS")
	for _, fill := range fills {
		stage, err := graph.includeEarlierFill(prior, &seen, fill)
		if err != nil {
			return unknown
		}
		result.EarlierStages = append(result.EarlierStages, stage)
	}
	if ctx.Err() != nil {
		return unknown
	}
	result.Scope = "unique root-input tuples; actual inputs at all composing and earlier body-fill activities compared with their selection/holdout cases and probes; duplicate expectations must all match; model-training exposure is unknown"
	return graph.measureKnownInputSeparation(suite, runtime.Traces, seen, result)
}

func (graph *compositionGraph) includeEarlierFill(prior Composition, sets *compositionSelectionInputs,
	fill bodycodegen.Result) (ConstructionInputStage, error) {
	report := fill.Report
	for i, node := range graph.nodes[:graph.count] {
		if node.ID != report.ActivityID || node.Name != report.Activity {
			continue
		}
		if node.Assembling || report.ProgramDigest != prior.Steps[i].Generation.Report.ProgramDigest {
			return ConstructionInputStage{}, fmt.Errorf("earlier fill is repeated or differs from the executed body")
		}
		tuples, err := earlierFillTuples(report)
		if err != nil {
			return ConstructionInputStage{}, err
		}
		set := make(map[string]struct{}, len(tuples))
		for _, tuple := range tuples {
			key, err := graph.selectionInputKey(node, tuple)
			if err != nil {
				return ConstructionInputStage{}, err
			}
			set[key] = struct{}{}
		}
		keys := make([]string, 0, len(set))
		for key := range set {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		sets.nodes[i], graph.nodes[i].Assembling = set, true
		return ConstructionInputStage{ActivityID: node.ID, SourceSHA256: report.BodyFill.OriginalSourceDigest,
			PlanSHA256: report.BodyFill.IRPlanSHA256, InputSetSHA256: compositionDigest(keys), UniqueInputs: len(set)}, nil
	}
	return ConstructionInputStage{}, fmt.Errorf("earlier fill activity is outside the executed graph")
}

func earlierFillTuples(report bodycodegen.Report) ([][]json.RawMessage, error) {
	f := report.BodyFill
	if f == nil || report.Decision != "PASS" || !report.TypecheckPassed || f.TestCasesTotal < 1 ||
		f.OriginalSourceDigest == "" || f.IRPlanSHA256 == "" ||
		len(f.SelectedCaseResults)+len(f.SelectedValueCaseResults) != f.TestCasesTotal ||
		len(f.HoldoutCaseResults)+len(f.SelectedValueHoldoutCaseResults) != f.HoldoutCasesTotal {
		return nil, fmt.Errorf("earlier fill input observations are incomplete")
	}
	tuples := recordedInputTuples(report)
	for _, cases := range [][]bodycodegen.RecordAssemblyCase{f.SelectedValueCaseResults, f.SelectedValueHoldoutCaseResults} {
		for _, c := range cases {
			var tuple []json.RawMessage
			if err := json.Unmarshal(c.Inputs, &tuple); err != nil {
				return nil, err
			}
			tuples = append(tuples, tuple)
		}
	}
	return tuples, nil
}
