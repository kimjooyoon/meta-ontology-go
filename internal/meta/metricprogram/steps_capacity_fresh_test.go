package metricprogram

import (
	"reflect"
	"testing"
)

func TestSelectionAllocationFresh1693(t *testing.T) {
	for _, tc := range []struct {
		name, selected string
		source, want   []string
		wantError      string
	}{
		{"empty-source", "replay-counterfactual", nil, []string{"replay-counterfactual"}, ""},
		{"selected-present", "replay-counterfactual", []string{"replay-counterfactual"},
			[]string{"replay-counterfactual"}, ""},
		{"append-selected", "replay-counterfactual", []string{"bind-exact-source-metrics"},
			[]string{"bind-exact-source-metrics", "replay-counterfactual"}, ""},
		{"exclude-other-terminal", "terminate-at-fixed-point", []string{"preserve-non-promoting-terminal"},
			[]string{"terminate-at-fixed-point"}, ""},
		{"duplicate", "replay-counterfactual", []string{"replay-counterfactual", "replay-counterfactual"},
			nil, "selected operation \"replay-counterfactual\" is duplicated"},
		{"unknown", "unknown-operation", nil, nil, "selected operation \"unknown-operation\" is invalid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := append([]string(nil), tc.source...)
			selection := StrategySelection{MetaOperation: tc.selected, SourceMetaOperations: tc.source}
			steps, err := buildSteps(StrategyPlan{Selection: selection})
			if tc.wantError != "" {
				if err == nil || err.Error() != tc.wantError {
					t.Fatalf("expected %q, got %v", tc.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, step := range steps {
				got = append(got, step.OperationID)
			}
			if !reflect.DeepEqual(got, tc.want) || !reflect.DeepEqual(tc.source, before) {
				t.Fatalf("selection was truncated, changed, or reordered: %v", got)
			}
		})
	}
}
