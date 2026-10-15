package bodyexecution

// Primitive fields can be passed directly to a Gooo record. The binding and
// field IDs describe source identity; frame/assignment/completion identify an
// actual instance. Projection ranges are explicitly separate from Gooo literals.
type CompositionFieldFact struct {
	CaseIndex            int    `json:"case_index"`
	EventIndex           int    `json:"event_index"`
	AssignmentIndex      int    `json:"assignment_index"`
	Frame                int    `json:"frame"`
	VariableOrdinal      int    `json:"variable_ordinal"`
	BindingIndex         int    `json:"binding_index"`
	UseStart             int    `json:"use_start"`
	UseEnd               int    `json:"use_end"`
	UsesDropped          int    `json:"uses_dropped"`
	Completion           int    `json:"completion"`
	ParentAssignment     int    `json:"parent_assignment"`
	Completed            bool   `json:"completed"`
	HasAssignment        bool   `json:"has_assignment"`
	EventKind            string `json:"event_kind"`
	RootActivityID       string `json:"root_activity_id"`
	ActivityID           string `json:"activity_id"`
	VariableName         string `json:"variable_name"`
	RecordID             string `json:"record_id"`
	FieldID              string `json:"field_id"`
	FieldName            string `json:"field_name"`
	ScalarTypeID         string `json:"scalar_type_id"`
	BindingKind          string `json:"binding_kind"`
	InputPort            string `json:"input_port"`
	InputFrom            string `json:"input_from_activity_id"`
	SiteKind             string `json:"site_kind"`
	DependencyMode       string `json:"dependency_mode"`
	Operation            string `json:"operation,omitempty"`
	AssignmentKind       string `json:"assignment_kind"`
	Reason               string `json:"reason"`
	ProjectionSHA256     string `json:"projection_sha256"`
	ProjectionStart      int    `json:"projection_start"`
	ProjectionEnd        int    `json:"projection_end"`
	ProjectionExpression string `json:"projection_expression"`
}

func isFieldOriginKind(kind string) bool {
	return kind == "field-writes" || kind == "field-uses" || kind == "field-input-locations"
}

func fieldFact(runtime CompositionRuntime, site CompositionFieldSite, c, event, frame int) CompositionFieldFact {
	binding := runtime.FieldBindings[site.Binding]
	field := binding.Record.Fields[site.Field]
	kind := binding.InputKind
	if kind == "" {
		kind = "local"
	}
	return CompositionFieldFact{CaseIndex: c, EventIndex: event, Frame: frame, VariableOrdinal: binding.VariableOrdinal, BindingIndex: site.Binding,
		RootActivityID: binding.RootActivityID, ActivityID: binding.ActivityID, VariableName: binding.Name,
		RecordID: binding.Record.ID, FieldID: field.ID, FieldName: field.Name, ScalarTypeID: field.TypeID,
		BindingKind: kind, InputPort: binding.InputPort, InputFrom: binding.InputFrom, SiteKind: site.Kind, DependencyMode: site.DependencyMode, Operation: site.Operation,
		ProjectionSHA256: runtime.FieldProjectionSHA256, ProjectionStart: site.Start, ProjectionEnd: site.End, ProjectionExpression: site.Expression,
		ParentAssignment: -1, AssignmentIndex: -1, Completion: -1}
}

func (page *nativeOriginPage) visitFields(runtime CompositionRuntime, trace CompositionTrace) {
	if page.query.Kind == "field-writes" {
		for i, write := range trace.FieldWrites {
			fact := fieldFact(runtime, runtime.FieldSites[write.SiteIndex], trace.CaseIndex, i, write.Frame)
			fact.EventKind, fact.AssignmentIndex, fact.HasAssignment = "write", i, write.Completed
			fillFieldFactWrite(&fact, write, runtime.FieldSites[write.SiteIndex])
			page.addField(fact, fact.RootActivityID, fact.ActivityID, fact.FieldID)
		}
		return
	}
	for i, use := range trace.FieldUses {
		fact := fieldFact(runtime, runtime.FieldUseSites[use.SiteIndex], trace.CaseIndex, i, use.Frame)
		fact.EventKind, fact.AssignmentIndex, fact.HasAssignment, fact.Reason = "use", use.Assignment, use.Assignment >= 0, use.Reason
		if use.Assignment >= 0 {
			write := trace.FieldWrites[use.Assignment]
			fillFieldFactWrite(&fact, write, runtime.FieldSites[write.SiteIndex])
		}
		page.addField(fact, fact.RootActivityID, fact.ActivityID, fact.FieldID)
	}
}

func fillFieldFactWrite(fact *CompositionFieldFact, write CompositionFieldWrite, site CompositionFieldSite) {
	fact.UseStart, fact.UseEnd, fact.UsesDropped = write.UseStart, write.UseEnd, write.UsesDropped
	fact.Completed, fact.Completion, fact.ParentAssignment = write.Completed, write.Completion, write.ParentAssignment
	fact.AssignmentKind, fact.DependencyMode = site.Kind, site.DependencyMode
	fact.Operation = site.Operation
}
