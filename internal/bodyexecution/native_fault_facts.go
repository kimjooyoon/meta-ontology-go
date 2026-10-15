package bodyexecution

// One row names one observed fault. SourceRelation distinguishes a local RHS
// anchor from the complete operation in the pure projection.
type CompositionFaultFact struct {
	CaseIndex        int    `json:"case_index"`
	DeliveryIndex    int    `json:"delivery_index"`
	RootActivityID   string `json:"root_activity_id"`
	ActivityID       string `json:"activity_id"`
	ExpressionID     string `json:"expression_id"`
	Kind             string `json:"kind"`
	Operator         string `json:"operator"`
	Assignment       string `json:"assignment"`
	Variable         string `json:"variable"`
	Left             int64  `json:"left"`
	Right            int64  `json:"right"`
	ProjectionSHA256 string `json:"projection_sha256"`
	ProjectionStart  int    `json:"projection_start"`
	ProjectionEnd    int    `json:"projection_end"`
	Expression       string `json:"expression"`
	SourceMapping    string `json:"source_mapping"`
	SourceRelation   string `json:"source_relation"`
	SourceSHA256     string `json:"source_sha256"`
	SourceStart      int    `json:"source_start"`
	SourceEnd        int    `json:"source_end"`
	SourceExpression string `json:"source_expression"`
}

func (page *nativeOriginPage) visitFaults(runtime CompositionRuntime, trace CompositionTrace) {
	for index, delivery := range trace.Deliveries {
		if fault := delivery.Fault; fault != nil {
			fact := recordedFaultFact(runtime, trace.CaseIndex, index, *fault)
			page.addField(fact, fact.RootActivityID, fact.ActivityID, "")
		}
	}
}

func recordedFaultFact(runtime CompositionRuntime, caseIndex, deliveryIndex int, fault CompositionFault) CompositionFaultFact {
	site := fault.Site
	fact := CompositionFaultFact{CaseIndex: caseIndex, DeliveryIndex: deliveryIndex, RootActivityID: site.RootActivityID,
		ActivityID: site.ActivityID, ExpressionID: site.ExpressionID, Kind: fault.Kind, Operator: site.Operator,
		Assignment: site.Assignment, Variable: site.Variable, Left: fault.Left, Right: fault.Right,
		ProjectionSHA256: site.ProjectionSHA256, ProjectionStart: site.Start, ProjectionEnd: site.End,
		Expression: site.Expression, SourceMapping: "UNMAPPED", SourceStart: -1, SourceEnd: -1}
	if source := RecordedFaultLocalRHS(runtime, site); source != nil {
		fact.SourceMapping, fact.SourceRelation = source.Status, "selected_local_rhs"
		fact.SourceSHA256, fact.SourceStart, fact.SourceEnd, fact.SourceExpression = source.SourceSHA256, source.Start, source.End, source.Expression
	}
	return fact
}
