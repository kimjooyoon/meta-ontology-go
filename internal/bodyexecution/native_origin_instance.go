package bodyexecution

// Instance facts include empty lists, whose element fact population is zero.
type CompositionConstructionInstanceFact struct {
	CaseIndex         int    `json:"case_index"`
	InstanceIndex     int    `json:"instance_index"`
	RootActivityID    string `json:"root_activity_id"`
	ActivityID        string `json:"activity_id"`
	Completed         bool   `json:"completed"`
	ElementCount      int    `json:"element_count"`
	AttemptedElements int    `json:"attempted_elements"`
	ReadStart         int    `json:"read_start"`
	ReadEnd           int    `json:"read_end"`
	SourceSHA256      string `json:"source_sha256"`
	SourceStart       int    `json:"source_start"`
	SourceEnd         int    `json:"source_end"`
	Expression        string `json:"expression"`
	Mapping           string `json:"mapping"`
}

func compositionConstructionInstanceFact(c, instance int, created CompositionConstruction, site CompositionConstructionSite) CompositionConstructionInstanceFact {
	fact := CompositionConstructionInstanceFact{CaseIndex: c, InstanceIndex: instance, RootActivityID: site.RootActivityID, ActivityID: site.ActivityID, Completed: created.Completed, ElementCount: len(site.Elements), AttemptedElements: len(created.Elements), ReadStart: created.ReadStart, ReadEnd: created.ReadEnd, Expression: site.Expression, Mapping: "UNMAPPED"}
	if source := site.Source; source != nil && source.Status == "MAPPED" {
		fact.Mapping = source.Status
		fact.SourceSHA256 = source.SourceSHA256
		fact.SourceStart = source.Start
		fact.SourceEnd = source.End
		fact.Expression = source.Expression
	}
	return fact
}
