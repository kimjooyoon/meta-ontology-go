package selfimprovementobservation

type ContractReport struct {
	Schema              string              `json:"schema"`
	CommitSHA           string              `json:"commit_sha"`
	SemanticHash        string              `json:"semantic_hash"`
	RegistryDigest      string              `json:"registry_digest"`
	Status              string              `json:"status"`
	Indicators          []ContractIndicator `json:"indicators"`
	PromotionAuthorized bool                `json:"promotion_authorized"`
}

type ContractIndicator struct {
	ID      string `json:"id"`
	Route   string `json:"route"`
	Verdict string `json:"verdict"`
}
