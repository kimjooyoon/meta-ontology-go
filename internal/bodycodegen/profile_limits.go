package bodycodegen

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

// ProfileLimits describes independent bounds in the pure activity body profile.
// These values do not establish syntax support, execution time or a total
// instruction bound. Call expansion and syntactic loop bounds have distinct scopes.
type ProfileLimits struct {
	RecordReferenceDepth                int `json:"max_record_reference_depth"`
	NestedRecordValueNodes              int `json:"max_nested_record_value_nodes"`
	RecordTypes                         int `json:"max_record_types"`
	FieldsPerRecord                     int `json:"max_fields_per_record"`
	ElementsPerCollection               int `json:"max_elements_per_collection"`
	ActivityCallsPerInvocation          int `json:"max_activity_calls_per_invocation"`
	ActivitiesPerCallClosure            int `json:"max_activities_per_call_closure"`
	ActivityCallDepth                   int `json:"max_activity_call_depth"`
	IterationsPerCountedLoop            int `json:"max_iterations_per_counted_loop"`
	SyntacticCollectionIterationProduct int `json:"max_syntactic_collection_iteration_product"`
}

// CurrentProfileLimits returns the same constants used by source validation.
// It reads no source files, model, saved report, environment or runtime sample.
func CurrentProfileLimits() ProfileLimits {
	return ProfileLimits{
		RecordReferenceDepth: syntax.RecordReferenceDepthLimit, NestedRecordValueNodes: syntax.RecordValueSlotLimit,
		RecordTypes: recordTypeLimit, FieldsPerRecord: RecordFieldLimit,
		ElementsPerCollection: RepeatedFieldLimit, ActivityCallsPerInvocation: pureCallLimit,
		ActivitiesPerCallClosure: pureActivityLimit, ActivityCallDepth: pureCallDepthLimit,
		IterationsPerCountedLoop:            countedLoopLimit,
		SyntacticCollectionIterationProduct: CollectionIterationProductLimit,
	}
}
