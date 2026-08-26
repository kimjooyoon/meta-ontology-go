package artifactemit

const SymbolicInvocationConformanceSchema = "gooo/symbolic-invocation-conformance/v1"

type SymbolicInvocationConformance struct {
	Schema             string              `json:"schema"`
	Decision           string              `json:"decision"`
	Resolution         string              `json:"resolution"`
	Reason             string              `json:"reason"`
	GeneratedVectors   int                 `json:"generated_vectors"`
	HandwrittenVectors int                 `json:"handwritten_vectors"`
	Vectors            []ConformanceVector `json:"vectors"`
	Effects            Effects             `json:"effects"`
	NotClaimed         []string            `json:"not_claimed"`
}

type ConformanceVector struct {
	ID            string              `json:"id"`
	Expected      string              `json:"expected"`
	ProofChoice   string              `json:"proof_choice"`
	MetaOperation string              `json:"meta_operation"`
	Instance      ConformanceInstance `json:"instance"`
}

type ConformanceInstance struct {
	Activity string   `json:"activity,omitempty"`
	Inputs   []string `json:"inputs"`
}
