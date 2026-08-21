package metriccounterfactual

const (
	ManifestSchema = "gooo/metric-counterfactual-manifest/v1"
	PlanSchema     = "gooo/metric-counterfactual-plan/v1"
)

type FileSpec struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Content  string `json:"content"`
}

type Manifest struct {
	Schema string     `json:"schema"`
	Files  []FileSpec `json:"files"`
	Digest string     `json:"digest"`
}

type Mutation struct {
	Kind    string `json:"kind"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

type Plan struct {
	Schema    string     `json:"schema"`
	Mutations []Mutation `json:"mutations"`
	Digest    string     `json:"digest"`
}
