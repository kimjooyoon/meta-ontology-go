package metricinterventionverify

const ReceiptSchema = "gooo/metric-intervention-verification/v1"

type Receipt struct {
	Schema                     string `json:"schema"`
	LedgerDigest               string `json:"ledger_digest"`
	SourceMetricsDigest        string `json:"source_metrics_digest"`
	CounterfactualReplayDigest string `json:"counterfactual_replay_digest"`
	ProjectionCount            int    `json:"projection_count"`
	IndicatorCount             int    `json:"indicator_count"`
	Status                     string `json:"status"`
	RepositoryWorkspaceWrites  bool   `json:"repository_workspace_writes"`
	PromotionAuthorized        bool   `json:"promotion_authorized"`
	Digest                     string `json:"digest"`
}
