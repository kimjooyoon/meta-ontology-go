package bodyexecution

// ConditionFillOriginFact is a flat source record for Gooo metaprograms. Source
// offsets locate the expression or grammar in the assembly input named here.
type ConditionFillOriginFact struct {
	SiteID      string `json:"site_id"`
	ActivityID  string `json:"activity_id"`
	CandidateID string `json:"candidate_id"`
	HoleID      string `json:"hole_id"`
	Category    string `json:"category"`
	Filename    string `json:"filename"`
	SourceSHA   string `json:"source_sha"`
	SourceStart int    `json:"source_start"`
	SourceEnd   int    `json:"source_end"`
}

// ConditionFillOriginFacts copies available construction origins. It does not
// infer provenance for old records or interpret the categories for the caller.
func ConditionFillOriginFacts(runtime CompositionRuntime) []ConditionFillOriginFact {
	var facts []ConditionFillOriginFact
	for _, site := range runtime.ConditionSites {
		if site.Source == nil || site.Source.Construction == nil {
			continue
		}
		origin := site.Source.Construction
		for _, fragment := range origin.Fragments {
			location := fragment.ExpressionSource
			if location == nil {
				location = fragment.GrammarSource
			}
			if location == nil {
				continue
			}
			facts = append(facts, ConditionFillOriginFact{site.ExpressionID, origin.ActivityID,
				origin.CandidateID, fragment.HoleID, fragment.Kind, origin.Filename,
				origin.SourceSHA256, location.Start, location.End})
		}
	}
	return facts
}
