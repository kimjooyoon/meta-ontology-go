package formatfix

const PlanSchema = "gooo/format-fix-plan/v1"

type Decision string
type Resolution string

const (
	DecisionChangePlanned Decision = "CHANGE_PLANNED"
	DecisionFixedPoint    Decision = "FIXED_POINT"
	DecisionFailClosed    Decision = "FAIL_CLOSED"

	ResolutionExact Resolution = "EXACT"
	ResolutionLower Resolution = "LOWER_RESOLUTION"
)

type Edit struct {
	Start       int    `json:"start"`
	End         int    `json:"end"`
	Replacement string `json:"replacement"`
}

type Plan struct {
	Schema             string     `json:"schema"`
	File               string     `json:"file"`
	Decision           Decision   `json:"decision"`
	Resolution         Resolution `json:"resolution"`
	ReasonCode         string     `json:"reason_code"`
	SourceDigest       string     `json:"source_digest"`
	ResultDigest       string     `json:"result_digest"`
	SourceBytes        int        `json:"source_bytes"`
	ResultBytes        int        `json:"result_bytes"`
	SemanticBefore     string     `json:"semantic_before"`
	SemanticAfter      string     `json:"semantic_after"`
	SemanticEqual      bool       `json:"semantic_equal"`
	Changed            bool       `json:"changed"`
	Edits              []Edit     `json:"edits"`
	Diagnostics        []string   `json:"diagnostics"`
	DirectWrites       int        `json:"direct_writes"`
	MutationAuthorized bool       `json:"mutation_authorized"`
	PlanDigest         string     `json:"plan_digest"`
}
