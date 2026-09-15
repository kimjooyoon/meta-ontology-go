package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"testing"
)

func TestR4SafePassAndV1Evidence(t *testing.T) {
	input := r4SafeInput(t, 10)
	raw, _ := workfrontier.EncodeR4JSON(input.R4Input)
	if baseline := workfrontier.FairBaseline(input.R4Input); !sameB1Values(baseline, []string{"path/root"}) {
		t.Fatalf("baseline selected %v", baseline)
	}
	got := ValidateR4Safe(input)
	if got.Decision != DecisionPass || got.Reason != ReasonNone || got.ExecutionAuthorized ||
		got.EnforcementEffect != EnforcementNoEffect || !sameB1Values(got.SafeSelectedIDs, []string{"path/root"}) ||
		!sameB1Values(got.SafeWorkIDs, []string{"work/root"}) || got.R4Result.Status != workfrontier.R4StatusPass {
		t.Fatalf("safe result = %#v", got)
	}
	if after, _ := workfrontier.EncodeR4JSON(input.R4Input); string(raw) != string(after) {
		t.Fatal("R4 v1 evidence changed")
	}
	relocated := r4SafeInput(t, 10)
	relocated.R4Input.RootObligationIDs = []string{"obligation/other"}
	if got := ValidateR4Safe(relocated); got.InputDigest == ValidateR4Safe(input).InputDigest {
		t.Fatal("root relocation did not change replay input")
	}
}

type r4SafeCase struct {
	name     string
	edit     func(*R4SafeInput)
	decision Decision
	reason   Reason
}

func runR4SafeCases(t *testing.T, cases []r4SafeCase) {
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := r4SafeInput(t, 10)
			test.edit(&input)
			got := ValidateR4Safe(input)
			if got.Decision != test.decision || got.Reason != test.reason ||
				got.SafeSelectedIDs != nil || got.SafeWorkIDs != nil {
				t.Fatalf("result = %#v", got)
			}
		})
	}
}
