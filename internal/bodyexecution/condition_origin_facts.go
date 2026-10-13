package bodyexecution

// ConditionOriginFact is a flat source-origin input for Gooo metaprograms.
// The adapter describes available links; source owns their interpretation.
type ConditionOriginFact struct {
	SiteID      string `json:"site_id"`
	Selected    bool   `json:"selected"`
	Original    bool   `json:"original"`
	Declaration bool   `json:"declaration"`
	Fields      int64  `json:"ref_count"`
	Located     int64  `json:"located_count"`
}

func ConditionOriginFacts(runtime CompositionRuntime) []ConditionOriginFact {
	facts := make([]ConditionOriginFact, len(runtime.ConditionSites))
	for i, site := range runtime.ConditionSites {
		fact := &facts[i]
		fact.SiteID = site.ExpressionID
		if site.Source == nil {
			continue
		}
		source := site.Source
		fact.Selected = source.Status == "MAPPED"
		fact.Original = source.Input != nil && source.Input.Status == "MAPPED"
		fact.Declaration = source.Declaration != nil && source.Declaration.Status == "MAPPED"
		fact.Fields = int64(len(source.Fields))
		for _, field := range source.Fields {
			if field.Declaration != nil && field.Declaration.Status == "MAPPED" {
				fact.Located++
			}
		}
	}
	return facts
}
