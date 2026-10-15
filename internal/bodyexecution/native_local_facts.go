package bodyexecution

type CompositionLocalWriteFact struct {
	CompositionLocalWrite
	CaseIndex       int    `json:"case_index"`
	AssignmentIndex int    `json:"assignment_index"`
	RootActivityID  string `json:"root_activity_id"`
	ActivityID      string `json:"activity_id"`
	VariableOrdinal int    `json:"variable_ordinal"`
	VariableName    string `json:"variable_name"`
	ScalarType      string `json:"scalar_type"`
	SourceSHA256    string `json:"source_sha256"`
	SourceStart     int    `json:"source_start"`
	SourceEnd       int    `json:"source_end"`
	Expression      string `json:"expression"`
	Mapping         string `json:"mapping"`
}

type CompositionLocalUseFact struct {
	SourceSHA256       string `json:"source_sha256"`
	SourceStart        int    `json:"source_start"`
	SourceEnd          int    `json:"source_end"`
	SourceMapping      string `json:"source_mapping"`
	CaseIndex          int    `json:"case_index"`
	UseIndex           int    `json:"use_index"`
	Frame              int    `json:"frame"`
	RootActivityID     string `json:"root_activity_id"`
	ActivityID         string `json:"activity_id"`
	VariableOrdinal    int    `json:"variable_ordinal"`
	VariableName       string `json:"variable_name"`
	AssignmentIndex    int    `json:"assignment_index"`
	Resolved           bool   `json:"resolved"`
	Reason             string `json:"reason"`
	ScalarType         string `json:"scalar_type"`
	ProjectionSHA256   string `json:"projection_sha256"`
	ProjectionStart    int    `json:"projection_start"`
	ProjectionEnd      int    `json:"projection_end"`
	DeclarationSHA256  string `json:"declaration_sha256"`
	DeclarationStart   int    `json:"declaration_start"`
	DeclarationEnd     int    `json:"declaration_end"`
	DeclarationMapping string `json:"declaration_mapping"`
}

type CompositionLocalElementFact struct {
	CaseIndex int `json:"case_index"`
	CompositionLocalElementWindow
}

func (page *nativeOriginPage) visitLocals(runtime CompositionRuntime, trace CompositionTrace) {
	switch page.query.Kind {
	case "local-input-locations":
		page.visitLocalFileLocations(runtime, trace)
	case "local-fill-fragments":
		page.visitLocalFillFragments(runtime, trace)
	case "local-writes":
		for index, w := range trace.LocalWrites {
			site := runtime.LocalSites[w.SiteIndex]
			fact := CompositionLocalWriteFact{CompositionLocalWrite: w, CaseIndex: trace.CaseIndex, AssignmentIndex: index,
				RootActivityID: site.RootActivityID, ActivityID: site.ActivityID, VariableOrdinal: site.VariableOrdinal,
				VariableName: site.Name, ScalarType: site.ScalarType, Expression: site.Expression, Mapping: "UNMAPPED"}
			if s := site.Source.Value; s != nil && s.Status == "MAPPED" {
				fact.SourceSHA256, fact.SourceStart, fact.SourceEnd, fact.Expression, fact.Mapping = s.SourceSHA256, s.Start, s.End, s.Expression, s.Status
			}
			page.add(fact)
		}
	case "local-uses":
		for index, u := range trace.LocalUses {
			site := runtime.LocalUseSites[u.SiteIndex]
			fact := CompositionLocalUseFact{CaseIndex: trace.CaseIndex, UseIndex: index, Frame: u.Frame, RootActivityID: site.RootActivityID, ActivityID: site.ActivityID,
				VariableOrdinal: site.VariableOrdinal, VariableName: site.Name, AssignmentIndex: u.Assignment, Resolved: u.Assignment >= 0, Reason: u.Reason,
				ScalarType: site.ScalarType, ProjectionSHA256: site.ProjectionSHA256, ProjectionStart: site.Start, ProjectionEnd: site.End, DeclarationMapping: "UNMAPPED", SourceMapping: "UNMAPPED"}
			if s := site.Source; s != nil && s.Status == "MAPPED" {
				fact.SourceSHA256, fact.SourceStart, fact.SourceEnd, fact.SourceMapping = s.SourceSHA256, s.Start, s.End, s.Status
			}
			if d := site.Declaration; d != nil && d.Status == "MAPPED" {
				fact.DeclarationSHA256, fact.DeclarationStart, fact.DeclarationEnd, fact.DeclarationMapping = d.SourceSHA256, d.Start, d.End, d.Status
			}
			page.add(fact)
		}
	case "local-elements":
		for _, e := range trace.LocalElements {
			page.add(CompositionLocalElementFact{trace.CaseIndex, e})
		}
	}
}
