package bodyexecution

// ConditionDependencyFact exposes bounded static origins to Gooo. AnalysisStatus
// describes the dependency analysis; SourceStatus describes the file mapping.
// Neither describes which path ran for a particular runtime input.
type ConditionDependencyFact struct {
	SiteID         string `json:"site_id"`
	ActivityID     string `json:"activity_id"`
	CandidateID    string `json:"candidate_id"`
	HoleID         string `json:"hole_id"`
	Relation       string `json:"relation"`
	AnalysisStatus string `json:"analysis_status"`
	SourceStatus   string `json:"source_status"`
	Filename       string `json:"filename"`
	SourceSHA      string `json:"source_sha"`
	SourceStart    int    `json:"source_start"`
	SourceEnd      int    `json:"source_end"`
}

func ConditionDependencyFacts(runtime CompositionRuntime) []ConditionDependencyFact {
	var facts []ConditionDependencyFact
	for _, site := range runtime.ConditionSites {
		if site.Source == nil || site.Source.Construction == nil {
			continue
		}
		link := site.Source.Construction
		if link.DataFlow == nil || link.Input == nil {
			continue
		}
		for _, fragment := range link.Input.Fragments {
			if fragment.HoleID == "" {
				continue
			}
			row := ConditionDependencyFact{SiteID: site.ExpressionID, ActivityID: link.Input.ActivityID,
				CandidateID: link.CandidateID, HoleID: fragment.HoleID, Relation: fragment.Relation,
				AnalysisStatus: link.DataFlow.Status, SourceStatus: fragment.Status,
				Filename: link.Input.Filename, SourceSHA: link.Input.SourceSHA256, SourceStart: -1, SourceEnd: -1}
			location := fragment.ExpressionSource
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
