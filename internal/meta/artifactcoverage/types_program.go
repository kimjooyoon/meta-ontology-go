package artifactcoverage

type MetaOperation struct {
	ID          string      `json:"id"`
	Activity    string      `json:"activity"`
	ProofChoice ProofChoice `json:"proof_choice"`
}

type Program struct {
	Schema           string            `json:"schema"`
	AuthorityPath    string            `json:"authority_path"`
	MetaOperations   []MetaOperation   `json:"meta_operations"`
	Indicators       []Indicator       `json:"indicators"`
	ArtifactBindings []ArtifactBinding `json:"artifact_bindings"`
}

func CanonicalProgram() Program {
	return Program{
		Schema: Schema, AuthorityPath: AuthorityPath,
		MetaOperations: CanonicalOperations(), Indicators: CanonicalIndicators(),
		ArtifactBindings: CanonicalBindings(),
	}
}
