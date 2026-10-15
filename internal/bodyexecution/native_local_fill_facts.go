package bodyexecution

// Fragment source ranges locate the assembly input, and can cover more than the
// written RHS when it overlaps a retained template fragment. Decoded ranges name
// the selected written expression separately.
type LocalFillFragmentFact struct {
	CaseIndex            int    `json:"case_index"`
	AssignmentIndex      int    `json:"assignment_index"`
	Frame                int    `json:"frame"`
	Completed            bool   `json:"completed"`
	SiteID               string `json:"site_id"`
	VariableName         string `json:"variable_name"`
	ActivityID           string `json:"activity_id"`
	CandidateID          string `json:"candidate_id"`
	HoleID               string `json:"hole_id"`
	Category             string `json:"category"`
	Filename             string `json:"filename"`
	SourceSHA            string `json:"source_sha"`
	SourceStart          int    `json:"source_start"`
	SourceEnd            int    `json:"source_end"`
	SelectedDecodedStart int    `json:"selected_decoded_start"`
	SelectedDecodedEnd   int    `json:"selected_decoded_end"`
}

func (page *nativeOriginPage) visitLocalFillFragments(runtime CompositionRuntime, trace CompositionTrace) {
	for index, write := range trace.LocalWrites {
		site := runtime.LocalSites[write.SiteIndex]
		origin := site.Source.Construction
		if origin == nil {
			continue
		}
		for _, fragment := range origin.Fragments {
			location := fragment.ExpressionSource
			if location == nil {
				location = fragment.GrammarSource
			}
			if location == nil {
				location = &fragment.TemplateSource
			}
			page.add(LocalFillFragmentFact{CaseIndex: trace.CaseIndex, AssignmentIndex: index,
				Frame: write.Frame, Completed: write.Completed, SiteID: site.ExpressionID,
				VariableName: site.Name, ActivityID: origin.ActivityID, CandidateID: origin.CandidateID,
				HoleID: fragment.HoleID, Category: fragment.Kind, Filename: origin.Filename,
				SourceSHA: origin.SourceSHA256, SourceStart: location.Start, SourceEnd: location.End,
				SelectedDecodedStart: origin.LocalDecoded.Start, SelectedDecodedEnd: origin.LocalDecoded.End})
		}
	}
}
