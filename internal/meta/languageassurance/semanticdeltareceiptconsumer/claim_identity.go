package semanticdeltareceiptconsumer

import "strings"

func propositionDigest(kind, subject, predicate, object string) string {
	return digestValue(normalizedProposition(kind, subject, predicate, object))
}

func normalizedProposition(kind, subject, predicate, object string) string {
	return strings.Join([]string{kind, subject, predicate, object}, "\x00")
}

func claimTypeID(kind, subject, predicate, object string) string {
	digest := propositionDigest(kind, subject, predicate, object)
	return "gooo://semantic-delta/claim-type/" + digest[len("sha256:"):]
}

func objectClaimID(normalized, target, rawDigest, semanticDigest string) string {
	digest := digestValue(strings.Join([]string{normalized, target, rawDigest, semanticDigest}, "\x00"))
	return "gooo://semantic-delta/claim/object/" + digest[len("sha256:"):]
}

func preservationClaimID(before, after Claim, afterObject string) string {
	digest := digestValue(strings.Join([]string{before.ID, after.ID, afterObject, before.NormalizedProposition, after.AfterSourceDigest, after.AfterSemanticDigest}, "\x00"))
	return "gooo://semantic-delta/claim/preservation/" + digest[len("sha256:"):]
}
