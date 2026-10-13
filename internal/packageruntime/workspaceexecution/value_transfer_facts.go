package workspaceexecution

// ProgramOriginFact is a bounded scalar record that a Gooo activity can consume.
// One row represents one returned path/location; counts are path structure,
// not runtime call counts or whole-program completeness.
type ProgramOriginFact struct {
	RootActivityID   string `json:"root_activity_id"`
	RootID           string `json:"root_id"`
	SourceActivityID string `json:"source_activity_id"`
	OriginKind       string `json:"origin_kind"`
	CandidateID      string `json:"candidate_id"`
	HoleID           string `json:"hole_id"`
	CallReturns      int    `json:"call_returns"`
	BindSteps        int    `json:"bind_steps"`
	SourceView       string `json:"source_view"`
	Filename         string `json:"filename"`
	SourceSHA        string `json:"source_sha"`
	SourceStart      int    `json:"source_start"`
	SourceEnd        int    `json:"source_end"`
}

func ProgramOriginFacts(report ProgramValueOrigins) []ProgramOriginFact {
	var out []ProgramOriginFact
	for _, row := range report.Rows {
		fact := ProgramOriginFact{RootActivityID: report.RootActivityID, RootID: row.RootID, SourceActivityID: row.ActivityID, OriginKind: row.Kind, CandidateID: row.CandidateID, HoleID: row.HoleID}
		for _, step := range row.Path {
			if step.Via == "call_return" {
				fact.CallReturns++
			}
			if step.Via == "bind" {
				fact.BindSteps++
			}
		}
		if len(row.Locations) == 0 {
			fact.SourceView = "unavailable"
			out = append(out, fact)
		}
		for _, location := range row.Locations {
			value := fact
			value.SourceView = location.View
			value.Filename = location.Filename
			value.SourceSHA = location.SourceSHA256
			value.SourceStart = location.Start
			value.SourceEnd = location.End
			out = append(out, value)
		}
	}
	return out
}
