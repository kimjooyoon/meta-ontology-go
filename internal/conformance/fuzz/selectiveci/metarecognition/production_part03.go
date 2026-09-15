package metarecognition

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/fullsoundness"
)

func soundnessInput(b BaselineConfig) fullsoundness.Input {
	obligationID := b.Obligation.ID
	if obligationID == "" {
		obligationID = "obl-impact"
	}
	authority := fullsoundness.AuthorityAuthoritative
	if b.Obligation.Authority == Candidate {
		authority = fullsoundness.AuthorityCandidate
	}
	obligations := []fullsoundness.Obligation{{ID: obligationID, Authority: authority}}
	commands := make([]fullsoundness.Command, 0, len(b.Commands))
	full, selected := make([]fullsoundness.Outcome, 0, len(b.Commands)), make([]fullsoundness.Outcome, 0, len(b.Commands))
	fullReceipts, selectedReceipts := make([]fullsoundness.ResourceReceipt, 0, len(b.Commands)), make([]fullsoundness.ResourceReceipt, 0, len(b.Commands))
	selectedIDs := make([]string, 0, len(b.Commands))
	for _, command := range b.Commands {
		commandObligation := obligationID
		if command.ID == "cmd-candidate" {
			commandObligation = "obl-candidate"
			if obligationID != commandObligation {
				obligations = append(obligations, fullsoundness.Obligation{ID: commandObligation, Authority: fullsoundness.AuthorityCandidate})
			}
		}
		commands = append(commands, fullsoundness.Command{ID: command.ID, ObligationIDs: []string{commandObligation}, GlobalGuard: command.GlobalGuard})
		full = append(full, fullOutcome(command.ID, command.FullStatus, command.FullDigest))
		fullReceipts = append(fullReceipts, receipt(command.ID, digest("s")))
		if command.Selected {
			selected = append(selected, fullOutcome(command.ID, command.SelectedStatus, command.SelectedDigest))
			selectedReceipts = append(selectedReceipts, receipt(command.ID, digest("s")))
			selectedIDs = append(selectedIDs, command.ID)
		}
	}
	input := fullsoundness.Input{SchemaVersion: fullsoundness.SchemaVersion, SnapshotDigest: digest("s"), PolicyDigest: digest("p"), RegistryDigest: digest("r"), SelectionDigest: digest("l"), ToolchainDigest: digest("t"), RunnerDigest: digest("u"), Obligations: obligations, Commands: commands, ImpactedObligationIDs: []string{obligationID}, SelectedCommandIDs: selectedIDs, SelectionReceipt: &fullsoundness.SelectionReceipt{SnapshotDigest: digest("s"), PolicyDigest: digest("p"), RegistryDigest: digest("r"), SelectionDigest: digest("l"), CommandIDs: selectedIDs}, FullOutcomes: full, SelectedOutcomes: selected, FullResourceReceipts: fullReceipts, SelectedResourceReceipts: selectedReceipts}
	if externalMissing(b.External) {
		if !b.External.Authenticity {
			input.ToolchainDigest = ""
		}
		if !b.External.Provider {
			input.RunnerDigest = ""
		}
		if !b.External.Phase {
			input.SelectionReceipt = nil
		}
		if !b.External.Observer {
			input.FullResourceReceipts, input.SelectedResourceReceipts = nil, nil
		}
	}
	return input
}
func fullOutcome(id string, status Status, seed string) fullsoundness.Outcome {
	result := fullsoundness.Outcome{CommandID: id, OutputDigest: digest(seed)}
	if status == Fail {
		result.Status, result.FailureCode = fullsoundness.OutcomeFail, "FAIL"
	} else {
		result.Status = fullsoundness.OutcomePass
	}
	return result
}
func receipt(id, snapshot string) fullsoundness.ResourceReceipt {
	return fullsoundness.ResourceReceipt{CommandID: id, SnapshotDigest: snapshot, ToolchainDigest: digest("t"), RunnerDigest: digest("u"), CPUCoreNS: 1, AllocatedCPUCount: 1, WallNS: 1, PeakRSSBytes: 1, ReadBytes: 1, WriteBytes: 1}
}
