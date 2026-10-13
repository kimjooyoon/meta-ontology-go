package bodyexecution

// ConditionFillInputFact carries the original package-file view into a Gooo
// metaprogram. An unresolved location uses -1 offsets and an explicit status.
type ConditionFillInputFact struct {
	SiteID      string `json:"site_id"`
	ActivityID  string `json:"activity_id"`
	CandidateID string `json:"candidate_id"`
	HoleID      string `json:"hole_id"`
	Category    string `json:"category"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
	PackagePath string `json:"package_path"`
	Filename    string `json:"filename"`
	SourceSHA   string `json:"source_sha"`
	SourceStart int    `json:"source_start"`
	SourceEnd   int    `json:"source_end"`
}

func ConditionFillInputFacts(runtime CompositionRuntime) []ConditionFillInputFact {
	var facts []ConditionFillInputFact
	for _, site := range runtime.ConditionSites {
		if site.Source == nil || site.Source.Construction == nil || site.Source.Construction.Input == nil {
			continue
		}
		origin := site.Source.Construction.Input
		for _, fragment := range origin.Fragments {
			row := ConditionFillInputFact{SiteID: site.ExpressionID, ActivityID: origin.ActivityID,
				CandidateID: origin.CandidateID, HoleID: fragment.HoleID, Category: fragment.Kind,
				Status: fragment.Status, Reason: fragment.Reason, PackagePath: origin.PackagePath,
				Filename: origin.Filename, SourceSHA: origin.SourceSHA256, SourceStart: -1, SourceEnd: -1}
			location := fragment.ExpressionSource
			if fragment.Kind == "retained_template" {
				location = fragment.TemplateSource
			}
			if fragment.Kind == "derived_grammar" {
				location = fragment.GrammarSource
			}
			if location != nil {
				row.SourceStart, row.SourceEnd = location.Start, location.End
			}
			facts = append(facts, row)
		}
	}
	return facts
}
