package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

// VerifyTypedPathProjection reconstructs the declared selection from the original
// source and document without loading a model. When an observation loop was used,
// its bounded candidate observations and oracle labels are independently replayed.
// This validates the emitted pure program; it does not attest earlier model execution.
func VerifyTypedPathProjection(ctx context.Context, filename string, source []byte,
	document pathplan.Document, prior Result) error {
	if ctx == nil || len(source) == 0 || len(source) > 128<<10 || len(prior.Source) > 256<<10 {
		return fmt.Errorf("typed-path replay requires bounded source and context")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	r, p := prior.Report, prior.Report.BodyPaths
	if r.Schema != schema || r.Decision != "PASS" || p == nil || r.BodyFill != nil || r.BodySearch != nil ||
		p.OriginalSourceSHA256 != digest(source) || !r.TypecheckPassed || !r.DeterministicReplay {
		return fmt.Errorf("typed-path generation observation is missing or unbound")
	}
	encoded, err := json.Marshal(r.CompletenessReceipt)
	if err != nil {
		return err
	}
	common, err := completeness.Decode(encoded)
	if err != nil || common.ProfileID != typedPathCompletenessProfile || common.Decision == "FAIL_CLOSED" {
		return fmt.Errorf("typed-path shared receipt is invalid: %v", err)
	}
	if r.PlanSHA256 != completenessPlanSHA(r) || common.Scope["plan_sha256"] != r.PlanSHA256 ||
		common.Scope["compiler_source_sha"] != r.CompilerSourceSHA {
		return fmt.Errorf("typed-path plan identity differs")
	}
	pathScope, ok := common.Scope["typed_path"].(map[string]any)
	pathBytes, err := json.Marshal(p)
	if !ok || err != nil || pathScope["observation_sha256"] != digest(pathBytes) {
		return fmt.Errorf("typed-path common receipt does not bind its observation")
	}
	prepared, err := document.Prepare()
	if err != nil {
		return err
	}
	observed := &BodyPathReceipt{OriginalSourceSHA256: digest(source)}
	if err := bindTypedPathDocument(document, observed); err != nil {
		return err
	}
	if observed.DocumentSHA256 != p.DocumentSHA256 || observed.TestSuiteSHA256 != p.TestSuiteSHA256 ||
		observed.DeclaredTestCases != p.DeclaredTestCases || len(p.Search.Selection.Choices) != len(document.Plan.Decisions) {
		return fmt.Errorf("typed-path document, finite suite or complete selection differs")
	}
	bound, err := bindTypedPathSource(ctx, filename, source, r.Activity, prepared, observed)
	if err != nil {
		return err
	}
	effectiveCases, err := replayPathObservation(ctx, filename, source, r.Activity, prepared, document.TestCases, p)
	if err != nil {
		return err
	}
	selected, err := prepared.Compile(p.Search.Selection.Choices)
	if err != nil {
		return err
	}
	completed, err := replaceActivityProgram(source, bound.activity.ValueProgramSpan, selected.GoooBody())
	if err != nil {
		return err
	}
	replayed, err := GenerateWithPlanner(ctx, filename, completed, r.Activity, "", "")
	if err != nil {
		return err
	}
	if replayed.Source != prior.Source || replayed.Report.SourceDigest != r.SourceDigest ||
		r.SourceDigest != p.SelectedSourceSHA256 || replayed.Report.GeneratedDigest != r.GeneratedDigest ||
		r.GeneratedDigest != r.ReplayDigest || replayed.Report.ActivityID != r.ActivityID ||
		r.ActivityID != bound.base.Report.ActivityID || replayed.Report.ProgramDigest != r.ProgramDigest {
		return fmt.Errorf("selected Gooo and generated projection do not replay exactly")
	}
	cases := make([]IRBodyFillTestCase, len(effectiveCases))
	for i, c := range effectiveCases {
		cases[i] = IRBodyFillTestCase{Input: c.Input, Expected: c.Expected}
	}
	results, passed, err := evaluateIntegerCasesContext(ctx, []byte(replayed.Source), r.Activity, cases)
	if err != nil {
		return err
	}
	if !slices.Equal(results, p.NativeCases) || passed != p.Search.SelectedTrainingPassed ||
		len(results) != p.Search.TrainingTotal {
		return fmt.Errorf("selected-body finite observations do not replay")
	}
	return nil
}
