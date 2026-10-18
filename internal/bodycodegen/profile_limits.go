package bodycodegen

// ProfileLimits describes independent bounds in the pure activity body profile.
// These values do not establish syntax support, execution time or a total
// instruction bound. Call expansion and syntactic loop bounds have distinct scopes.
type ProfileLimits struct {
	RecordTypes                         int `json:"record_types"`
	FieldsPerRecord                     int `json:"fields_per_record"`
	ElementsPerCollection               int `json:"elements_per_collection"`
	ActivityCallsPerInvocation          int `json:"activity_calls_per_invocation"`
	IterationsPerCountedLoop            int `json:"iterations_per_counted_loop"`
	SyntacticCollectionIterationProduct int `json:"syntactic_collection_iteration_product"`
}

// CurrentProfileLimits returns the same constants used by source validation.
// It reads no source files, model, saved report, environment or runtime sample.
func CurrentProfileLimits() ProfileLimits {
	return ProfileLimits{
		RecordTypes: recordTypeLimit, FieldsPerRecord: RecordFieldLimit,
		ElementsPerCollection: RepeatedFieldLimit, ActivityCallsPerInvocation: pureCallLimit,
		IterationsPerCountedLoop:            countedLoopLimit,
		SyntacticCollectionIterationProduct: CollectionIterationProductLimit,
	}
}
