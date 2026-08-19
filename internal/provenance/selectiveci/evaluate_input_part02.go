package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strings"
)

func normalizeReceipts(raw []CommandReceipt, input Input, state *issueState) []CommandReceipt {
	receipts := make([]CommandReceipt, 0, len(raw))
	commands, receiptIDs := map[semantic.ID]struct{}{}, map[semantic.ID]struct{}{}
	for _, rawReceipt := range raw {
		receipt := normalizeCommandReceipt(rawReceipt, input, state)
		if _, exists := commands[receipt.CommandID]; exists {
			state.add(issueFailClosed, CodeDuplicate)
		}
		if _, exists := receiptIDs[receipt.ReceiptID]; exists {
			state.add(issueFailClosed, CodeDuplicate)
		}
		commands[receipt.CommandID], receiptIDs[receipt.ReceiptID] = struct{}{}, struct{}{}
		receipts = append(receipts, receipt)
	}
	sort.Slice(receipts, func(i, j int) bool { return receipts[i].CommandID < receipts[j].CommandID })
	if len(receipts) == 0 {
		state.add(issueUnknown, CodeMissing)
	}
	return receipts
}
func normalizePathEvidence(raw Input, input Input, state *issueState) (Input, issueState) {
	normalized, err := raw.InferencePath.Normalized()
	if err != nil {
		state.add(classifyInferencePathError(err), classifyInferencePathCode(err))
		return input, *state
	}
	input.InferencePath = normalized
	for _, edge := range normalized.Edges {
		if edge.Kind == semantic.InferenceObservationCandidate {
			state.add(issueUnknown, CodeCandidate)
		}
	}
	if err := bindEvidenceIDs(input); err != nil {
		state.add(classifyBindingError(err), classifyBindingCode(err))
	}
	if err := bindSnapshots(input); err != nil {
		state.add(classifyBindingError(err), classifyBindingCode(err))
	}
	return input, *state
}
func classifyInferencePathError(err error) issueClass {
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "stable-id-collision") || strings.Contains(message, "duplicate") || strings.Contains(message, "stale-evidence") {
		return issueFailClosed
	}
	if strings.Contains(message, "evidence") || strings.Contains(message, "snapshot") || strings.Contains(message, "candidate") {
		return issueUnknown
	}
	return issueFailClosed
}
