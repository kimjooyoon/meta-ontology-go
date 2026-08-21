package metricstrategyverify

const ReceiptSchema = "gooo/metric-meta-strategy-verification/v1"

type Receipt struct {
	Schema                    string `json:"schema"`
	PlanDigest                string `json:"plan_digest"`
	SourceMetricsDigest       string `json:"source_metrics_digest"`
	InterventionDigest        string `json:"intervention_digest"`
	BindingCount              int    `json:"binding_count"`
	CandidateCount            int    `json:"candidate_count"`
	SelectedProofChoice       string `json:"selected_proof_choice"`
	Status                    string `json:"status"`
	RepositoryWorkspaceWrites bool   `json:"repository_workspace_writes"`
	PromotionAuthorized       bool   `json:"promotion_authorized"`
	Digest                    string `json:"digest"`
}

