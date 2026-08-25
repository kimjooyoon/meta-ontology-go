package artifactemit

const (
	OperationManifestKind        = "operation-manifest"
	OperationManifestSchema      = "gooo/operation-manifest/v1"
	SymbolicInvocationSchemaKind = "symbolic-invocation-schema"
	SymbolicInvocationArtifact   = "gooo/symbolic-invocation-schema-artifact/v1"
	SymbolicInvocationResolution = "SYMBOLIC_ONLY"
	JSONSchemaDraft202012        = "https://json-schema.org/draft/2020-12/schema"
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
	JSONSchema    *InvocationSchema `json:"json_schema,omitempty"`
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

type InvocationSchema struct {
	Dialect              string                     `json:"$schema"`
	Title                string                     `json:"title"`
	Type                 string                     `json:"type"`
	Properties           InvocationSchemaProperties `json:"properties"`
	Required             []string                   `json:"required"`
	AdditionalProperties bool                       `json:"additionalProperties"`
}

type InvocationSchemaProperties struct {
	Activity ConstSchema `json:"activity"`
	Inputs   TupleSchema `json:"inputs"`
}

type ConstSchema struct {
	Const string `json:"const"`
}

type TupleSchema struct {
	Type        string        `json:"type"`
	PrefixItems []ConstSchema `json:"prefixItems"`
	Items       bool          `json:"items"`
	MinItems    int           `json:"minItems"`
	MaxItems    int           `json:"maxItems"`
}

type DefinitionSet struct {
	Language string       `json:"language"`
	Files    []Definition `json:"files"`
}
