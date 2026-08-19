package fullsoundness

import (
	"reflect"
	"strings"
	"testing"
)

func assertDecisionSemanticStage(t *testing.T, output Output) {
	t.Helper()
	semantic, err := semanticProjection(output)
	if err != nil {
		t.Fatal(err)
	}
	if !output.SemanticEvaluated {
		if semantic.FullCount != 0 || semantic.SelectedCount != 0 || semantic.FullPassCount != 0 || semantic.SelectedPassCount != 0 || semantic.FullFailCount != 0 || semantic.SelectedFailCount != 0 || len(semantic.FullFailureIDs) != 0 || len(semantic.SelectedFailureIDs) != 0 || len(semantic.OmittedIDs) != 0 {
			t.Fatalf("unevaluated semantic projection = %#v", semantic)
		}
		return
	}
	if semantic.FullCount != 3 || semantic.SelectedCount != 2 || semantic.FullPassCount != 3 || semantic.SelectedPassCount != 2 || len(semantic.FullFailureIDs) != 0 || len(semantic.SelectedFailureIDs) != 0 || !reflect.DeepEqual(semantic.OmittedIDs, []string{id("command/pass")}) {
		t.Fatalf("evaluated semantic projection = %#v", semantic)
	}
}
func TestStrictJSON(t *testing.T) {
	encoded, err := EncodeInputJSON(soundInput())
	if err != nil {
		t.Fatal(err)
	}
	if got := ClassifyJSON(encoded); got.Decision != DecisionSound {
		t.Fatalf("encoded input got %s/%s", got.Decision, got.Reason)
	}
	canonical := strings.TrimSpace(string(encoded))
	duplicate := strings.Replace(canonical, `"schema_version":"`+SchemaVersion+`"`, `"schema_version":"`+SchemaVersion+`","schema_version":"`+SchemaVersion+`"`, 1)
	unknown := strings.TrimSuffix(canonical, "}") + `,"extra":true}`
	for name, data := range map[string]string{"duplicate": duplicate, "unknown": unknown, "trailing": canonical + " {}"} {
		t.Run(name, func(t *testing.T) {
			got := ClassifyJSON([]byte(data))
			if got.Decision != DecisionUnknown || got.Reason != ReasonFullSuiteRequired {
				t.Fatalf("got %s/%s, want UNKNOWN/FULL_SUITE_REQUIRED", got.Decision, got.Reason)
			}
		})
	}
}
func zeroCommandInput(input *Input) {
	input.Obligations = []Obligation{}
	input.Commands = []Command{}
	input.ImpactedObligationIDs = []string{}
	input.SelectedCommandIDs = []string{}
	input.SelectionReceipt.CommandIDs = []string{}
	input.FullOutcomes = []Outcome{}
	input.SelectedOutcomes = []Outcome{}
	input.FullResourceReceipts = []ResourceReceipt{}
	input.SelectedResourceReceipts = []ResourceReceipt{}
}
func reverseInput(input *Input) {
	reverseObligations(input.Obligations)
	reverseCommands(input.Commands)
	reverseOutcomes(input.FullOutcomes)
	reverseOutcomes(input.SelectedOutcomes)
	reverseReceipts(input.FullResourceReceipts)
	reverseReceipts(input.SelectedResourceReceipts)
	reverseStrings(input.ImpactedObligationIDs)
	reverseStrings(input.SelectedCommandIDs)
	reverseStrings(input.SelectionReceipt.CommandIDs)
}
