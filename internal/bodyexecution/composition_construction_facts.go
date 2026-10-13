package bodyexecution

// Scalar facts fit Gooo records and retain case-local links to read observations.
type CompositionConstructionFact struct {
	CaseIndex      int    `json:"case_index"`
	InstanceIndex  int    `json:"instance_index"`
	ElementIndex   int    `json:"element_index"`
	RootActivityID string `json:"root_activity_id"`
	ActivityID     string `json:"activity_id"`
	Completed      bool   `json:"completed"`
	ReadStart      int    `json:"read_start"`
	ReadEnd        int    `json:"read_end"`
	ReadsDropped   int    `json:"reads_dropped"`
	SourceSHA256   string `json:"source_sha256"`
	SourceStart    int    `json:"source_start"`
	SourceEnd      int    `json:"source_end"`
	Expression     string `json:"expression"`
	Mapping        string `json:"mapping"`
}

func CompositionConstructionFacts(runtime CompositionRuntime) []CompositionConstructionFact {
	var facts []CompositionConstructionFact
	for _, trace := range runtime.Traces {
		for instance, created := range trace.Constructions {
			if created.SiteIndex < 0 || created.SiteIndex >= len(runtime.ConstructionSites) {
				continue
			}
			site := runtime.ConstructionSites[created.SiteIndex]
			for i, element := range created.Elements {
				if i >= len(site.Elements) {
					continue
				}
				f := CompositionConstructionFact{CaseIndex: trace.CaseIndex, InstanceIndex: instance, ElementIndex: i, RootActivityID: site.RootActivityID, ActivityID: site.ActivityID, Completed: element.Completed, ReadStart: element.ReadStart, ReadEnd: element.ReadEnd, ReadsDropped: element.ReadsDropped, Expression: site.Elements[i].Expression, Mapping: "UNMAPPED"}
				if source := site.Elements[i].Source; source != nil && source.Status == "MAPPED" {
					f.Mapping = source.Status
					f.SourceSHA256 = source.SourceSHA256
					f.SourceStart = source.Start
					f.SourceEnd = source.End
					f.Expression = source.Expression
				}
				facts = append(facts, f)
			}
		}
	}
	return facts
}

type CompositionConstructedReadFact struct {
	CaseIndex     int `json:"case_index"`
	ReadIndex     int `json:"read_index"`
	InstanceIndex int `json:"instance_index"`
	ElementIndex  int `json:"element_index"`
}

func CompositionConstructedReadFacts(runtime CompositionRuntime) []CompositionConstructedReadFact {
	var facts []CompositionConstructedReadFact
	for _, trace := range runtime.Traces {
		for i, read := range trace.Reads {
			if origin := read.Construction; origin != nil {
				facts = append(facts, CompositionConstructedReadFact{trace.CaseIndex, i, origin.Instance, origin.Index})
			}
		}
	}
	return facts
}
