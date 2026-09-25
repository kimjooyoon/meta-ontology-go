package metainvocation

type Claim struct {
	ID        string   `json:"id"`
	Statement string   `json:"statement"`
	Status    string   `json:"status"`
	Stage     string   `json:"stage"`
	Step      string   `json:"step"`
	Reason    string   `json:"reason"`
	Evidence  []string `json:"evidence"`
	DependsOn []string `json:"depends_on"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}
