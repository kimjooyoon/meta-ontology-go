package promotioncontinuity

const (
	Schema         = "gooo/language-promotion-authorized-continuity/v1"
	guardSchema    = "gooo/autonomy-guarded-promotion/v1"
	recoverySchema = "gooo/language-rollback-fixed-point/v1"
)

type Input struct {
	ExpectedHeadSHA string
	GuardPath       string
	RecoveryPath    string
}

const (
	DecisionFailClosed = "FAIL_CLOSED"
	DecisionPass       = "PASS"
	ReasonMixed        = "NON_PROMOTING_TERMINAL_PRESERVED"
	ModeMixed          = "NON_PROMOTING_TERMINAL_PRESERVED"
	OperationMixed     = "preserve-non-promoting-terminal"
)

type GuardEvidence struct {
	Schema              string `json:"schema"`
	FileSHA256          string `json:"file_sha256"`
	ReportDigest        string `json:"report_digest"`
	HeadSHA             string `json:"head_sha"`
	Decision            string `json:"decision"`
	Reason              string `json:"reason"`
	Resolution          string `json:"resolution"`
	Satisfied           int    `json:"satisfied"`
	Total               int    `json:"total"`
	Unresolved          int    `json:"unresolved"`
	RepositoryWrites    int    `json:"repository_writes"`
	PromotionAuthorized bool   `json:"promotion_authorized"`
	MutationAuthorized  bool   `json:"mutation_authorized"`
}

type RecoveryEvidence struct {
	Schema                                      string `json:"schema"`
	FileSHA256                                  string `json:"file_sha256"`
	ReportDigest                                string `json:"report_digest"`
	HeadSHA                                     string `json:"head_sha"`
	Decision                                    string `json:"decision"`
	Reason                                      string `json:"reason"`
	Resolution                                  string `json:"resolution"`
	Mode                                        string `json:"mode"`
	GuardDecision                               string `json:"guard_decision"`
	GuardReason                                 string `json:"guard_reason"`
	GuardResolution                             string `json:"guard_resolution"`
	GuardSatisfied                              int    `json:"guard_satisfied"`
	GuardTotal                                  int    `json:"guard_total"`
	GuardUnresolved                             int    `json:"guard_unresolved"`
	GuardRepositoryWrites                       int    `json:"guard_repository_writes"`
	GuardMutationAuthorized                     bool   `json:"guard_mutation_authorized"`
	Satisfied                                   int    `json:"satisfied"`
	Total                                       int    `json:"total"`
	Unresolved                                  int    `json:"unresolved"`
	ReadinessBPS                                int    `json:"readiness_bps"`
	RecoveredFixedPoints                        int    `json:"recovered_fixed_points"`
	AuthorizedPromotions                        int    `json:"authorized_promotions"`
	TransformationDecision                      string `json:"transformation_decision"`
	TransformationReason                        string `json:"transformation_reason"`
	TransformationHeadSHA                       string `json:"transformation_head_sha"`
	TransformationWorkspaceMode                 string `json:"transformation_workspace_mode"`
	TransformationEffects                       int    `json:"transformation_effects"`
	TransformationAppliedEffects                int    `json:"transformation_applied_effects"`
	TransformationRefutedEffects                int    `json:"transformation_refuted_effects"`
	TransformationOperationOutcome              string `json:"transformation_operation_outcome"`
	TransformationReceiptDecision               string `json:"transformation_receipt_decision"`
	TransformationReceiptCount                  int    `json:"transformation_receipt_count"`
	TransformationFailureCount                  int    `json:"transformation_failure_count"`
	TransformationUnknownCount                  int    `json:"transformation_unknown_count"`
	TransformationDirectUnknownCount            int    `json:"transformation_direct_unknown_count"`
	TransformationDependencyBlockedUnknownCount int    `json:"transformation_dependency_blocked_unknown_count"`
	TransformationUnknownCausalDigest           string `json:"transformation_unknown_causal_digest"`
	WriteBoundary                               string `json:"write_boundary"`
	SourceWorkspaceUnchanged                    bool   `json:"source_workspace_unchanged"`
	TransformationAuthorization                 bool   `json:"transformation_authorization"`
	SourceRepositoryWrites                      int    `json:"source_repository_writes"`
	SummaryRepositoryWrites                     int    `json:"summary_repository_writes"`
	RepositoryWrites                            int    `json:"repository_writes"`
	MutationAuthorized                          bool   `json:"mutation_authorized"`
}

type Source struct {
	ExpectedHeadSHA string           `json:"expected_head_sha"`
	Guard           GuardEvidence    `json:"guard"`
	Recovery        RecoveryEvidence `json:"recovery"`
}
