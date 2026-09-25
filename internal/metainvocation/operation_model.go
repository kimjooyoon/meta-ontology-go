package metainvocation

type OperationSpec struct {
	ID                string `json:"id"`
	InputEntity       string `json:"input_entity"`
	OutputEntity      string `json:"output_entity"`
	Phase             string `json:"phase"`
	RepositoryWrites  int    `json:"repository_writes"`
	MutationAuthority bool   `json:"mutation_authority"`
}

type BoundOperation struct {
	Activity   string           `json:"activity"`
	Program    string           `json:"program"`
	Input      string           `json:"input"`
	Output     string           `json:"output"`
	SpecDigest string           `json:"spec_digest"`
	Source     SourceCoordinate `json:"source"`
}

type Program struct {
	SourcePath   string                    `json:"source_path"`
	Package      string                    `json:"package"`
	Namespace    string                    `json:"namespace"`
	SourceDigest string                    `json:"source_digest"`
	Entities     map[string]string         `json:"entities"`
	Operations   map[string]BoundOperation `json:"operations"`
}
