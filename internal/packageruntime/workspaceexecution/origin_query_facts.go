package workspaceexecution

// OriginUseFact gives Gooo one root/location pair at a time. Query paths and
// truncation remain available in OriginUseReport; facts do not imply coverage.
type OriginUseFact struct {
	ActivityID     string `json:"activity_id"`
	CandidateID    string `json:"candidate_id"`
	HoleID         string `json:"hole_id"`
	RootID         string `json:"root_id"`
	Kind           string `json:"kind"`
	Relation       string `json:"relation"`
	AnalysisStatus string `json:"analysis_status"`
	Expression     string `json:"expression"`
	SourceView     string `json:"source_view"`
	Filename       string `json:"filename"`
	SourceStart    int    `json:"source_start"`
	SourceEnd      int    `json:"source_end"`
	SourceSHA      string `json:"source_sha"`
}

func OriginUseFacts(report OriginUseReport) []OriginUseFact {
	var out []OriginUseFact
	for _, row := range report.Rows {
		expression := ""
		if len(row.Expressions) > 0 {
			expression = row.Expressions[0]
		}
		for _, location := range row.Locations {
			out = append(out, OriginUseFact{ActivityID: row.ActivityID, CandidateID: row.CandidateID, HoleID: row.Use.HoleID,
				RootID: row.Use.RootID, Kind: row.Use.RootKind, Relation: row.Use.Relation, AnalysisStatus: row.AnalysisStatus,
				Expression: expression, SourceView: location.View, Filename: location.Filename,
				SourceStart: location.Start, SourceEnd: location.End, SourceSHA: location.SourceSHA256})
		}
	}
	return out
}
