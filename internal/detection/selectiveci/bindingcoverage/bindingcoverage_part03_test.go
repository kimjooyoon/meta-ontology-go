package bindingcoverage

import (
	"strings"
	"testing"
)

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
