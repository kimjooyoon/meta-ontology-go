package artifactemit

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
