package fullsoundness

import (
	"bytes"
	"reflect"
	"testing"
)

func TestSharedObligationClosure(t *testing.T) {
	input := soundInput()
	shared := id("obligation/impact")
	input.Commands = append(input.Commands, Command{ID: id("command/shared"), ObligationIDs: []string{shared}})
	input.FullOutcomes = append(input.FullOutcomes, outcome("command/shared", OutcomePass, "", "4"))
	input.FullResourceReceipts = append(input.FullResourceReceipts, receipt(&input, "command/shared", 1, 1, 1, 1, 1, 1))
	got := Evaluate(input)
	if got.Decision != DecisionUnsound || got.Reason != ReasonImpactedCommandOmitted {
		t.Fatalf("got %s/%s, want UNSOUND/IMPACTED_COMMAND_OMITTED", got.Decision, got.Reason)
	}
	selectOnly(&input, []string{id("command/guard"), id("command/impact"), id("command/shared")})
	got = Evaluate(input)
	if got.Decision != DecisionSound || got.Reason != ReasonSound {
		t.Fatalf("shared closure got %s/%s, want SOUND/SOUND", got.Decision, got.Reason)
	}
}
func TestPermutationCanonicalOutput(t *testing.T) {
	first := soundInput()
	second := soundInput()
	reverseInput(&second)
	leftOutput := Evaluate(first)
	rightOutput := Evaluate(second)
	if !reflect.DeepEqual(leftOutput, rightOutput) {
		t.Fatalf("permuted outputs differ:\n%#v\n%#v", leftOutput, rightOutput)
	}
	left, err := EncodeJSON(leftOutput)
	if err != nil {
		t.Fatal(err)
	}
	right, err := EncodeJSON(rightOutput)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		t.Fatalf("permutations differ:\n%s\n%s", left, right)
	}
}
func TestSnapshotChangesOnlyEnvelopeDigest(t *testing.T) {
	first := soundInput()
	second := soundInput()
	rebindSnapshot(&second, digest("0"))
	left := Evaluate(first)
	right := Evaluate(second)
	if left.Decision != DecisionSound || right.Decision != DecisionSound {
		t.Fatalf("snapshot inputs not sound: %s/%s", left.Decision, right.Decision)
	}
	if left.DecisionDigest != right.DecisionDigest {
		t.Fatalf("decision digest changed: %s != %s", left.DecisionDigest, right.DecisionDigest)
	}
	if left.CanonicalDigest == right.CanonicalDigest {
		t.Fatal("envelope digest did not bind snapshot")
	}
}
