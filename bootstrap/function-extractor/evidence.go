package main

type splitPlan struct {
	Schema    string        `json:"schema"`
	SourceSHA string        `json:"source_sha"`
	Subjects  []planSubject `json:"subjects"`
}

type planSubject struct {
	Logical string `json:"logical"`
	Lines   int    `json:"lines"`
	Reason  string `json:"reason"`
}

type densityReport struct {
	Schema    string           `json:"schema"`
	SourceSHA string           `json:"source_sha"`
	Subjects  []densitySubject `json:"subjects"`
}

type densitySubject struct {
	Logical string `json:"logical"`
	Status  string `json:"status"`
}

type extractionSubject struct {
	Logical      string   `json:"logical"`
	State        string   `json:"state"`
	Before       int      `json:"before_lines"`
	After        int      `json:"after_lines"`
	Files        []string `json:"changed_files"`
	CreatedFiles []string `json:"created_files,omitempty"`
	Consumer     string   `json:"consumer"`
	Operation    string   `json:"meta_operation"`
	Operations   []string `json:"meta_operations,omitempty"`
	Proof        string   `json:"proof_choice"`
}

type extractionFailureRecord struct {
	Logical       string   `json:"logical"`
	BlockerID     string   `json:"blocker_id,omitempty"`
	Decision      string   `json:"decision"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class,omitempty"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
	Diagnostics   []string `json:"diagnostics,omitempty"`
}

type extractionIndicator struct {
	ID        string `json:"id"`
	Value     int    `json:"value"`
	Limit     int    `json:"limit"`
	Blocking  bool   `json:"blocking"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
	Proof     string `json:"proof_choice"`
}

type namespaceReplacementReceipt struct {
	LogicalPath           string `json:"logical_path"`
	Primitive             string `json:"primitive"`
	Contract              string `json:"contract"`
	GOOS                  string `json:"goos"`
	SameDirectory         bool   `json:"same_directory"`
	DestinationPreexisted bool   `json:"destination_preexisted"`
	TempDigest            string `json:"temp_digest"`
	ReplacementSuccess    bool   `json:"replacement_success"`
	FinalDigest           string `json:"final_digest"`
}

type backupCleanupObservation struct {
	Status    string `json:"status"`
	Attempted int    `json:"attempted"`
	Removed   int    `json:"removed"`
	Failures  int    `json:"failures"`
}

type extractionReport struct {
	Schema                string                        `json:"schema"`
	SourceSHA             string                        `json:"source_sha"`
	StagedSubjects        int                           `json:"staged_subjects"`
	Subjects              []extractionSubject           `json:"subjects"`
	Unhandled             []string                      `json:"unhandled"`
	Failures              []extractionFailureRecord     `json:"failures,omitempty"`
	Indicators            []extractionIndicator         `json:"indicators"`
	NamespaceReplacements []namespaceReplacementReceipt `json:"namespace_replacements,omitempty"`
	BackupCleanup         backupCleanupObservation      `json:"backup_cleanup"`
}
