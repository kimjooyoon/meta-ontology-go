package main

type ledgerCounts struct {
	FileWitnesses                 int `json:"file_witnesses"`
	FunctionWitnesses             int `json:"function_witnesses"`
	GoFiles                       int `json:"go_files"`
	GoooFiles                     int `json:"gooo_files"`
	OtherFiles                    int `json:"other_files"`
	LogicalDirectories            int `json:"logical_directories"`
	StorageDirectories            int `json:"storage_directories"`
	FileSourceBindings            int `json:"file_source_bindings"`
	FunctionSourceBindings        int `json:"function_source_bindings"`
	StorageSourceBindings         int `json:"storage_source_bindings"`
	RootSummaryIndicators         int `json:"root_summary_indicators"`
	DerivedBindings               int `json:"derived_bindings"`
	SubjectWitnesses              int `json:"subject_witnesses"`
	MetaIndicators                int `json:"meta_indicators"`
	SourceIndicatorsApplicable    int `json:"source_indicators_applicable"`
	SourceIndicatorsNotApplicable int `json:"source_indicators_not_applicable"`
	WorkflowDiscoveryExemptions   int `json:"workflow_discovery_exemptions"`
}

type witnessLedger struct {
	Schema               string            `json:"schema"`
	Repository           string            `json:"repository"`
	CommitSHA            string            `json:"commit_sha"`
	SourceSchema         string            `json:"source_schema"`
	Policy               sourcePolicy      `json:"policy"`
	PolicyDigest         string            `json:"policy_digest"`
	RootTopologyExempt   bool              `json:"root_topology_exempt"`
	RootREADMEExempt     bool              `json:"root_readme_exempt"`
	Counts               ledgerCounts      `json:"counts"`
	SubjectWitnessDigest string            `json:"subject_witness_digest"`
	MetaIndicatorDigest  string            `json:"meta_indicator_digest"`
	IndicatorDigest      string            `json:"indicator_digest"`
	SemanticDigest       string            `json:"semantic_digest"`
	Status               string            `json:"status"`
	ObservationState     string            `json:"observation_state,omitempty"`
	ClaimState           string            `json:"claim_state,omitempty"`
	SourceObservations   []sourceIndicator `json:"source_observations,omitempty"`
	SourceObservationDigest string          `json:"source_observation_digest,omitempty"`
	Indicators           []ledgerIndicator `json:"indicators"`
	Witnesses            []subjectWitness  `json:"witnesses"`
}
