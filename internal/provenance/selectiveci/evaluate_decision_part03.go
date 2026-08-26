package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateBindings(input Input) issueState {
	state := issueState{}
	commandSet := make(map[semantic.ID]struct{}, len(input.SelectedCommandIDs))
	for _, commandID := range input.SelectedCommandIDs {
		commandSet[commandID] = struct{}{}
	}
	obligationSet := make(map[semantic.ID]struct{}, len(input.ObligationIDs))
	for _, obligationID := range input.ObligationIDs {
		obligationSet[obligationID] = struct{}{}
	}
	rootSet := make(map[semantic.ID]struct{}, len(input.ChangedRootIDs))
	for _, rootID := range input.ChangedRootIDs {
		rootSet[rootID] = struct{}{}
	}
	seenCommands := map[semantic.ID]struct{}{}
	seenObligations := map[semantic.ID]struct{}{}
	seenRoots := map[semantic.ID]struct{}{}
	receipts := receiptsFor(input)
	for _, path := range input.Paths {
		if _, ok := commandSet[path.CommandID]; !ok {
			state.add(issueFailClosed, CodeWrongEndpoint)
		}
		if _, ok := obligationSet[path.ObligationID]; !ok {
			state.add(issueFailClosed, CodeWrongEndpoint)
		}
		if _, ok := rootSet[path.RootID]; !ok {
			state.add(issueFailClosed, CodeWrongEndpoint)
		}
		if receipt, ok := receipts[path.CommandID]; !ok || receipt.ReceiptID != path.ReceiptID {
			state.add(issueFailClosed, CodeWrongEndpoint)
		}
		if _, exists := seenCommands[path.CommandID]; exists {
			state.add(issueFailClosed, CodeDuplicate)
		}
		if _, exists := seenObligations[path.ObligationID]; exists {
			state.add(issueFailClosed, CodeDuplicate)
		}
		seenCommands[path.CommandID] = struct{}{}
		seenObligations[path.ObligationID] = struct{}{}
		seenRoots[path.RootID] = struct{}{}
	}
	if len(seenCommands) != len(commandSet) || len(seenObligations) != len(obligationSet) {
		state.add(issueUnknown, CodeMissing)
	}
	if len(seenRoots) != len(rootSet) {
		state.add(issueUnknown, CodeMissing)
	}
	for _, commandID := range input.SelectedCommandIDs {
		receipt, exists := receipts[commandID]
		if !exists {
			state.add(issueUnknown, CodeMissing)
			continue
		}
		if receipt.Status != ReceiptVerified {
			state.add(issueUnknown, CodeCandidate)
		}
	}
	for _, receipt := range input.CommandReceipts {
		if !containsID(input.SelectedCommandIDs, receipt.CommandID) {
			state.add(issueFailClosed, CodeWrongEndpoint)
		}
		if receipt.Digest != receipt.expectedDigest(input.Snapshots) {
			state.add(issueFailClosed, CodeReceiptMismatch)
		}
	}
	return state
}
