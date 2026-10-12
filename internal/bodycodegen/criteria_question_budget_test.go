package bodycodegen

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestCriteriaQuestionsRespectCombinedInputBudget(t *testing.T) {
	source := strings.Replace(criteriaQuestionFixture(), "return input1.value == 0-input0", "return true", 1)
	source = strings.ReplaceAll(source, `holdout_input "[9]"`, `holdout_input "[999]"`)
	source = strings.ReplaceAll(source, `criteria_input "[7]"`, `criteria_input "[1001]"`)
	source = strings.ReplaceAll(source, `criteria_input "[-5]"`, `criteria_input "[-1002]"`)
	var initial strings.Builder
	for i := range 127 {
		fmt.Fprintf(&initial, " input \"[%d]\"\n", i)
	}
	source = strings.Replace(source, `input "[0]"`, initial.String(), 1)
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(context.Background(), "budget.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal(err)
	}
	o := result.Report.RecordAssembly.CriteriaQuestions
	if o.Status != "CASE_BUDGET" || o.ResolvedMask != nil || len(o.AcquiredInputs) != 0 || o.EffectiveInputs != 127 || o.Rounds[0].RecommendedIndex == nil {
		t.Fatal("question silently exceeded the shared input budget", o)
	}
	if result.Report.RecordAssembly.Criteria.Total != 127 {
		t.Fatal("unasked input became an obligation")
	}
}
