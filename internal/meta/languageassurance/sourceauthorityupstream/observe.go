package sourceauthorityupstream

import "context"

func Observe(ctx context.Context, policy Policy, request Request, fetcher Fetcher) Receipt {
	receipt := Receipt{
		Schema: ReceiptSchema, SubjectSHA: request.SubjectSHA, RequestDigest: digestValue(request), Mode: ModeExternal,
		Observation: ObservationUnknown, Resolution: ResolutionInvariantOnly, Enforcement: EnforcementBlock,
	}
	if reason := validate(policy, request); reason != "" {
		return finishReceipt(receipt, reason)
	}
	if fetcher == nil {
		return finishReceipt(receipt, ReasonRequestInvalid)
	}
	document, err := fetcher.Fetch(ctx, request.URL)
	if err != nil {
		return finishReceipt(receipt, ReasonFetchFailed)
	}
	selected, err := selectLines(document, request.Selection)
	if err != nil {
		return finishReceipt(receipt, ReasonSelectionFailed)
	}
	snapshot := Snapshot{
		SourceRef: request.SourceRef, AuthorityRef: request.AuthorityRef, URL: request.URL,
		Authority: request.Authority, Selection: request.Selection, Digest: digestBytes(selected), Bytes: len(selected),
	}
	receipt.Snapshot = &snapshot
	if snapshot.Bytes != policy.ExpectedBytes {
		return finishReceipt(receipt, ReasonSourceSizeMismatch)
	}
	if snapshot.Digest != policy.ExpectedDigest {
		return finishReceipt(receipt, ReasonSourceDigestMismatch)
	}
	receipt.Observation = ObservationSatisfied
	receipt.Resolution = ResolutionExact
	receipt.Enforcement = EnforcementAllow
	return finishReceipt(receipt, ReasonSourceSnapshotExact)
}

func finishReceipt(receipt Receipt, reason string) Receipt {
	receipt.Reason = reason
	receipt.Indicators = buildIndicators(receipt)
	unsigned := receipt
	unsigned.ReceiptDigest = ""
	receipt.ReceiptDigest = digestValue(unsigned)
	return receipt
}
