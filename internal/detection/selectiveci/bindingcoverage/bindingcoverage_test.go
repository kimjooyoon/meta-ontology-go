package bindingcoverage

import (
	"bytes"
	"math"
	"strings"
	"testing"
)

func TestExactFixture(t *testing.T) {
	input := fixtureInput()
	got := Observe(input)
	if got.Decision != DecisionExact || got.Reason != ReasonComplete {
		t.Fatalf("got %s/%s, want EXACT/COMPLETE", got.Decision, got.Reason)
	}
	if got.RequiredBindingCount != 9 || got.MatchCoveredCount != 9 || got.MismatchCoveredCount != 9 || got.PartitionCount != 18 {
		t.Fatalf("coverage counts = %d/%d/%d/%d, want 9/9/9/18", got.RequiredBindingCount, got.MatchCoveredCount, got.MismatchCoveredCount, got.PartitionCount)
	}
	if got.EndpointReferenceCount != 18 || got.DeterministicWorkUnits != 45 || got.InputBytes == 0 {
		t.Fatalf("work/input bytes = %d/%d, want work 45 and nonzero input", got.DeterministicWorkUnits, got.InputBytes)
	}
	if got.InputDigest == "" {
		t.Fatal("canonical input digest is empty")
	}
	if len(got.MissingMatchBindingIDs) != 0 || len(got.MissingMismatchBindingIDs) != 0 {
		t.Fatal("exact fixture reported missing coverage")
	}
	if got.CanonicalDigest != got.StableDigest() {
		t.Fatalf("digest = %q, stable digest = %q", got.CanonicalDigest, got.StableDigest())
	}
}

func TestIncompleteCoverage(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Input)
		reason Reason
		check  func(*testing.T, Output)
	}{
		{"omit lane registry mismatch", func(input *Input) {
			input.Partitions = withoutPartition(input.Partitions, bindingID("lane-registry"), PolarityMismatch)
		}, ReasonMissingMismatch, func(t *testing.T, got Output) {
			assertMissing(t, got.MissingMismatchBindingIDs, bindingID("lane-registry"))
			if len(got.MissingMatchBindingIDs) != 0 {
				t.Fatalf("unexpected missing MATCH IDs: %v", got.MissingMatchBindingIDs)
			}
		}},
		{"missing match", func(input *Input) {
			input.Partitions = withoutPartition(input.Partitions, bindingID("base-head"), PolarityMatch)
		}, ReasonMissingMatch, func(t *testing.T, got Output) {
			assertMissing(t, got.MissingMatchBindingIDs, bindingID("base-head"))
		}},
		{"zero denominator", func(input *Input) {
			input.RequiredBindings = []RequiredBinding{}
			input.Partitions = []Partition{}
		}, ReasonZeroDenominator, func(t *testing.T, got Output) {
			if got.RequiredBindingCount != 0 || got.PartitionCount != 0 || got.EndpointReferenceCount != 0 || got.DeterministicWorkUnits != 0 || len(got.MissingMatchBindingIDs) != 0 || len(got.MissingMismatchBindingIDs) != 0 {
				t.Fatalf("zero denominator output = %#v", got)
			}
		}},
		{"missing match and mismatch", func(input *Input) {
			input.Partitions = withoutPartition(input.Partitions, bindingID("base-head"), PolarityMatch)
			input.Partitions = withoutPartition(input.Partitions, bindingID("base-head"), PolarityMismatch)
		}, ReasonMissingMatchAndMismatch, func(t *testing.T, got Output) {
			assertMissing(t, got.MissingMatchBindingIDs, bindingID("base-head"))
			assertMissing(t, got.MissingMismatchBindingIDs, bindingID("base-head"))
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := fixtureInput()
			test.mutate(&input)
			got := Observe(input)
			if got.Decision != DecisionIncomplete || got.Reason != test.reason {
				t.Fatalf("got %s/%s, want INCOMPLETE/%s", got.Decision, got.Reason, test.reason)
			}
			test.check(t, got)
		})
	}
}

func TestMalformedAndAmbiguousCases(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Input)
		reason Reason
	}{
		{"duplicate binding ID", func(input *Input) {
			input.RequiredBindings = append(input.RequiredBindings, input.RequiredBindings[0])
		}, ReasonDuplicateID},
		{"duplicate partition ID", func(input *Input) {
			input.Partitions = append(input.Partitions, input.Partitions[0])
		}, ReasonDuplicateID},
		{"cross-kind duplicate ID", func(input *Input) {
			duplicate := input.Partitions[0]
			duplicate.PartitionID = input.RequiredBindings[0].BindingID
			input.Partitions[0] = duplicate
		}, ReasonDuplicateID},
		{"duplicate polarity", func(input *Input) {
			duplicate := input.Partitions[0]
			duplicate.PartitionID = id("partition/duplicate-polarity")
			input.Partitions = append(input.Partitions, duplicate)
		}, ReasonDuplicatePolarity},
		{"dangling binding", func(input *Input) {
			input.Partitions = append(input.Partitions, Partition{PartitionID: id("partition/dangling"), BindingID: id("binding/missing"), Polarity: PolarityMatch, ExpectedStage: "stage:dangling", ExpectedReason: "reason:dangling"})
		}, ReasonUnknownReference},
		{"malformed digest", func(input *Input) { input.SnapshotDigest = strings.Repeat("A", 64) }, ReasonInvalidDigest},
		{"snapshot mismatch", func(input *Input) { input.ExpectedSnapshotDigest = strings.Repeat("b", 64) }, ReasonSnapshotMismatch},
		{"malformed ID", func(input *Input) { input.ContractID = "not-a-stable-id" }, ReasonInvalidID},
		{"empty token", func(input *Input) { input.Partitions[0].ExpectedStage = "" }, ReasonMissingInput},
		{"wrong prefix token", func(input *Input) { input.Partitions[0].ExpectedStage = "phase:wrong" }, ReasonInvalidToken},
		{"invalid token", func(input *Input) { input.Partitions[0].ExpectedReason = "reason:has spaces" }, ReasonInvalidToken},
		{"invalid enum", func(input *Input) { input.RequiredBindings[0].Kind = "UNKNOWN" }, ReasonInvalidEnum},
		{"malformed precedence token", func(input *Input) { input.PrecedenceRegistry[0].Stage = "phase:wrong" }, ReasonInvalidPrecedence},
		{"duplicate precedence rank", func(input *Input) { input.PrecedenceRegistry[1].Rank = input.PrecedenceRegistry[0].Rank }, ReasonDuplicatePrecedence},
		{"duplicate precedence pair", func(input *Input) {
			input.PrecedenceRegistry = append(input.PrecedenceRegistry, PrecedenceEntry{Rank: 10, Stage: input.PrecedenceRegistry[0].Stage, Reason: input.PrecedenceRegistry[0].Reason})
		}, ReasonDuplicatePrecedence},
		{"unregistered binding pair", func(input *Input) { input.RequiredBindings[0].ExpectedReason = "reason:unregistered" }, ReasonUnregisteredPair},
		{"stale partition pair", func(input *Input) { input.Partitions[0].ExpectedStage = "stage:wrong" }, ReasonStalePartition},
		{"self-link", func(input *Input) { input.RequiredBindings[0].ToFieldID = input.RequiredBindings[0].FromFieldID }, ReasonSelfLink},
		{"unsupported schema", func(input *Input) { input.SchemaVersion = "gooo/other/v1" }, ReasonUnknownSchema},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := fixtureInput()
			test.mutate(&input)
			got := Observe(input)
			if got.Decision != DecisionUnknown || got.Reason != test.reason {
				t.Fatalf("got %s/%s, want UNKNOWN/%s", got.Decision, got.Reason, test.reason)
			}
			assertShapeCounts(t, got, input)
			if len(got.MissingMatchBindingIDs) != 0 || len(got.MissingMismatchBindingIDs) != 0 {
				t.Fatalf("UNKNOWN output reported coverage gaps: %#v", got)
			}
		})
	}
}

func TestPermutationCanonicalEquality(t *testing.T) {
	first := fixtureInput()
	second := fixtureInput()
	second.RequiredBindings = reverseBindings(second.RequiredBindings)
	second.Partitions = reversePartitions(second.Partitions)
	second.PrecedenceRegistry = reversePrecedence(second.PrecedenceRegistry)
	left, err := EncodeJSON(Observe(first))
	if err != nil {
		t.Fatal(err)
	}
	right, err := EncodeJSON(Observe(second))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		t.Fatalf("permutations differ:\n%s\n%s", left, right)
	}
}

func TestZeroPrecedenceRankIsValid(t *testing.T) {
	input := fixtureInput()
	input.PrecedenceRegistry[0].Rank = 0
	got := Observe(input)
	if got.Decision != DecisionExact || got.Reason != ReasonComplete {
		t.Fatalf("got %s/%s, want EXACT/COMPLETE", got.Decision, got.Reason)
	}
}

func TestStrictJSON(t *testing.T) {
	encoded, err := EncodeInputJSON(fixtureInput())
	if err != nil {
		t.Fatal(err)
	}
	if got := ClassifyJSON(encoded); got.Decision != DecisionExact {
		t.Fatalf("encoded fixture got %s/%s", got.Decision, got.Reason)
	}
	canonical := strings.TrimSpace(string(encoded))
	duplicate := strings.Replace(canonical, `"schema_version":"`+SchemaVersion+`"`, `"schema_version":"`+SchemaVersion+`","schema_version":"`+SchemaVersion+`"`, 1)
	unknown := strings.TrimSuffix(canonical, "}") + `,"extra":true}`
	trailing := canonical + " {}"
	for name, data := range map[string]string{"duplicate": duplicate, "unknown": unknown, "trailing": trailing} {
		t.Run(name, func(t *testing.T) {
			if got := ClassifyJSON([]byte(data)); got.Decision != DecisionUnknown || got.Reason != ReasonMissingInput {
				t.Fatalf("got %s/%s", got.Decision, got.Reason)
			}
		})
	}
	missingLists := fixtureInput()
	missingLists.RequiredBindings = nil
	if got := Observe(missingLists); got.Decision != DecisionUnknown || got.Reason != ReasonMissingInput {
		t.Fatalf("nil list got %s/%s", got.Decision, got.Reason)
	}
	missingLists = fixtureInput()
	missingLists.Partitions = nil
	if got := Observe(missingLists); got.Decision != DecisionUnknown || got.Reason != ReasonMissingInput {
		t.Fatalf("nil partitions got %s/%s", got.Decision, got.Reason)
	}
	missingLists = fixtureInput()
	missingLists.PrecedenceRegistry = nil
	if got := Observe(missingLists); got.Decision != DecisionUnknown || got.Reason != ReasonMissingInput {
		t.Fatalf("nil precedence got %s/%s", got.Decision, got.Reason)
	}
	if _, err := DecodeJSON([]byte(unknown)); err == nil {
		t.Fatal("strict decoder accepted unknown field")
	}
	if got := ClassifyJSON([]byte(`{"schema_version":"gooo/other/v1"}`)); got.Decision != DecisionUnknown || got.Reason != ReasonUnknownSchema {
		t.Fatalf("unsupported JSON schema got %s/%s", got.Decision, got.Reason)
	}
}

func TestWorkAccountingOverflow(t *testing.T) {
	if _, ok := addUint64(math.MaxUint64, 1); ok {
		t.Fatal("addition overflow was accepted")
	}
	if _, ok := workUnits(math.MaxUint64, 0, 1); ok {
		t.Fatal("work overflow was accepted")
	}
	if got, ok := workUnits(9, 18, 18); !ok || got != 45 {
		t.Fatalf("work units = %d/%v, want 45/true", got, ok)
	}
}

func TestSharedEndpointReferences(t *testing.T) {
	got := Observe(sharedEndpointInput())
	if got.Decision != DecisionExact || got.Reason != ReasonComplete {
		t.Fatalf("got %s/%s, want EXACT/COMPLETE", got.Decision, got.Reason)
	}
	if got.RequiredBindingCount != 2 || got.MatchCoveredCount != 2 || got.MismatchCoveredCount != 2 || got.PartitionCount != 4 {
		t.Fatalf("coverage counts = %d/%d/%d/%d, want 2/2/2/4", got.RequiredBindingCount, got.MatchCoveredCount, got.MismatchCoveredCount, got.PartitionCount)
	}
	if got.EndpointReferenceCount != 4 || got.DeterministicWorkUnits != 10 {
		t.Fatalf("endpoint/work counts = %d/%d, want 4/10", got.EndpointReferenceCount, got.DeterministicWorkUnits)
	}
	t.Logf("shared endpoint fixture digest=%s", got.CanonicalDigest)
}

func TestCanonicalFixtureDigest(t *testing.T) {
	got := Observe(fixtureInput())
	t.Logf("binding coverage fixture digest=%s counts=%d/%d/%d/%d work=%d input_bytes=%d", got.CanonicalDigest, got.RequiredBindingCount, got.MatchCoveredCount, got.MismatchCoveredCount, got.PartitionCount, got.DeterministicWorkUnits, got.InputBytes)
	if got.CanonicalDigest != got.StableDigest() {
		t.Fatal("fixture digest changed after canonicalization")
	}
}

func assertMissing(t *testing.T, got []string, want string) {
	t.Helper()
	if len(got) != 1 || got[0] != want {
		t.Fatalf("missing IDs = %v, want [%s]", got, want)
	}
}
