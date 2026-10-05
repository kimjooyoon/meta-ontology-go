package packageruntime

const (
	ManifestSchema = "gooo/package-runtime-manifest/v1"
	ImageSchema    = "gooo/package-runtime-image/v1"
	ResultSchema   = "gooo/package-runtime-result/v1"
)

type Manifest struct {
	Schema             string        `json:"schema"`
	Entry              EntrySpec     `json:"entry"`
	Packages           []PackageSpec `json:"packages"`
	MutationAuthorized bool          `json:"mutation_authorized"`
}

type EntrySpec struct {
	PackagePath string `json:"package_path"`
	Activity    string `json:"activity"`
}

type PackageSpec struct {
	Path    string   `json:"path"`
	Name    string   `json:"name"`
	Imports []string `json:"imports"`
	Sources []Source `json:"sources"`
}

type Source struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type SourceImage struct {
	Filename       string `json:"filename"`
	SourceDigest   string `json:"source_digest"`
	SemanticDigest string `json:"semantic_digest"`
	Declarations   int    `json:"declarations"`
}

type PackageImage struct {
	Path           string           `json:"path"`
	Name           string           `json:"name"`
	Namespace      string           `json:"namespace"`
	Imports        []string         `json:"imports"`
	Exports        []Export         `json:"exports"`
	Bindings       []PackageBinding `json:"bindings,omitempty"`
	Sources        []SourceImage    `json:"sources"`
	Declarations   int              `json:"declarations"`
	SemanticDigest string           `json:"semantic_digest"`
}

// PackageBinding is a checked edge from an imported activity output to a
// local activity input. It describes the package-level execution graph; it
// does not claim that generated code has executed.
type PackageBinding struct {
	ProducerPackage  string `json:"producer_package"`
	ProducerActivity string `json:"producer_activity"`
	ProducerPort     string `json:"producer_port"`
	ConsumerPackage  string `json:"consumer_package"`
	ConsumerActivity string `json:"consumer_activity"`
	ConsumerPort     string `json:"consumer_port"`
	EntityID         string `json:"entity_id"`
	Feedback         bool   `json:"feedback,omitempty"`
}

// Export is the statically resolved public surface of a Gooo package.
// Activity input and output types use canonical package-path-qualified names.
type Export struct {
	Name       string   `json:"name"`
	Kind       string   `json:"kind"`
	ID         string   `json:"id,omitempty"`
	InputTypes []string `json:"input_types,omitempty"`
	OutputType string   `json:"output_type,omitempty"`
}

type EntryPlan struct {
	PackagePath string   `json:"package_path"`
	Source      string   `json:"source"`
	Activity    string   `json:"activity"`
	Inputs      []string `json:"inputs"`
	Output      string   `json:"output"`
}

type Image struct {
	Schema    string         `json:"schema"`
	InitOrder []string       `json:"init_order"`
	Packages  []PackageImage `json:"packages"`
	Entry     EntryPlan      `json:"entry"`
	Digest    string         `json:"digest"`
}
