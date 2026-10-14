package bodyexecution

// Each fact is one original-file anchor for an observed write/use. Fragment
// anchors may cover an entire candidate expression or grammar declaration.
type CompositionLocalFileFact struct {
	CaseIndex          int    `json:"case_index"`
	EventKind          string `json:"event_kind"`
	EventIndex         int    `json:"event_index"`
	AssignmentIndex    int    `json:"assignment_index"`
	Frame              int    `json:"frame"`
	RootActivityID     string `json:"root_activity_id"`
	ActivityID         string `json:"activity_id"`
	VariableOrdinal    int    `json:"variable_ordinal"`
	VariableName       string `json:"variable_name"`
	ScalarType         string `json:"scalar_type"`
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

type localFileBindingIndices struct{ writes, uses []int }

// Build dense integer indexes once per selection, sharing the retained table
// across all cases. Zero means absent; other entries are binding index + 1.
func (page *nativeOriginPage) prepareLocalFileIndices(runtime CompositionRuntime) {
	if page.fileBindings != nil {
		return
	}
	page.fileBindings = &localFileBindingIndices{}
	if runtime.LocalFileSources == nil {
		return
	}
	page.fileBindings.writes = make([]int, len(runtime.LocalSites))
	page.fileBindings.uses = make([]int, len(runtime.LocalUseSites))
	for index, binding := range runtime.LocalFileSources.Bindings {
		positions := page.fileBindings.writes
		if binding.Kind == "use" {
			positions = page.fileBindings.uses
		}
		positions[binding.SiteIndex] = index + 1
	}
}

func (page *nativeOriginPage) localFileBinding(runtime CompositionRuntime, index int, write bool) CompositionLocalFileBinding {
	positions := page.fileBindings.uses
	if write {
		positions = page.fileBindings.writes
	}
	if len(positions) == 0 || positions[index] == 0 {
		return CompositionLocalFileBinding{}
	}
	return runtime.LocalFileSources.Bindings[positions[index]-1]
}

func (page *nativeOriginPage) visitLocalFileLocations(runtime CompositionRuntime, trace CompositionTrace) {
	page.prepareLocalFileIndices(runtime)
	for index, w := range trace.LocalWrites {
		site := runtime.LocalSites[w.SiteIndex]
		fact := CompositionLocalFileFact{CaseIndex: trace.CaseIndex, EventKind: "write", EventIndex: index, AssignmentIndex: index, Frame: w.Frame,
			RootActivityID: site.RootActivityID, ActivityID: site.ActivityID, VariableOrdinal: site.VariableOrdinal, VariableName: site.Name, ScalarType: site.ScalarType}
		page.addLocalFileBinding(fact, page.localFileBinding(runtime, w.SiteIndex, true), "value")
	}
	for index, u := range trace.LocalUses {
		site := runtime.LocalUseSites[u.SiteIndex]
		fact := CompositionLocalFileFact{CaseIndex: trace.CaseIndex, EventKind: "use", EventIndex: index, AssignmentIndex: u.Assignment, Frame: u.Frame,
			RootActivityID: site.RootActivityID, ActivityID: site.ActivityID, VariableOrdinal: site.VariableOrdinal, VariableName: site.Name, ScalarType: site.ScalarType}
		page.addLocalFileBinding(fact, page.localFileBinding(runtime, u.SiteIndex, false), "use")
	}
}

func (page *nativeOriginPage) addLocalFileBinding(base CompositionLocalFileFact, binding CompositionLocalFileBinding, category string) {
	walkLocalFileAnchors(base, binding, category, func(fact CompositionLocalFileFact) { page.add(fact) })
}

func walkLocalFileAnchors(base CompositionLocalFileFact, binding CompositionLocalFileBinding, category string, visit func(CompositionLocalFileFact)) {
	for index, source := range []CompositionLocalFileLocation{binding.Declaration, binding.Value} {
		fact := base
		fact.Category = "declaration"
		if index == 1 {
			fact.Category = category
		}
		fact.Mapping, fact.Reason = source.Status, source.Reason
		if fact.Mapping == "" {
			fact.Mapping, fact.Reason = "UNMAPPED", "ORIGINAL_FILE_ANNOTATION_ABSENT"
		}
		fact.PackagePath, fact.Filename, fact.OriginalActivityID, fact.SourceSHA256 = source.PackagePath, source.Filename, source.ActivityID, source.SourceSHA256
		fact.SourceStart, fact.SourceEnd, fact.Expression = source.Start, source.End, source.Expression
		visit(fact)
	}
	c := binding.Construction
	if c == nil {
		return
	}
	for _, fragment := range c.Fragments {
		fact := base
		fact.Category, fact.Mapping, fact.Reason = fragment.Kind, fragment.Status, fragment.Reason
		fact.PackagePath, fact.Filename, fact.OriginalActivityID, fact.SourceSHA256 = c.PackagePath, c.Filename, c.ActivityID, c.SourceSHA256
		fact.CandidateID, fact.HoleID, fact.Grammar, fact.Expression = c.CandidateID, fragment.HoleID, fragment.Grammar, fragment.Expression
		at := fragment.TemplateSource
		if fragment.Kind == "declared_candidate" {
			at = fragment.ExpressionSource
		}
		if fragment.Kind == "derived_grammar" {
			at = fragment.GrammarSource
		}
		if fact.Mapping == "MAPPED" && at != nil {
			fact.SourceStart, fact.SourceEnd = at.Start, at.End
		}
		visit(fact)
	}
}
