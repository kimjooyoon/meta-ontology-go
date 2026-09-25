package coupling

func finish(output Output, decision Decision, reason Reason) Output {
	output.Decision, output.Reason = decision, reason
	output.AcceptedSurfaces = sortedUnique(output.AcceptedSurfaces)
	output.ChangedSurfaces = sortedUnique(output.ChangedSurfaces)
	output.ReceiptSurfaces = sortedUnique(output.ReceiptSurfaces)
	output.CanonicalOutputDigest = CanonicalOutputDigest(output)
	output.ReplayDigest = ReplayDigest(output.InputDigest, output.CanonicalOutputDigest)
	return output
}
func validateRequiredInput(input Input) oracleValidation {
	if input.Schema != SchemaV1 || input.Config.ToolchainDigest == "" || input.Config.Profile.ID == "" ||
		input.Config.Profile.Version == "" || input.Config.Profile.Digest == "" ||
		input.AuthoritySourceBefore == "" || input.AuthoritySourceAfter == "" ||
		len(input.Registry) == 0 || len(input.ResourceReceipts) == 0 || !input.Manifest.Complete ||
		(len(input.Changes) == 0 && !input.Manifest.ZeroChange) {
		return oracleValidation{DecisionUnknown, ReasonRequiredInputMissing}
	}
	if !validDigest(input.Config.ToolchainDigest) || !validDigest(input.Config.Profile.Digest) {
		return oracleValidation{DecisionUnknown, ReasonRequiredInputMissing}
	}
	return oracleValidation{}
}
