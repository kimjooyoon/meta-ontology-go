package languageassurance

type RoleBinding struct {
	Principal string `json:"principal"`
	Roles     []Role `json:"roles"`
}

type DecisionTransition struct {
	ID     string   `json:"id"`
	Input  Decision `json:"input"`
	Output Decision `json:"output"`
}

type Transaction struct {
	Schema              string               `json:"schema"`
	TransactionID       string               `json:"transaction_id"`
	AuthorityRoutes     []AuthorityRoute     `json:"authority_routes"`
	RoleBindings        []RoleBinding        `json:"role_bindings"`
	DecisionTransitions []DecisionTransition `json:"decision_transitions"`
}

type ObligationDefinition struct {
	MetricID              string         `json:"metric_id"`
	Priority              Priority       `json:"priority"`
	Class                 IndicatorClass `json:"class"`
	ProofChoice           ProofChoice    `json:"proof_choice"`
	RequiredMetaOperation string         `json:"required_meta_operation"`
}

type MetaOperation struct {
	ID          string      `json:"id"`
	Activity    string      `json:"activity"`
	ProofChoice ProofChoice `json:"proof_choice"`
}

type ObligationObservation struct {
	MetricID      string     `json:"metric_id"`
	Status        string     `json:"status"`
	Resolution    Resolution `json:"resolution"`
	MetaOperation string     `json:"meta_operation,omitempty"`
}

type RolePair struct {
	Left  Role `json:"left"`
	Right Role `json:"right"`
}

type Finding struct {
	MetricID   string   `json:"metric_id"`
	PathID     string   `json:"path_id"`
	Principal  string   `json:"principal,omitempty"`
	RuleID     string   `json:"rule_id,omitempty"`
	Roles      []Role   `json:"roles,omitempty"`
	DecisionID string   `json:"decision_id,omitempty"`
	Input      Decision `json:"input,omitempty"`
	Output     Decision `json:"output,omitempty"`
}

type Indicator struct {
	MetricID      string         `json:"metric_id"`
	Class         IndicatorClass `json:"class"`
	ProofChoice   ProofChoice    `json:"proof_choice"`
	Producer      string         `json:"producer"`
	Consumer      string         `json:"consumer"`
	MetaOperation string         `json:"meta_operation"`
	Value         *int           `json:"value"`
	Target        int            `json:"target"`
	Unit          string         `json:"unit"`
	Relation      Relation       `json:"relation"`
	Resolution    Resolution     `json:"resolution"`
	Satisfied     bool           `json:"satisfied"`
}
