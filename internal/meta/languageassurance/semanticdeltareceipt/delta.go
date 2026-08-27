package semanticdeltareceipt

import (
	"reflect"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/semanticdelta"
)

func textualDelta(before, after []byte) TextualDelta {
	changedBytes := 0
	for index := 0; index < len(before) && index < len(after); index++ {
		if before[index] != after[index] {
			changedBytes++
		}
	}
	if len(before) > len(after) {
		changedBytes += len(before) - len(after)
	} else {
		changedBytes += len(after) - len(before)
	}
	return TextualDelta{Changed: !reflect.DeepEqual(before, after), BeforeBytes: len(before), AfterBytes: len(after),
		ChangedBytes: changedBytes, BeforeDigest: digestBytes(before), AfterDigest: digestBytes(after)}
}

func structuralDelta(before, after projectedSource) (StructuralDelta, error) {
	delta, err := semanticdelta.DiffSnapshots(toSemanticSnapshot(before), toSemanticSnapshot(after))
	if err != nil {
		return StructuralDelta{}, err
	}
	result := StructuralDelta{Status: "KNOWN"}
	for _, node := range delta.AddedNodes {
		result.AddedNodes = append(result.AddedNodes, Node{ID: node.ID, Kind: node.Kind})
	}
	for _, node := range delta.RemovedNodes {
		result.RemovedNodes = append(result.RemovedNodes, Node{ID: node.ID, Kind: node.Kind})
	}
	for _, fact := range delta.AddedFacts {
		result.AddedFacts = append(result.AddedFacts, Fact{Subject: fact.Subject, Predicate: fact.Predicate, Object: fact.Object})
	}
	for _, fact := range delta.RemovedFacts {
		result.RemovedFacts = append(result.RemovedFacts, Fact{Subject: fact.Subject, Predicate: fact.Predicate, Object: fact.Object})
	}
	return result, nil
}

func unknownStructuralDelta() StructuralDelta { return StructuralDelta{Status: "UNKNOWN"} }

func claimDelta(before, after projectedSource) ClaimDelta {
	left, right := claimMap(before.claims), claimMap(after.claims)
	result := ClaimDelta{Status: "KNOWN"}
	for id, claim := range left {
		other, exists := right[id]
		if !exists {
			result.Removed = append(result.Removed, claim)
			continue
		}
		if !claimMeaningEqual(claim, other) {
			result.Changed = append(result.Changed, ClaimChange{ID: id, Before: claim, After: other})
		}
	}
	for id, claim := range right {
		if _, exists := left[id]; !exists {
			result.Added = append(result.Added, claim)
		}
	}
	sort.Slice(result.Added, func(i, j int) bool { return result.Added[i].ID < result.Added[j].ID })
	sort.Slice(result.Removed, func(i, j int) bool { return result.Removed[i].ID < result.Removed[j].ID })
	sort.Slice(result.Changed, func(i, j int) bool { return result.Changed[i].ID < result.Changed[j].ID })
	return result
}

func unknownClaimDelta() ClaimDelta { return ClaimDelta{Status: "UNKNOWN"} }

func claimMap(claims []Claim) map[string]Claim {
	result := make(map[string]Claim, len(claims))
	for _, claim := range claims {
		result[claim.ID] = claim
	}
	return result
}

func claimMeaningEqual(left, right Claim) bool {
	return left.ID == right.ID && left.Subject == right.Subject && left.Predicate == right.Predicate &&
		left.Object == right.Object && left.Status == right.Status
}

func claimDigest(claims []Claim) string { return digestValue(claims) }

func structuralDigest(source projectedSource) string {
	return digestValue(structuralImage{Nodes: source.nodes, Facts: source.facts})
}

type structuralImage struct {
	Nodes []Node `json:"nodes"`
	Facts []Fact `json:"facts"`
}

func snapshot(raw []byte, source projectedSource, err error) Snapshot {
	result := Snapshot{SourceDigest: digestBytes(raw), Bytes: len(raw), Lines: lineCount(raw), ParseStatus: "EXACT", ParseReason: "SOURCE_PARSED"}
	if err != nil {
		result.ParseStatus, result.ParseReason = "UNKNOWN", "UNSUPPORTED_GOOO_SOURCE"
		return result
	}
	result.StructuralDigest = structuralDigest(source)
	result.ClaimDigest = claimDigest(source.claims)
	result.Nodes, result.Facts, result.Claims = source.nodes, source.facts, source.claims
	return result
}

func lineCount(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	count := 1
	for _, value := range raw {
		if value == '\n' {
			count++
		}
	}
	return count
}

func hasSemanticDelta(delta StructuralDelta, claims ClaimDelta) bool {
	return len(delta.AddedNodes)+len(delta.RemovedNodes)+len(delta.AddedFacts)+len(delta.RemovedFacts)+
		len(claims.Added)+len(claims.Removed)+len(claims.Changed) > 0
}

func transitions(before, after projectedSource, class, reason string) []ClaimTransition {
	if class == ClassIndeterminate {
		return []ClaimTransition{{ClaimID: "semantic-delta/unknown", FromStatus: "ASSERTED", ToStatus: "UNKNOWN",
			Stage: "adjudicate", Step: "claim-transition", Reason: reason}}
	}
	left, right := claimMap(before.claims), claimMap(after.claims)
	ids := make([]string, 0, len(left)+len(right))
	seen := map[string]bool{}
	for id := range left {
		seen[id] = true
		ids = append(ids, id)
	}
	for id := range right {
		if !seen[id] {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	result := make([]ClaimTransition, 0, len(ids))
	for _, id := range ids {
		from, to := "ABSENT", "ABSENT"
		fromObject, toObject := "", ""
		if claim, ok := left[id]; ok {
			from = claim.Status
			fromObject = claim.Object
		}
		if claim, ok := right[id]; ok {
			to = claim.Status
			toObject = claim.Object
		}
		transitionReason := "CLAIM_SEMANTIC_FIXED_POINT"
		if class == ClassChanged {
			transitionReason = "CLAIM_SEMANTIC_DELTA"
		}
		result = append(result, ClaimTransition{ClaimID: id, FromStatus: from, ToStatus: to, FromObject: fromObject, ToObject: toObject,
			Stage: "adjudicate", Step: "claim-transition", Reason: transitionReason})
	}
	return result
}
