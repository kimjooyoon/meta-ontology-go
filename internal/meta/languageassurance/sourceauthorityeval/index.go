package sourceauthorityeval

import "sort"

type evidenceIndex struct {
	sources     map[string]Source
	authorities map[string]Authority
}

func buildIndex(bundle Bundle) (evidenceIndex, string) {
	index := evidenceIndex{
		sources:     make(map[string]Source, len(bundle.Sources)),
		authorities: make(map[string]Authority, len(bundle.Authorities)),
	}
	for _, source := range bundle.Sources {
		if source.ID == "" {
			return evidenceIndex{}, "SOURCE_ID_MISSING"
		}
		if _, exists := index.sources[source.ID]; exists {
			return evidenceIndex{}, "DUPLICATE_SOURCE_ID"
		}
		index.sources[source.ID] = source
	}
	for _, authority := range bundle.Authorities {
		if authority.ID == "" {
			return evidenceIndex{}, "AUTHORITY_ID_MISSING"
		}
		if _, exists := index.authorities[authority.ID]; exists {
			return evidenceIndex{}, "DUPLICATE_AUTHORITY_ID"
		}
		index.authorities[authority.ID] = authority
	}
	return index, ""
}

func acceptedFacts(facts []Fact) ([]Fact, string) {
	seen := make(map[string]struct{}, len(facts))
	accepted := make([]Fact, 0, len(facts))
	for _, fact := range facts {
		if fact.ID == "" {
			return nil, "FACT_ID_MISSING"
		}
		if _, exists := seen[fact.ID]; exists {
			return nil, "DUPLICATE_FACT_ID"
		}
		seen[fact.ID] = struct{}{}
		switch fact.State {
		case "ACCEPTED":
			accepted = append(accepted, fact)
		case "CANDIDATE":
		default:
			return nil, "FACT_STATE_UNKNOWN"
		}
	}
	sort.Slice(accepted, func(left, right int) bool {
		return accepted[left].ID < accepted[right].ID
	})
	return accepted, ""
}
