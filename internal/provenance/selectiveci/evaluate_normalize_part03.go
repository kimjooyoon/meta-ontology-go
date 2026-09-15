package selectiveci

func normalizeCommandReceipt(raw CommandReceipt, input Input, state *issueState) CommandReceipt {
	receipt := raw
	var err error
	if receipt.CommandID, err = normalizeID(raw.CommandID, "command receipt command ID"); err != nil {
		state.add(issueFailClosed, CodeMalformed)
	}
	if receipt.ReceiptID, err = normalizeID(raw.ReceiptID, "command receipt ID"); err != nil {
		state.add(issueFailClosed, CodeMalformed)
	}
	for label, value := range map[string]string{
		"provider receipt": raw.ProviderReceiptDigest,
		"phase receipt":    raw.PhaseReceiptDigest,
		"resource receipt": raw.ResourceReceiptDigest,
		"registry":         raw.RegistryDigest,
		"plan":             raw.PlanDigest,
	} {
		if value == "" {
			state.add(issueUnknown, CodeMissing)
			continue
		}
		if _, err := normalizeDigest(value, label+" digest"); err != nil {
			state.add(issueFailClosed, CodeDigestMismatch)
		}
	}
	receipt.ProviderReceiptDigest, _ = normalizeDigest(raw.ProviderReceiptDigest, "provider receipt")
	receipt.PhaseReceiptDigest, _ = normalizeDigest(raw.PhaseReceiptDigest, "phase receipt")
	receipt.ResourceReceiptDigest, _ = normalizeDigest(raw.ResourceReceiptDigest, "resource receipt")
	receipt.RegistryDigest, _ = normalizeDigest(raw.RegistryDigest, "registry")
	receipt.PlanDigest, _ = normalizeDigest(raw.PlanDigest, "plan")
	if receipt.RegistryDigest != input.RegistryDigest || receipt.PlanDigest != input.PlanDigest {
		state.add(issueFailClosed, CodeDigestMismatch)
	}
	switch receipt.Status {
	case ReceiptVerified, ReceiptCandidate, ReceiptDeferred, ReceiptNotRun:
	default:
		state.add(issueFailClosed, CodeMalformed)
	}
	if receipt.Digest == "" {
		state.add(issueUnknown, CodeMissing)
	} else if _, err := normalizeDigest(receipt.Digest, "command receipt digest"); err != nil {
		state.add(issueFailClosed, CodeReceiptMismatch)
	}
	receipt.Digest, _ = normalizeDigest(raw.Digest, "command receipt digest")
	return receipt
}
