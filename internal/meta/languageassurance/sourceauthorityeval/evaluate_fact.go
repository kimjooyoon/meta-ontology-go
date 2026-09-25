package sourceauthorityeval

import "bytes"

func evaluateFact(fact Fact, index evidenceIndex) FactReceipt {
	receipt := FactReceipt{
		FactID:       fact.ID,
		SourceRef:    fact.SourceRef,
		AuthorityRef: fact.AuthorityRef,
	}
	if fact.SourceRef == "" || fact.SourceSnapshotDigest == "" ||
		fact.Span.Digest == "" || fact.ClaimDigest == "" ||
		fact.AuthorityRef == "" {
		return knownFailure(receipt, "REQUIRED_BINDING_MISSING")
	}
	source, exists := index.sources[fact.SourceRef]
	if !exists {
		return unknownFailure(receipt, "SOURCE_EVIDENCE_UNKNOWN")
	}
	actualSnapshot := DigestBytes(source.Bytes)
	if source.URI == "" || source.SnapshotDigest != actualSnapshot ||
		fact.SourceSnapshotDigest != actualSnapshot {
		return knownFailure(receipt, "SOURCE_SNAPSHOT_MISMATCH")
	}
	if fact.Span.Start < 0 || fact.Span.End <= fact.Span.Start ||
		fact.Span.End > len(source.Bytes) {
		return knownFailure(receipt, "SOURCE_SPAN_INVALID")
	}
	spanBytes := source.Bytes[fact.Span.Start:fact.Span.End]
	if fact.Span.Digest != DigestBytes(spanBytes) {
		return knownFailure(receipt, "SOURCE_SPAN_DIGEST_MISMATCH")
	}
	if fact.ClaimDigest != DigestBytes(fact.Claim) ||
		!bytes.Equal(fact.Claim, spanBytes) {
		return knownFailure(receipt, "CLAIM_NOT_EXACT_SOURCE_BYTES")
	}
	authority, exists := index.authorities[fact.AuthorityRef]
	if !exists {
		return unknownFailure(receipt, "AUTHORITY_EVIDENCE_UNKNOWN")
	}
	if authority.SourceRef != fact.SourceRef ||
		authority.SnapshotDigest != actualSnapshot ||
		authority.Start != fact.Span.Start || authority.End != fact.Span.End {
		return knownFailure(receipt, "AUTHORITY_SCOPE_MISMATCH")
	}
	receipt.Observation = "SATISFIED"
	receipt.Resolution = "EXACT"
	receipt.Reason = "SOURCE_AUTHORITY_EXACT"
	return receipt
}

func knownFailure(receipt FactReceipt, reason string) FactReceipt {
	receipt.Observation = "NOT_SATISFIED"
	receipt.Resolution = "EXACT"
	receipt.Reason = reason
	return receipt
}

func unknownFailure(receipt FactReceipt, reason string) FactReceipt {
	receipt.Observation = "UNKNOWN"
	receipt.Resolution = "INVARIANT_ONLY"
	receipt.Reason = reason
	return receipt
}
