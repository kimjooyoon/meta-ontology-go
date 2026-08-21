package coupling

func validateReceiptIdentity(receipt CouplingReceipt, entry ManifestEntry, config Config, surface Surface) *evaluationIssue {
	if receipt.Schema != ReceiptSchemaV1 {
		return failIssue(ReasonMalformedBinding, receipt.SurfaceID.String())
	}
	if receipt.State != ReceiptStateCurrent {
		return unknownIssue(ReasonStaleInput, receipt.SurfaceID.String())
	}
	if _, issue := normalizeID(receipt.ReceiptID, "receipt ID"); issue != nil {
		return issue
	}
	if receipt.SurfaceID != surface.SurfaceID || receipt.CodeSymbolID != surface.CodeSymbolID || receipt.SemanticOwnerID != surface.SemanticOwnerID {
		return failIssue(ReasonSourceMapMismatch, receipt.SurfaceID.String())
	}
	if receipt.SourceMapBindingDigest != surface.Binding.BindingDigest {
		return failIssue(ReasonSourceMapMismatch, receipt.SurfaceID.String())
	}
	if receipt.SnapshotDigest != config.SnapshotDigest || receipt.RegistryDigest != config.RegistryDigest ||
		receipt.ToolchainDigest != config.ToolchainDigest || receipt.ProfileDigest != config.ProfileDigest {
		return unknownIssue(ReasonStaleInput, receipt.SurfaceID.String())
	}
	for _, value := range []struct {
		value string
		name  string
	}{
		{receipt.SourceMapBindingDigest, "receipt binding digest"},
		{receipt.SnapshotDigest, "receipt snapshot digest"},
		{receipt.RegistryDigest, "receipt registry digest"},
		{receipt.ToolchainDigest, "receipt toolchain digest"},
		{receipt.ProfileDigest, "receipt profile digest"},
		{receipt.BeforeBlobDigest, "receipt before blob digest"},
		{receipt.AfterBlobDigest, "receipt after blob digest"},
		{receipt.BeforeAuthoritySourceDigest, "receipt before source digest"},
		{receipt.AfterAuthoritySourceDigest, "receipt after source digest"},
		{receipt.BeforeCanonicalSemanticDigest, "receipt before semantic digest"},
		{receipt.AfterCanonicalSemanticDigest, "receipt after semantic digest"},
	} {
		if issue := normalizeDigestValue(value.value, value.name); issue != nil {
			return issue
		}
	}
	if receipt.BeforeBlobDigest != entry.BeforeBlobDigest || receipt.AfterBlobDigest != entry.AfterBlobDigest {
		return failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
	}
	return nil
}
