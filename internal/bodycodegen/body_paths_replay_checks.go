package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	"reflect"
)

func validateTypedPathReplayInput(ctx context.Context, source []byte, prior Result) error {
	if ctx == nil || len(source) == 0 || len(source) > 128<<10 || len(prior.Source) > 256<<10 ||
		len(prior.GoooSource) > 128<<10 {
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
	if p.SourceFormat != "" && p.SourceFormat != sourceAssemblyCheckpointFormat {
		return fmt.Errorf("unknown typed-path source format %q", p.SourceFormat)
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
	return nil
}

func compareTypedPathReplayProjection(prior, replayed Result, activityID string) error {
	r := prior.Report
	p := r.BodyPaths
	if replayed.Source != prior.Source || replayed.Report.SourceDigest != r.SourceDigest ||
		r.SourceDigest != p.SelectedSourceSHA256 || replayed.Report.GeneratedDigest != r.GeneratedDigest ||
		r.GeneratedDigest != r.ReplayDigest || replayed.Report.ActivityID != r.ActivityID ||
		r.ActivityID != activityID || replayed.Report.ProgramDigest != r.ProgramDigest ||
		!EqualRetainedConditionOrigins(r.BodyConditions, replayed.Report.BodyConditions) ||
		!EqualRetainedReadOrigins(r.BodyReads, replayed.Report.BodyReads) ||
		!EqualRetainedConstructionOrigins(r.BodyConstructions, replayed.Report.BodyConstructions) ||
		!EqualRetainedLocalOrigins(r.BodyLocals, replayed.Report.BodyLocals) ||
		!EqualRetainedFieldOrigins(r.BodyFields, replayed.Report.BodyFields) ||
		!EqualRetainedReturnOrigins(r.BodyReturns, replayed.Report.BodyReturns) {
		return fmt.Errorf("selected Gooo and generated projection do not replay exactly")
	}
	return nil
}

func verifyTypedPathFiniteObservations(ctx context.Context, source, completed []byte, replayed Result, bound typedPathSource, effectiveCases []pathplan.TestCase, p *BodyPathReceipt) error {
	r := replayed.Report
	cases := make([]IRBodyFillTestCase, len(effectiveCases))
	for i, c := range effectiveCases {
		cases[i] = IRBodyFillTestCase{Input: c.Input, Expected: c.Expected}
	}
	results, passed, err := evaluateIntegerCasesContext(ctx, []byte(replayed.Source), r.Activity, cases)
	if err != nil {
		return err
	}
	selectedPassed, selectedTotal := bodyPathSelectedScore(p)
	if !equalIRBodyFillCaseResults(results, p.NativeCases) || passed != selectedPassed || len(results) != selectedTotal {
		return fmt.Errorf("selected-body finite observations do not replay")
	}
	var holdout *PathHoldoutReceipt
	if bound.activity.Assembly != nil {
		holdout, err = evaluateSelectedPathHoldout(ctx, source, completed, &bound.activity.Assembly.Spec, replayed)
	}
	if err != nil || !reflect.DeepEqual(holdout, p.Holdout) {
		return fmt.Errorf("selected-body holdout observations do not replay: %v", err)
	}
	return nil
}
