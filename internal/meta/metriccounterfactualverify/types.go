package metriccounterfactualverify

const Schema = "gooo/metric-counterfactual-verification/v1"

type Receipt struct {
	Schema               string `json:"schema"`
	LedgerDigest         string `json:"ledger_digest"`
	ReplayDigest         string `json:"replay_digest"`
	IndicatorCount       int    `json:"indicator_count"`
	Status               string `json:"status"`
	PromotionAuthorized  bool   `json:"promotion_authorized"`
	Digest               string `json:"digest"`
}
