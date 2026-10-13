package bodyexecution

// Read facts use scalar fields so a Gooo record program can consume them as
// input. No process pointer or duplicated collection value crosses this boundary.
type CompositionReadFact struct {
	CaseIndex      int    `json:"case_index"`
	RootActivityID string `json:"root_activity_id"`
	ActivityID     string `json:"activity_id"`
	Kind           string `json:"kind"`
	Index          int    `json:"index"`
	Length         int    `json:"length"`
	Status         string `json:"status"`
	InputKey       string `json:"input_key"`
	FieldID        string `json:"field_id"`
	OriginIndex    int    `json:"origin_index"`
	SourceSHA256   string `json:"source_sha256"`
	SourceStart    int    `json:"source_start"`
	SourceEnd      int    `json:"source_end"`
	Expression     string `json:"expression"`
}

func CompositionReadFacts(runtime CompositionRuntime) []CompositionReadFact {
	var facts []CompositionReadFact
	for _, trace := range runtime.Traces {
		for _, event := range trace.Reads {
			if event.SiteIndex < 0 || event.SiteIndex >= len(runtime.ReadSites) {
				continue
			}
			site := runtime.ReadSites[event.SiteIndex]
			fact := CompositionReadFact{CaseIndex: trace.CaseIndex, RootActivityID: site.RootActivityID, ActivityID: site.ActivityID, Kind: site.Kind,
				Index: event.Index, Length: event.Length, Status: "UNRESOLVED", OriginIndex: -1, Expression: site.Expression}
			if site.Source != nil && site.Source.Status == "MAPPED" {
				fact.SourceSHA256 = site.Source.SourceSHA256
				fact.SourceStart = site.Source.Start
				fact.SourceEnd = site.Source.End
				fact.Expression = site.Source.Expression
			}
			if origin := event.Origin; origin != nil && origin.Input >= 0 && origin.Input < len(runtime.ReadInputs) {
				input := runtime.ReadInputs[origin.Input]
				fact.Status = "RESOLVED"
				fact.InputKey = input.InputKey
				fact.FieldID = input.FieldID
				fact.OriginIndex = origin.Index
			}
			if origin := event.Construction; origin != nil {
				fact.Status = "CONSTRUCTED"
				fact.OriginIndex = origin.Index
			}
			facts = append(facts, fact)
		}
	}
	return facts
}
