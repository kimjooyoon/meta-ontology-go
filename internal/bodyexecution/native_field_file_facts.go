package bodyexecution

// Each row is one static original-file anchor attached to a dynamic field event.
// Stable field identity and the recorded assignment remain available for joins.
type CompositionFieldFileFact struct {
	CaseIndex          int    `json:"case_index"`
	EventKind          string `json:"event_kind"`
	EventIndex         int    `json:"event_index"`
	AssignmentIndex    int    `json:"assignment_index"`
	Frame              int    `json:"frame"`
	RootActivityID     string `json:"root_activity_id"`
	ActivityID         string `json:"activity_id"`
	VariableOrdinal    int    `json:"variable_ordinal"`
	VariableName       string `json:"variable_name"`
	RecordID           string `json:"record_id"`
	FieldID            string `json:"field_id"`
	SiteKind           string `json:"site_kind"`
	Category           string `json:"category"`
	Mapping            string `json:"mapping"`
	Reason             string `json:"reason"`
	PackagePath        string `json:"package_path"`
	Filename           string `json:"filename"`
	OriginalActivityID string `json:"original_activity_id"`
	SourceSHA256       string `json:"source_sha256"`
	SourceStart        int    `json:"source_start"`
	SourceEnd          int    `json:"source_end"`
	Expression         string `json:"expression"`
	CandidateID        string `json:"candidate_id"`
	HoleID             string `json:"hole_id"`
	Grammar            string `json:"grammar"`
}

func (page *nativeOriginPage) prepareFieldFileIndices(r CompositionRuntime) {
	if page.fieldFileBindings != nil {
		return
	}
	page.fieldFileBindings = &localFileBindingIndices{}
	if r.FieldFileSources == nil {
		return
	}
	page.fieldFileBindings.writes = make([]int, len(r.FieldSites))
	page.fieldFileBindings.uses = make([]int, len(r.FieldUseSites))
	for i, b := range r.FieldFileSources.Bindings {
		positions := page.fieldFileBindings.writes
		if b.Kind == "use" {
			positions = page.fieldFileBindings.uses
		}
		positions[b.SiteIndex] = i + 1
	}
}

func (page *nativeOriginPage) visitFieldFileLocations(r CompositionRuntime, t CompositionTrace) {
	page.prepareFieldFileIndices(r)
	for i, w := range t.FieldWrites {
		page.fieldFileLocation(r, t.CaseIndex, i, i, w.Frame, w.SiteIndex, true)
	}
	for i, u := range t.FieldUses {
		page.fieldFileLocation(r, t.CaseIndex, i, u.Assignment, u.Frame, u.SiteIndex, false)
	}
}

func (page *nativeOriginPage) fieldFileLocation(r CompositionRuntime, c, event, assignment, frame, index int, write bool) {
	sites, positions, kind := r.FieldUseSites, page.fieldFileBindings.uses, "use"
	if write {
		sites, positions, kind = r.FieldSites, page.fieldFileBindings.writes, "write"
	}
	site := sites[index]
	b := r.FieldBindings[site.Binding]
	base := CompositionFieldFileFact{CaseIndex: c, EventKind: kind, EventIndex: event, AssignmentIndex: assignment, Frame: frame, RootActivityID: b.RootActivityID, ActivityID: b.ActivityID, VariableOrdinal: b.VariableOrdinal, VariableName: b.Name, RecordID: b.Record.ID, FieldID: b.Record.Fields[site.Field].ID, SiteKind: site.Kind}
	binding := CompositionFieldFileBinding{}
	if len(positions) > 0 && positions[index] > 0 {
		binding = r.FieldFileSources.Bindings[positions[index]-1]
	}
	walkLocalFileAnchors(CompositionLocalFileFact{}, CompositionLocalFileBinding{Declaration: binding.Target, Value: binding.Value, Construction: binding.Construction}, "value", func(loc CompositionLocalFileFact) {
		fact := base
		fact.Category = loc.Category
		if fact.Category == "declaration" {
			fact.Category = "target"
		}
		fact.Mapping, fact.Reason, fact.PackagePath, fact.Filename, fact.OriginalActivityID = loc.Mapping, loc.Reason, loc.PackagePath, loc.Filename, loc.OriginalActivityID
		fact.SourceSHA256, fact.SourceStart, fact.SourceEnd, fact.Expression = loc.SourceSHA256, loc.SourceStart, loc.SourceEnd, loc.Expression
		fact.CandidateID, fact.HoleID, fact.Grammar = loc.CandidateID, loc.HoleID, loc.Grammar
		page.addField(fact, fact.RootActivityID, fact.ActivityID, fact.FieldID)
	})
	if binding.Choice != nil {
		fact := base
		location := binding.Choice.Value
		fact.Category = "record_choice"
		fact.Mapping, fact.Reason, fact.PackagePath, fact.Filename, fact.OriginalActivityID = location.Status, location.Reason, location.PackagePath, location.Filename, location.ActivityID
		fact.SourceSHA256, fact.SourceStart, fact.SourceEnd, fact.Expression = location.SourceSHA256, location.Start, location.End, location.Expression
		page.addField(fact, fact.RootActivityID, fact.ActivityID, fact.FieldID)
	}
}
