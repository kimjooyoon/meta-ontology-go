package semanticdeltareceiptconsumer

import (
	"fmt"
	"sort"
)

func claimsFromFacts(facts []Fact) []Claim {
	counts := map[string]int{}
	claims := make([]Claim, 0, len(facts))
	for _, fact := range facts {
		subject, predicate, object := fact.Subject, "uses", fact.Object
		if fact.Predicate == semanticWasGeneratedBy {
			subject, predicate, object = fact.Object, "generates", fact.Subject
		}
		key := subject + "\x00" + predicate
		index := counts[key]
		counts[key]++
		claims = append(claims, Claim{ID: fmt.Sprintf("%s/claim/%s/%d", subject, predicate, index), Subject: subject, Predicate: predicate, Object: object, Status: statusOpen, Stage: "semantic-extraction", Step: "bind-canonical-fact", Reason: "CANONICAL_LOWERING_BOUND"})
	}
	sort.Slice(claims, func(i, j int) bool { return claims[i].ID < claims[j].ID })
	return claims
}

const (
	semanticWasGeneratedBy = "wasGeneratedBy"
	statusOpen             = "OPEN"
)
