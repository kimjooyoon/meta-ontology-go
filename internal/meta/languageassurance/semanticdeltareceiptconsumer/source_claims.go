package semanticdeltareceiptconsumer

import "sort"

func claimsFromFacts(facts []Fact, filename, rawDigest, semanticDigest string, before bool) []Claim {
	claims := make([]Claim, 0, len(facts))
	for _, fact := range facts {
		subject, predicate, object := fact.Subject, "uses", fact.Object
		if fact.Predicate == semanticWasGeneratedBy {
			subject, predicate, object = fact.Object, "generates", fact.Subject
		}
		normalized := normalizedProposition(claimKindObject, subject, predicate, object)
		proposition := propositionDigest(claimKindObject, subject, predicate, object)
		claim := Claim{ID: objectClaimID(normalized, filename, rawDigest, semanticDigest), ClaimTypeID: claimTypeID(claimKindObject, subject, predicate, object), Kind: claimKindObject, Subject: subject, Predicate: predicate, Object: object, Status: statusOpen, Stage: "semantic-extraction", Step: "bind-canonical-fact", Reason: "CANONICAL_LOWERING_BOUND", NormalizedProposition: normalized, PropositionDigest: proposition, TargetAddress: filename}
		if before {
			claim.BeforeSourceDigest, claim.BeforeSemanticDigest = rawDigest, semanticDigest
		} else {
			claim.AfterSourceDigest, claim.AfterSemanticDigest = rawDigest, semanticDigest
		}
		claims = append(claims, claim)
	}
	sort.Slice(claims, func(i, j int) bool { return claims[i].ID < claims[j].ID })
	return claims
}

func factLess(left, right Fact) bool {
	if left.Subject != right.Subject {
		return left.Subject < right.Subject
	}
	if left.Predicate != right.Predicate {
		return left.Predicate < right.Predicate
	}
	return left.Object < right.Object
}

const semanticWasGeneratedBy = "wasGeneratedBy"
