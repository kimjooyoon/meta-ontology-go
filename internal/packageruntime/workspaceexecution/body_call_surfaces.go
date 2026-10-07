package workspaceexecution

func (c workspaceCalls) rewriteActivityCalls(key string, d workspaceCallDeclaration) ([]WorkspaceCallSite, error) {
	var sites []WorkspaceCallSite
	rewrite := func(text *string, surface string, body bool) error {
		next, calls, err := c.rewriteCalls(key, d, *text, surface, body)
		if err == nil {
			*text = next
			sites = append(sites, calls...)
		}
		return err
	}
	if err := rewrite(&d.activity.ValueProgram, "computes", true); err != nil {
		return nil, err
	}
	if d.activity.Assembly == nil {
		return sites, nil
	}
	spec := &d.activity.Assembly.Spec
	if err := rewrite(&spec.Baseline, "baseline", true); err != nil {
		return nil, err
	}
	for i := range spec.Choices {
		choice := &spec.Choices[i]
		if choice.Kind == "field_value" || choice.Kind == "field_update" {
			if err := rewrite(&choice.Alternative, "choice:"+choice.ID, false); err != nil {
				return nil, err
			}
		}
	}
	if spec.FillPlan != nil {
		for i := range spec.FillPlan.Candidates {
			candidate := &spec.FillPlan.Candidates[i]
			for j := range candidate.Fills {
				fill := &candidate.Fills[j]
				if err := rewrite(&fill.Expression, "fill:"+candidate.ID+":"+fill.HoleID, false); err != nil {
					return nil, err
				}
			}
		}
	}
	return sites, nil
}
