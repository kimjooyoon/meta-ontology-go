package unresolvedpolicy

// SourceDisposition is the native transport shape. Gooo owns all field values;
// tags bind display-independent source field identities to this boundary.
type SourceDisposition struct {
	Resolution                string `json:"resolution" gooo:"gooo://metric-operation-origin/disposition-record/resolution"`
	Reason                    string `json:"reason" gooo:"gooo://metric-operation-origin/disposition-record/reason"`
	NextOperation             string `json:"next_operation" gooo:"gooo://metric-operation-origin/disposition-record/next-operation"`
	ReadinessCredit           bool   `json:"readiness_credit" gooo:"gooo://metric-operation-origin/disposition-record/readiness-credit"`
	RepositoryWorkspaceWrites bool   `json:"repository_workspace_writes" gooo:"gooo://metric-operation-origin/disposition-record/repository-workspace-writes"`
	PromotionAuthorized       bool   `json:"promotion_authorized" gooo:"gooo://metric-operation-origin/disposition-record/promotion-authorized"`
}
