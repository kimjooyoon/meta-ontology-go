package adapter

// Fact is a compact semantic relation record.
type Fact struct {
	SubjectID string `json:"subject_id"`
	Predicate string `json:"predicate"`
	ObjectID  string `json:"object_id"`
	Class     string `json:"fact_class"`
	SourceURI string `json:"source_uri,omitempty"`
	Start     int    `json:"start_offset,omitempty"`
	End       int    `json:"end_offset,omitempty"`
}

// Conflict describes a rejected or unresolved semantic change.
type Conflict struct {
	Code       string `json:"code"`
	SemanticID string `json:"semantic_id,omitempty"`
	Detail     string `json:"detail,omitempty"`
}

// Measurements are correctness counts, not performance claims.
type Measurements struct {
	SourceSpanCount        int  `json:"source_span_count"`
	ProtectedBytesEqual    bool `json:"protected_bytes_equal"`
	UnrelatedRegionCount   int  `json:"unrelated_region_count"`
	RepeatCount            int  `json:"repeat_count"`
	CanonicalEqualCount    int  `json:"canonical_equal_count"`
	SourceEqualCount       int  `json:"source_equal_count"`
	SemanticEqualCount     int  `json:"semantic_equal_count"`
	RegionEqualCount       int  `json:"region_equal_count"`
	SourceMapResolvedCount int  `json:"source_map_resolved_count"`
	FalseAcceptanceCount   int  `json:"false_acceptance_count"`
	EnvironmentLeakCount   int  `json:"environment_leak_count"`
}
