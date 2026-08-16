package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func requirements(input Input) []pathclosure.Requirement {
	requirements := make([]pathclosure.Requirement, 0, len(input.Paths))
	for _, path := range input.Paths {
		requirements = append(requirements, pathclosure.Requirement{
			PathID: path.PathID, RecordIDs: path.RecordIDs, ExpectedKinds: path.ExpectedKinds,
			StartID: path.RootID, EndID: path.ReceiptID,
		})
	}
	return requirements
}

func makeReceipt(input Input, status DecisionStatus, fallback FallbackMode, code string) Receipt {
	paths := make([]semantic.ID, 0, len(input.Paths))
	for _, path := range input.Paths {
		paths = append(paths, path.PathID)
	}
	receipt := Receipt{
		Schema: SchemaVersion, Status: status, Fallback: fallback, Code: code,
		Snapshots: input.Snapshots, RegistryDigest: input.RegistryDigest, PlanDigest: input.PlanDigest,
		SelectedCommandIDs: append([]semantic.ID(nil), input.SelectedCommandIDs...),
		ObligationIDs:      append([]semantic.ID(nil), input.ObligationIDs...), PathIDs: paths,
		RequiredCommandCount: len(input.SelectedCommandIDs), RequiredObligationCount: len(input.ObligationIDs),
	}
	if status == Verified {
		receipt.VerifiedCommandCount = len(input.SelectedCommandIDs)
		receipt.VerifiedObligationCount = len(input.ObligationIDs)
		receipt.VerifiedPathCount = len(input.Paths)
		receipt.VerifiedCommandIDs = append([]semantic.ID(nil), input.SelectedCommandIDs...)
		receipt.VerifiedObligationIDs = append([]semantic.ID(nil), input.ObligationIDs...)
		receipt.VerifiedPathIDs = append([]semantic.ID(nil), paths...)
	}
	receipt.Digest = receipt.expectedDigest()
	return receipt
}

// Evaluate performs no selection or execution. It verifies only the supplied
// finite named paths and returns a deterministic receipt for that evidence.
func Evaluate(raw Input) Receipt {
	input, state := normalizeInput(raw)
	if state.class != issueNone {
		if state.class == issueFailClosed {
			return makeReceipt(input, FailClosed, FullSuite, state.code)
		}
		return makeReceipt(input, Unknown, FullSuite, state.code)
	}
	if len(input.SelectedCommandIDs) == 0 || len(input.ObligationIDs) == 0 {
		return makeReceipt(input, Unknown, FullSuite, CodeMissing)
	}
	if len(input.SelectedCommandIDs) != len(input.ObligationIDs) {
		return makeReceipt(input, FailClosed, FullSuite, CodeAmbiguous)
	}
	if len(input.Paths) < len(input.SelectedCommandIDs) {
		return makeReceipt(input, Unknown, FullSuite, CodeMissing)
	}
	if len(input.Paths) > len(input.SelectedCommandIDs) {
		return makeReceipt(input, FailClosed, FullSuite, CodeAmbiguous)
	}
	if issue := validateBindings(input); issue.class != issueNone {
		return makeReceipt(input, issueStatus(issue.class), FullSuite, issue.code)
	}
	if issue := validateChains(input); issue.class != issueNone {
		return makeReceipt(input, issueStatus(issue.class), FullSuite, issue.code)
	}
	closure := pathclosure.Evaluate(input.InferencePath, requirements(input))
	if closure.Status == pathclosure.UNKNOWN {
		return makeReceipt(input, Unknown, FullSuite, CodeMissing)
	}
	if closure.Status != pathclosure.PASS {
		return makeReceipt(input, FailClosed, FullSuite, CodeMalformed)
	}
	return makeReceipt(input, Verified, NoFallback, CodeVerified)
}

func issueStatus(class issueClass) DecisionStatus {
	if class == issueFailClosed {
		return FailClosed
	}
	return Unknown
}

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

func receiptsFor(input Input) map[semantic.ID]CommandReceipt {
	receipts := make(map[semantic.ID]CommandReceipt, len(input.CommandReceipts))
	for _, receipt := range input.CommandReceipts {
		receipts[receipt.CommandID] = receipt
	}
	return receipts
}
