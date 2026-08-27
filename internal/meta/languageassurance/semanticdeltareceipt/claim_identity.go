package semanticdeltareceipt

import "strings"

const claimIdentityVersion = "gooo://semantic-delta/claim-identity/v3"

const (
	rawIdentityRecreationMarker  = "fixture:claim-identity-recreated-due-only-to-raw-digest"
	rawIdentityRecreationVersion = claimIdentityVersion + "|fixture/raw-identity-recreation"
)

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

func objectClaimID(target, relationRole string) string {
	return objectClaimIDWithVersion(target, relationRole, claimIdentityVersion)
}

func objectClaimIDWithVersion(target, relationRole, identityVersion string) string {
	digest := digestValue(strings.Join([]string{identityVersion, ClaimKindObject, target, relationRole}, "\x00"))
	return "gooo://semantic-delta/claim/object/" + digest[len("sha256:"):]
}

func claimIdentityVersionForRaw(raw []byte) string {
	if strings.Contains(string(raw), rawIdentityRecreationMarker) {
		return rawIdentityRecreationVersion
	}
	return claimIdentityVersion
}

func boundedClaimID(target, relationRole string) string {
	digest := digestValue(strings.Join([]string{claimIdentityVersion, ClaimKindBounded, target, relationRole}, "\x00"))
	return "gooo://semantic-delta/claim/bounded-equivalence/" + digest[len("sha256:"):]
}

func preservationClaimIDForParts(preservationOf, target, relationRole string) string {
	digest := digestValue(strings.Join([]string{claimIdentityVersion, ClaimKindPreserve, preservationOf, target, relationRole}, "\x00"))
	return "gooo://semantic-delta/claim/preservation/" + digest[len("sha256:"):]
}

func preservationClaimID(before Claim) string {
	return preservationClaimIDForParts(before.ID, before.TargetAddress, "preserves")
}

func canonicalTargetAddress(subject, predicate, object string) string {
	return strings.Join([]string{subject, predicate, object}, "\x00")
}

func targetAddressDigest(address string) string { return digestValue(address) }

func canonicalPairTargetAddress(before, after string) string { return before + "->" + after }
