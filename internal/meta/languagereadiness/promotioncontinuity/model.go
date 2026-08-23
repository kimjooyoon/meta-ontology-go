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
	Schema                      string `json:"schema"`
	FileSHA256                  string `json:"file_sha256"`
	ReportDigest                string `json:"report_digest"`
	HeadSHA                     string `json:"head_sha"`
	Decision                    string `json:"decision"`
	Reason                      string `json:"reason"`
	Resolution                  string `json:"resolution"`
	Mode                        string `json:"mode"`
	GuardDecision               string `json:"guard_decision"`
	GuardResolution             string `json:"guard_resolution"`
	Satisfied                   int    `json:"satisfied"`
	Total                       int    `json:"total"`
	Unresolved                  int    `json:"unresolved"`
	ReadinessBPS                int    `json:"readiness_bps"`
	RecoveredFixedPoints        int    `json:"recovered_fixed_points"`
	AuthorizedPromotions        int    `json:"authorized_promotions"`
	TransformationDecision      string `json:"transformation_decision"`
	TransformationEffects       int    `json:"transformation_effects"`
	WriteBoundary               string `json:"write_boundary"`
	SourceWorkspaceUnchanged    bool   `json:"source_workspace_unchanged"`
	TransformationAuthorization bool   `json:"transformation_authorization"`
	SourceRepositoryWrites      int    `json:"source_repository_writes"`
	SummaryRepositoryWrites     int    `json:"summary_repository_writes"`
	RepositoryWrites            int    `json:"repository_writes"`
	MutationAuthorized          bool   `json:"mutation_authorized"`
}

type Source struct {
	ExpectedHeadSHA string           `json:"expected_head_sha"`
	Guard           GuardEvidence    `json:"guard"`
	Recovery        RecoveryEvidence `json:"recovery"`
}
