package semanticdeltareceiptconsumer

type claimProposition struct {
	Kind      string `json:"kind"`
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

func propositionDigest(kind, subject, predicate, object string) string {
	return digestValue(claimProposition{Kind: kind, Subject: subject, Predicate: predicate, Object: object})
}

func objectClaimID(digest string) string {
	return "gooo://semantic-delta/claim/object/" + digest[len("sha256:"):]
}

func preservationClaimID(claim Claim) string {
	digest := propositionDigest(claimKindPreserve, claim.ID, "preserves", claim.PropositionDigest)
	return "gooo://semantic-delta/claim/preservation/" + digest[len("sha256:"):]
}
