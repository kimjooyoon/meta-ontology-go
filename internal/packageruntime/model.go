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
	Path           string        `json:"path"`
	Name           string        `json:"name"`
	Namespace      string        `json:"namespace"`
	Imports        []string      `json:"imports"`
	Sources        []SourceImage `json:"sources"`
	Declarations   int           `json:"declarations"`
	SemanticDigest string        `json:"semantic_digest"`
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
