package bodycodegen

type CriteriaDependencies struct {
	Schema     string                   `json:"schema"`
	Status     string                   `json:"status"`
	Reason     string                   `json:"reason,omitempty"`
	ActivityID string                   `json:"activity_id"`
	Edges      []CriteriaDependencyEdge `json:"edges,omitempty"`
	Scope      string                   `json:"scope"`
}

type CriteriaDependencyEdge struct {
	From        string `json:"from_activity_id"`
	To          string `json:"to_activity_id"`
	Kind        string `json:"kind"`
	CriterionID string `json:"criterion_id"`
	Ordinal     int    `json:"ordinal"`
}

// Project the already prepared source requirements and pure-call closures.
// This view adds no relation authority, execution or model input contract.
func criteriaDependencies(owner recordAssemblyPlan) *CriteriaDependencies {
	r := &CriteriaDependencies{Schema: "gooo/criteria-dependencies/v1", Status: "RESOLVED",
		ActivityID: owner.body.activityID,
		Scope:      "outgoing requires_predicate edges and bounded same-source pure calls; static source closure includes unexecuted branches; ordinal is predicate order or caller-local call order; no runtime result or reuse authorization"}
	for ordinal, criterion := range owner.criteria {
		if r.ActivityID == "" || criterion.activityID == "" {
			r.Status, r.Reason, r.Edges = "UNRESOLVED", "SOURCE_ACTIVITY_ID_UNAVAILABLE", nil
			return r
		}
		id := criterion.declaration.ID
		r.Edges = append(r.Edges, CriteriaDependencyEdge{From: r.ActivityID, To: criterion.activityID,
			Kind: "requires_predicate", CriterionID: id, Ordinal: ordinal})
		if closure := criterion.program.Report.CallClosure; closure != nil {
			for _, edge := range closure.Edges {
				if edge.CallerID == "" || edge.CalleeID == "" {
					r.Status, r.Reason, r.Edges = "UNRESOLVED", "SOURCE_CALL_ID_UNAVAILABLE", nil
					return r
				}
				r.Edges = append(r.Edges, CriteriaDependencyEdge{From: edge.CallerID, To: edge.CalleeID,
					Kind: "calls", CriterionID: id, Ordinal: edge.Ordinal})
			}
		}
	}
	return r
}
