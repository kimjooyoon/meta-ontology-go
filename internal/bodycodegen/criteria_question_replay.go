package bodycodegen

import (
	"context"
	"fmt"
)

func replayCriteriaQuestions(ctx context.Context, source []byte, plans []recordAssemblyPlan, history []RecordAssemblyControlStage,
	prior, expected *RecordAssemblyReceipt) error {
	if prior.Observation != nil {
		return fmt.Errorf("criteria questions cannot carry a record oracle observation")
	}
	original := source
	if len(history) > 0 && history[0].Source != "" {
		original = []byte(history[0].Source)
	}
	r := newRecordAssemblyReceipt(original, plans[0])
	if err := observeRecordCriteriaQuestions(ctx, original, &plans[0], r); err != nil {
		return err
	}
	if !equalLocalCallEvidence(r.CriteriaQuestions, prior.CriteriaQuestions) {
		return fmt.Errorf("criteria question history does not replay")
	}
	expected.CriteriaQuestions = r.CriteriaQuestions
	for i := 1; i < len(plans); i++ {
		plans[i].effectiveCriteriaInputs = append([]string(nil), plans[0].effectiveCriteriaInputs...)
	}
	return nil
}
