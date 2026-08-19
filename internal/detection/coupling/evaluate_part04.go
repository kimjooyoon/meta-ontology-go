package coupling

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
