package semanticdeltareceiptconsumer

import "sort"

func claimsFromFacts(facts []Fact) []Claim {
	claims := make([]Claim, 0, len(facts))
	for _, fact := range facts {
		subject, predicate, object := fact.Subject, "uses", fact.Object
		if fact.Predicate == semanticWasGeneratedBy {
			subject, predicate, object = fact.Object, "generates", fact.Subject
		}
		digest := propositionDigest(claimKindObject, subject, predicate, object)
		claims = append(claims, Claim{ID: objectClaimID(digest), Kind: claimKindObject, Subject: subject, Predicate: predicate, Object: object, Status: statusOpen, Stage: "semantic-extraction", Step: "bind-canonical-fact", Reason: "CANONICAL_LOWERING_BOUND", PropositionDigest: digest})
	}
	sort.Slice(claims, func(i, j int) bool { return claims[i].ID < claims[j].ID })
	return claims
}

const (
	semanticWasGeneratedBy = "wasGeneratedBy"
)
