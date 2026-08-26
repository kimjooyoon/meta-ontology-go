package metriccounterfactual

import artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"

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

func SealManifest(value Manifest) (Manifest, error) {
	value.Digest = ""
	digest, err := artifact.Digest(value)
	value.Digest = digest
	return value, err
}

func ValidManifest(value Manifest) bool {
	digest := value.Digest
	sealed, err := SealManifest(value)
	return err == nil && digest == sealed.Digest
}

func SealPlan(value Plan) (Plan, error) {
	value.Digest = ""
	digest, err := artifact.Digest(value)
	value.Digest = digest
	return value, err
}

func ValidPlan(value Plan) bool {
	digest := value.Digest
	sealed, err := SealPlan(value)
	return err == nil && digest == sealed.Digest
}
