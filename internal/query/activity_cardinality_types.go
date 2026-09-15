package query

const ActivityCardinalityResolutionSchema = "gooo/activity-cardinality-resolution/v1"

type ActivityResolutionDecision string

const (
	ActivityResolutionUnknown ActivityResolutionDecision = "UNKNOWN"
	ActivityResolutionClosed  ActivityResolutionDecision = "CLOSED"
	ActivityResolutionRefuted ActivityResolutionDecision = "REFUTED"
)

type ActivitySelector struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	IDPrefix  string `json:"id_prefix,omitempty"`
}

type ActivityResolutionSubject struct {
	GraphHash      string `json:"graph_hash"`
	SemanticDigest string `json:"semantic_digest,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	SourceDigest   string `json:"source_digest,omitempty"`
	SourceStatus   string `json:"source_status"`
	SourceFile     string `json:"source_file,omitempty"`
}

type ActivityResolutionMatch struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type ActivityResolutionClaim struct {
	State         ActivityResolutionDecision `json:"state"`
	Stage         string                     `json:"stage"`
	Step          string                     `json:"step"`
	Reason        string                     `json:"reason"`
	UnknownClass  string                     `json:"unknown_class,omitempty"`
	NextOperation string                     `json:"next_operation"`
	ProofChoice   string                     `json:"proof_choice"`
	BlockedBy     []string                   `json:"blocked_by"`
}

type ActivityCardinalityResolution struct {
	Schema      string                     `json:"schema"`
	Decision    ActivityResolutionDecision `json:"decision"`
	Selector    ActivitySelector           `json:"selector"`
	Occurrences int                        `json:"occurrences"`
	Matches     []ActivityResolutionMatch  `json:"matches"`
	Subject     ActivityResolutionSubject  `json:"subject"`
	Claim       ActivityResolutionClaim    `json:"claim"`
}
