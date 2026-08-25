package artifactemit

const (
	OperationManifestKind   = "operation-manifest"
	OperationManifestSchema = "gooo/operation-manifest/v1"
)

type Artifact struct {
	Schema        string            `json:"schema"`
	Decision      string            `json:"decision"`
	Resolution    string            `json:"resolution"`
	Reason        string            `json:"reason"`
	Kind          string            `json:"kind"`
	SubjectDigest string            `json:"subject_digest,omitempty"`
	Package       Package           `json:"package"`
	Operation     Operation         `json:"operation"`
	Definitions   DefinitionSet     `json:"definitions"`
	Extensions    ExtensionRegistry `json:"extensions"`
	Effects       Effects           `json:"effects"`
	Digest        string            `json:"digest"`
}

type Package struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Operation struct {
	Activity string    `json:"activity"`
	Inputs   []Binding `json:"inputs"`
	Output   Binding   `json:"output"`
}

type Binding struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type DefinitionSet struct {
	Language string       `json:"language"`
	Files    []Definition `json:"files"`
}

type Definition struct {
	Filename         string `json:"filename"`
	Digest           string `json:"digest"`
	DeclarationCount int    `json:"declaration_count"`
}

type ExtensionRegistry struct {
	RegisteredEmitters int      `json:"registered_emitters"`
	Kinds              []string `json:"kinds"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}
