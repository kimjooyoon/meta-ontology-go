package workfrontier

// R4Result is the sealed finite-frontier decision. It intentionally contains
// neither proof_valid nor promotion_authorized.
type R4Result struct {
	SchemaVersion      string        `json:"schema_version"`
	Status             string        `json:"status"`
	Reason             string        `json:"reason"`
	Selected           []string      `json:"selected"`
	SelectedIDs        []string      `json:"selected_ids"`
	WorkIDs            []string      `json:"work_ids"`
	Unknown            []string      `json:"unknown"`
	Blocked            []string      `json:"blocked"`
	Shortfall          []string      `json:"shortfall"`
	Quality            string        `json:"quality"`
	FullSuiteRequired  bool          `json:"full_suite_required"`
	GraphDigest        string        `json:"graph_digest"`
	SCCDigest          string        `json:"scc_digest"`
	CondensationDigest string        `json:"condensation_digest"`
	RuleDigest         string        `json:"rule_digest"`
	WorkReceipt        R4WorkReceipt `json:"work_receipt"`
}

// R4SCC is a canonical graph component summary useful to callers that need
// to bind a rule before evaluating the envelope.
type R4SCC struct {
	Digest  string   `json:"digest"`
	Members []string `json:"members"`
	Cyclic  bool     `json:"cyclic"`
}

// R4GraphSummary exposes only source-derived graph facts and digests.
type R4GraphSummary struct {
	GraphDigest        string   `json:"graph_digest"`
	SCCDigest          string   `json:"scc_digest"`
	CondensationDigest string   `json:"condensation_digest"`
	SCCs               []R4SCC  `json:"sccs"`
	ReachableNodes     []string `json:"reachable_nodes"`
}
