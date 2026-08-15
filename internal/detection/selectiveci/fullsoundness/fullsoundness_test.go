package fullsoundness

import (
	"bytes"
	"math"
	"strings"
	"testing"
)

func TestSoundFixture(t *testing.T) {
	input := soundInput()
	got := Evaluate(input)
	if got.Decision != DecisionSound || got.Reason != ReasonSound {
		t.Fatalf("got %s/%s, want SOUND/SOUND", got.Decision, got.Reason)
	}
	if got.CommandCount != 3 || got.SelectedCommandCount != 2 || got.ObligationCount != 2 || got.AuthoritativeImpactedObligationCount != 1 {
		t.Fatalf("semantic counts = %#v", got)
	}
	if got.ResourceVector == nil || got.ResourceVector.Class != ResourceImproved {
		t.Fatalf("resource vector = %#v, want IMPROVED", got.ResourceVector)
	}
	if got.ExecutionAuthorized || got.CIAuthorized || got.CanonicalDigest != got.StableDigest() {
		t.Fatalf("output flags or digest invalid: %#v", got)
	}
}

func TestClosedIDs(t *testing.T) {
	valid := []string{"c1", "o1", "a", "z_9", "command-guard"}
	for _, value := range valid {
		if !validID(value) {
			t.Errorf("validID(%q) = false", value)
		}
	}
	invalid := []string{"", strings.Repeat("a", 65), "C1", "c/1", "c:1", "c 1", "1command"}
	for _, value := range invalid {
		if validID(value) {
			t.Errorf("validID(%q) = true", value)
		}
	}
}

func TestClosedReasons(t *testing.T) {
	cases := []struct {
		name     string
		mutate   func(*Input)
		decision Decision
		reason   Reason
	}{
		{"authorization", func(input *Input) { input.ExecutionAuthorized = true }, DecisionUnsound, ReasonAuthorizationPresent},
		{"CI authorization", func(input *Input) { input.CIAuthorized = true }, DecisionUnsound, ReasonAuthorizationPresent},
		{"selected receipt set", func(input *Input) {
			input.SelectionReceipt.CommandIDs = append(input.SelectionReceipt.CommandIDs, id("command/pass"))
		}, DecisionUnsound, ReasonSelectedSetMismatch},
		{"global guard", func(input *Input) { selectOnly(input, []string{id("command/impact")}) }, DecisionUnsound, ReasonGlobalGuardOmitted},
		{"selected extra failure", func(input *Input) {
			selected := findOutcome(input.SelectedOutcomes, id("command/impact"))
			selected.Status = OutcomeFail
		}, DecisionUnsound, ReasonSelectedExtraFailure},
		{"status mismatch", func(input *Input) {
			full := findOutcome(input.FullOutcomes, id("command/impact"))
			full.Status, full.FailureCode = OutcomeFail, "full"
		}, DecisionUnsound, ReasonSelectedFullStatusMismatch},
		{"failure code", func(input *Input) {
			full := findOutcome(input.FullOutcomes, id("command/impact"))
			selected := findOutcome(input.SelectedOutcomes, id("command/impact"))
			full.Status, selected.Status, full.FailureCode, selected.FailureCode = OutcomeFail, OutcomeFail, "full", "selected"
		}, DecisionUnsound, ReasonFailureCodeMismatch},
		{"output digest", func(input *Input) {
			findOutcome(input.SelectedOutcomes, id("command/impact")).OutputDigest = digest("9")
		}, DecisionUnsound, ReasonOutputDigestMismatch},
		{"omitted full failure", func(input *Input) {
			selectOnly(input, []string{id("command/guard")})
			full := findOutcome(input.FullOutcomes, id("command/impact"))
			full.Status, full.FailureCode = OutcomeFail, "failed"
		}, DecisionUnsound, ReasonOmittedFullFailure},
		{"impacted omitted", func(input *Input) { selectOnly(input, []string{id("command/guard")}) }, DecisionUnsound, ReasonImpactedCommandOmitted},
		{"missing selected receipt", func(input *Input) { input.SelectedResourceReceipts = nil }, DecisionUnknown, ReasonFullSuiteRequired},
		{"missing full receipt", func(input *Input) { input.FullResourceReceipts = nil }, DecisionUnknown, ReasonFullSuiteRequired},
		{"missing selection receipt", func(input *Input) { input.SelectionReceipt = nil }, DecisionUnknown, ReasonFullSuiteRequired},
		{"duplicate command obligation", func(input *Input) {
			input.Commands[1].ObligationIDs = append(input.Commands[1].ObligationIDs, input.Commands[1].ObligationIDs[0])
		}, DecisionUnknown, ReasonFullSuiteRequired},
		{"unregistered obligation", func(input *Input) {
			input.Commands[1].ObligationIDs = append(input.Commands[1].ObligationIDs, id("obligation/missing"))
		}, DecisionUnknown, ReasonUnregisteredObligation},
		{"unprovable obligation", func(input *Input) { input.Obligations[0].Authority = AuthorityCandidate }, DecisionUnknown, ReasonUnprovableObligation},
		{"zero commands", zeroCommandInput, DecisionUnknown, ReasonZeroCommandDenominator},
		{"invalid outcome", func(input *Input) { findOutcome(input.FullOutcomes, id("command/impact")).Status = "OTHER" }, DecisionUnknown, ReasonInvalidOutcome},
		{"digest binding", func(input *Input) { input.SelectionReceipt.SnapshotDigest = digest("0") }, DecisionUnknown, ReasonDigestBindingMismatch},
		{"resource overflow", func(input *Input) { input.FullResourceReceipts[0].CPUCoreNS = math.MaxInt64 }, DecisionUnknown, ReasonResourceOverflow},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := soundInput()
			test.mutate(&input)
			got := Evaluate(input)
			if got.Decision != test.decision || got.Reason != test.reason {
				t.Fatalf("got %s/%s, want %s/%s", got.Decision, got.Reason, test.decision, test.reason)
			}
			if got.ExecutionAuthorized || got.CIAuthorized {
				t.Fatalf("result authorized execution: %#v", got)
			}
		})
	}
}

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
	left, err := EncodeJSON(Evaluate(first))
	if err != nil {
		t.Fatal(err)
	}
	right, err := EncodeJSON(Evaluate(second))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		t.Fatalf("permutations differ:\n%s\n%s", left, right)
	}
}

func TestResourceClasses(t *testing.T) {
	equal := soundInput()
	selectOnly(&equal, []string{id("command/guard"), id("command/impact"), id("command/pass")})
	if got := Evaluate(equal); got.Decision != DecisionSound || got.ResourceVector.Class != ResourceEqual {
		t.Fatalf("equal resources got %#v", got)
	}
	regressed := soundInput()
	findReceipt(regressed.SelectedResourceReceipts, id("command/impact")).CPUCoreNS = 20
	if got := Evaluate(regressed); got.Decision != DecisionSound || got.ResourceVector.Class != ResourceRegressed {
		t.Fatalf("regressed resources got %#v", got)
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
