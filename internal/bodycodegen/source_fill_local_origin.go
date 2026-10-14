package bodycodegen

// SourceFillLocalOrigin locates fragments used to assemble a written expression.
// It describes source construction, not a minimal runtime dependency slice.
type SourceFillLocalOrigin struct {
	Schema               string               `json:"schema"`
	Filename             string               `json:"filename"`
	View                 string               `json:"view"`
	SourceSHA256         string               `json:"source_sha256"`
	SelectedSourceSHA256 string               `json:"selected_source_sha256"`
	SelectedBodySHA256   string               `json:"selected_body_sha256"`
	ActivityID           string               `json:"activity_id"`
	Intent               string               `json:"intent"`
	IntentSource         FillOriginRange      `json:"intent_source"`
	CandidateID          string               `json:"candidate_id"`
	CandidateSource      *FillOriginRange     `json:"candidate_source,omitempty"`
	PlanSHA256           string               `json:"plan_sha256"`
	LocalDecoded         FillOriginRange      `json:"local_decoded"`
	Fragments            []SourceFillFragment `json:"fragments"`
}

func (o SourceFillOrigin) LocalOrigin(local BodyLocalOrigin) *SourceFillLocalOrigin {
	if o.Schema != "gooo/source-fill-origin/v1" || o.ActivityID == "" || o.SourceSHA256 == "" || o.SelectedBodySHA256 == "" ||
		local.ActivityID != o.ActivityID || local.BodySHA256 != o.SelectedBodySHA256 ||
		local.Start < 0 || local.End <= local.Start {
		return nil
	}
	at := FillOriginRange{local.Start, local.End}
	fragments := localFillFragments(o.Fragments, at)
	if len(fragments) == 0 {
		return nil
	}
	return &SourceFillLocalOrigin{Schema: "gooo/source-fill-local-origin/v1",
		Filename: o.Filename, View: o.View, SourceSHA256: o.SourceSHA256,
		SelectedSourceSHA256: o.SelectedSourceSHA256, SelectedBodySHA256: o.SelectedBodySHA256,
		ActivityID: o.ActivityID, Intent: o.Intent, IntentSource: o.IntentSource,
		CandidateID: o.CandidateID, CandidateSource: cloneFillOriginRange(o.CandidateSource),
		PlanSHA256: o.PlanSHA256, LocalDecoded: at, Fragments: fragments}
}

func localFillFragments(fragments []SourceFillFragment, at FillOriginRange) []SourceFillFragment {
	var result []SourceFillFragment
	cursor := at.Start
	for _, fragment := range fragments {
		if !fillRangesOverlap(fragment.SelectedDecoded, at) {
			continue
		}
		if fragment.SelectedDecoded.Start > cursor {
			return nil
		}
		cursor = max(cursor, fragment.SelectedDecoded.End)
		fragment.Relation = "written_expression"
		fragment.HoleDeclaration = cloneFillOriginRange(fragment.HoleDeclaration)
		fragment.ExpressionSource = cloneFillOriginRange(fragment.ExpressionSource)
		fragment.GrammarSource = cloneFillOriginRange(fragment.GrammarSource)
		result = append(result, fragment)
	}
	if cursor < at.End {
		return nil
	}
	return result
}
