package fullsoundness

import (
	"strings"
)

func soundInput() Input {
	input := Input{
		SchemaVersion: SchemaVersion, SnapshotDigest: digest("a"), PolicyDigest: digest("b"), RegistryDigest: digest("c"), SelectionDigest: digest("d"), ToolchainDigest: digest("e"), RunnerDigest: digest("f"),
		Obligations: []Obligation{{ID: id("obligation/impact"), Authority: AuthorityAuthoritative}, {ID: id("obligation/pass"), Authority: AuthorityAuthoritative}},
		Commands: []Command{
			{ID: id("command/guard"), ObligationIDs: []string{}, GlobalGuard: true},
			{ID: id("command/impact"), ObligationIDs: []string{id("obligation/impact")}},
			{ID: id("command/pass"), ObligationIDs: []string{id("obligation/pass")}},
		},
		ImpactedObligationIDs: []string{id("obligation/impact")},
	}
	setSelected(&input, []string{id("command/guard"), id("command/impact")})
	input.FullOutcomes = []Outcome{outcome("command/guard", OutcomePass, "", "1"), outcome("command/impact", OutcomePass, "", "2"), outcome("command/pass", OutcomePass, "", "3")}
	input.SelectedOutcomes = []Outcome{outcome("command/guard", OutcomePass, "", "1"), outcome("command/impact", OutcomePass, "", "2")}
	input.FullResourceReceipts = []ResourceReceipt{receipt(&input, "command/guard", 2, 1, 2, 4, 2, 1), receipt(&input, "command/impact", 3, 1, 3, 6, 3, 1), receipt(&input, "command/pass", 5, 1, 5, 5, 5, 2)}
	input.SelectedResourceReceipts = []ResourceReceipt{receipt(&input, "command/guard", 2, 1, 2, 4, 2, 1), receipt(&input, "command/impact", 3, 1, 3, 6, 3, 1)}
	return input
}
func setSelected(input *Input, ids []string) {
	input.SelectedCommandIDs = append([]string(nil), ids...)
	input.SelectionReceipt = &SelectionReceipt{SnapshotDigest: input.SnapshotDigest, PolicyDigest: input.PolicyDigest, RegistryDigest: input.RegistryDigest, SelectionDigest: input.SelectionDigest, CommandIDs: append([]string(nil), ids...)}
}
func outcome(name string, status OutcomeStatus, failureCode, marker string) Outcome {
	return Outcome{CommandID: id(name), Status: status, FailureCode: failureCode, OutputDigest: digest(marker)}
}
func receipt(input *Input, name string, cpu, allocated, wall, peak, read, write int64) ResourceReceipt {
	return ResourceReceipt{CommandID: id(name), SnapshotDigest: input.SnapshotDigest, ToolchainDigest: input.ToolchainDigest, RunnerDigest: input.RunnerDigest, CPUCoreNS: cpu, AllocatedCPUCount: allocated, WallNS: wall, PeakRSSBytes: peak, ReadBytes: read, WriteBytes: write}
}
func id(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "/", "-"), ":", "-")
}
func digest(marker string) string { return strings.Repeat(marker, 64) }
func selectOnly(input *Input, ids []string) {
	setSelected(input, ids)
	input.SelectedOutcomes = filterOutcomes(input.FullOutcomes, ids)
	input.SelectedResourceReceipts = filterReceipts(input.FullResourceReceipts, ids)
}
func filterOutcomes(values []Outcome, ids []string) []Outcome {
	set, _ := stringSet(ids)
	result := make([]Outcome, 0, len(ids))
	for _, value := range values {
		if _, keep := set[value.CommandID]; keep {
			result = append(result, value)
		}
	}
	return result
}
func filterReceipts(values []ResourceReceipt, ids []string) []ResourceReceipt {
	set, _ := stringSet(ids)
	result := make([]ResourceReceipt, 0, len(ids))
	for _, value := range values {
		if _, keep := set[value.CommandID]; keep {
			result = append(result, value)
		}
	}
	return result
}
func findOutcome(values []Outcome, commandID string) *Outcome {
	for index := range values {
		if values[index].CommandID == commandID {
			return &values[index]
		}
	}
	return nil
}
