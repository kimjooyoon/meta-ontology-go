package predecessorbinding

const BindingTransitionSchema = "gooo/language-readiness-predecessor-binding-transition/v1"

type TransitionIndicator struct {
	ID     string `json:"id"`
	Class  string `json:"class"`
	Before int    `json:"before"`
	After  int    `json:"after"`
	Delta  int    `json:"delta"`
	Unit   string `json:"unit"`
}

type BindingTransition struct {
	Schema           string                `json:"schema"`
	Decision         string                `json:"decision"`
	Reason           string                `json:"reason"`
	Comparable       bool                  `json:"comparable"`
	RegistryDigest   string                `json:"registry_digest"`
	BeforeHeadSHA    string                `json:"before_head_sha"`
	AfterHeadSHA     string                `json:"after_head_sha"`
	Total            int                   `json:"total"`
	BeforeStatic     int                   `json:"before_static"`
	AfterStatic      int                   `json:"after_static"`
	StaticDelta      int                   `json:"static_delta"`
	BeforeDynamic    int                   `json:"before_dynamic"`
	AfterDynamic     int                   `json:"after_dynamic"`
	DynamicDelta     int                   `json:"dynamic_delta"`
	BeforeBPS        int                   `json:"before_bps"`
	AfterBPS         int                   `json:"after_bps"`
	BPSDelta         int                   `json:"bps_delta"`
	Unknown          int                   `json:"unknown"`
	RepositoryWrites int                   `json:"repository_writes"`
	Indicators       []TransitionIndicator `json:"indicators"`
	Proofs           []Proof               `json:"proofs"`
	Digest           string                `json:"digest"`
}
