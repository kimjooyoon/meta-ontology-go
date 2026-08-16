package coupling

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

func Evaluate(input Input, authorities ...AuthorityContext) (result Result) {
	authority := AuthorityContext{}
	if len(authorities) == 1 {
		authority = authorities[0]
	}
	inputDigest := inputIdentityDigest(input, authority)
	defer func() {
		result.InputDigest = inputDigest
		result.Digest = stableDigest(resultCanonical(result))
	}()
	if len(authorities) != 1 || authority.Schema == "" {
		return resultFor(StatusUnknown, ReasonAuthorityInputSelfBound, "evaluator authority context is missing", ObservationVector{})
	}
	authorityConfig, issue := normalizeAuthorityContext(authority)
	if issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	if input.Schema != InputSchemaV1 {
		return resultFor(StatusFailClosed, ReasonMalformedBinding, "input schema", ObservationVector{})
	}
	if issue := normalizeConfig(input.Config); issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	if issue := validatePacketApplicability(input, authority); issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	if issue := comparePacketAuthority(input, authorityConfig, authority); issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	registry, issue := normalizeRegistry(input.Registry, authorityConfig)
	if issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	changed, issue := normalizeManifest(input.Manifest, authorityConfig, registry)
	if issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, ObservationVector{})
	}
	observation := baseObservation(len(changed), len(input.Receipts))
	if issue := validateExternalReceipt(input.ExternalReceipt, authorityConfig, &observation); issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, observation)
	}
	if len(changed) == 0 {
		if len(input.Receipts) != 0 {
			return resultFor(StatusFailClosed, ReasonOrphanReceipt, "receipt for zero-change manifest", observation)
		}
		observation.DeterministicWork = knownDimension(0)
		return passResult(nil, observation)
	}
	path, issue := normalizeInferencePath(input.InferencePath)
	if issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, observation)
	}
	observation.InferenceRecords = knownDimension(uint64(len(path.Edges) + len(path.Claims) + len(path.Evidence)))
	observation.InferencePaths = knownDimension(uint64(len(input.Receipts)))
	observation.DeterministicWork = knownDimension(uint64(len(changed) + len(input.Receipts) + len(path.Edges) + len(path.Claims) + len(path.Evidence)))
	receipts, issue := indexReceipts(input.Receipts, changed)
	if issue != nil {
		return resultFor(issue.status, issue.code, issue.detail, observation)
	}
	accepted := make([]semantic.ID, 0, len(changed))
	for _, entry := range changed {
		receipt, exists := receipts[entry.SurfaceID]
		if !exists {
			return resultFor(StatusUnknown, ReasonRequiredInputMissing, "receipt for changed surface", observation)
		}
		if issue := validateReceipt(receipt, entry, authorityConfig, registry[entry.SurfaceID]); issue != nil {
			return resultFor(issue.status, issue.code, issue.detail, observation)
		}
		if issue := validateReceiptPath(receipt, entry, path); issue != nil {
			return resultFor(issue.status, issue.code, issue.detail, observation)
		}
		accepted = append(accepted, entry.SurfaceID)
	}
	return passResult(accepted, observation)
}

func baseObservation(changed, receipts int) ObservationVector {
	return ObservationVector{
		ChangedSurfaces: knownDimension(uint64(changed)),
		Receipts:        knownDimension(uint64(receipts)),
	}
}

func knownDimension(value uint64) CountDimension { return CountDimension{Known: true, Value: value} }

func indexReceipts(receipts []CouplingReceipt, changed []ManifestEntry) (map[semantic.ID]CouplingReceipt, *evaluationIssue) {
	changedIDs := make(map[semantic.ID]struct{}, len(changed))
	for _, entry := range changed {
		changedIDs[entry.SurfaceID] = struct{}{}
	}
	indexed := make(map[semantic.ID]CouplingReceipt, len(receipts))
	seenReceiptIDs := make(map[semantic.ID]struct{}, len(receipts))
	for _, receipt := range receipts {
		if _, duplicate := seenReceiptIDs[receipt.ReceiptID]; duplicate {
			return nil, failIssue(ReasonDuplicateReceipt, receipt.SurfaceID.String())
		}
		seenReceiptIDs[receipt.ReceiptID] = struct{}{}
		if _, expected := changedIDs[receipt.SurfaceID]; !expected {
			return nil, failIssue(ReasonOrphanReceipt, receipt.SurfaceID.String())
		}
		if _, duplicate := indexed[receipt.SurfaceID]; duplicate {
			return nil, failIssue(ReasonDuplicateReceipt, receipt.SurfaceID.String())
		}
		indexed[receipt.SurfaceID] = receipt
	}
	return indexed, nil
}

func validateReceipt(receipt CouplingReceipt, entry ManifestEntry, config Config, surface Surface) *evaluationIssue {
	if issue := validateReceiptIdentity(receipt, entry, config, surface); issue != nil {
		return issue
	}
	if issue := validateReceiptClaim(receipt); issue != nil {
		return issue
	}
	return validateReceiptReferences(receipt)
}

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

func validateReceiptClaim(receipt CouplingReceipt) *evaluationIssue {
	wantKind, validClaim := semanticKindForClaim(receipt.ChangeClaim)
	if !validClaim || receipt.ReceiptKind != wantKind {
		return failIssue(ReasonContradictoryReceipt, receipt.SurfaceID.String())
	}
	switch receipt.ChangeClaim {
	case ChangeClaimDelta:
		if receipt.BeforeCanonicalSemanticDigest == receipt.AfterCanonicalSemanticDigest {
			return failIssue(ReasonContradictoryReceipt, receipt.SurfaceID.String())
		}
		if receipt.BeforeAuthoritySourceDigest == receipt.AfterAuthoritySourceDigest {
			return failIssue(ReasonDeltaWithoutSource, receipt.SurfaceID.String())
		}
		if receipt.CanonicalDelta == "" || receipt.DeltaDigest == "" ||
			stableDigest(receipt.CanonicalDelta) != receipt.DeltaDigest {
			return failIssue(ReasonDigestMismatch, receipt.SurfaceID.String())
		}
	case ChangeClaimNoDelta:
		if receipt.BeforeCanonicalSemanticDigest != receipt.AfterCanonicalSemanticDigest ||
			receipt.BeforeAuthoritySourceDigest != receipt.AfterAuthoritySourceDigest {
			return failIssue(ReasonNoDeltaWithoutEquality, receipt.SurfaceID.String())
		}
		if receipt.CanonicalDelta != "" || receipt.DeltaDigest != "" || receipt.AuthoritativeSource != nil {
			return failIssue(ReasonNoDeltaWithoutEquality, receipt.SurfaceID.String())
		}
	default:
		return failIssue(ReasonMalformedBinding, receipt.SurfaceID.String())
	}
	if receipt.AuthoritativeSource != nil {
		if _, issue := normalizeID(receipt.AuthoritativeSource.SourceID, "authority source ID"); issue != nil {
			return issue
		}
	}
	return nil
}
