package artifactemit

type symbolicValueArtifactInput struct {
	Schema      string `json:"schema"`
	Decision    string `json:"decision"`
	Digest      string `json:"digest"`
	Conformance struct {
		Schema                     string                     `json:"schema"`
		Decision                   string                     `json:"decision"`
		Resolution                 string                     `json:"resolution"`
		GeneratedVectors           int                        `json:"generated_vectors"`
		EmbeddedHandwrittenVectors int                        `json:"embedded_handwritten_vectors"`
		Vectors                    []symbolicValueVectorInput `json:"vectors"`
		Effects                    struct {
			RepositoryWrites  int  `json:"repository_writes"`
			MutationAuthority bool `json:"mutation_authority"`
		} `json:"effects"`
	} `json:"conformance"`
}

type symbolicValueVectorInput struct {
	ID            string `json:"id"`
	Expected      string `json:"expected"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Instance      struct {
		Activity string   `json:"activity"`
		Inputs   []string `json:"inputs"`
	} `json:"instance"`
}
