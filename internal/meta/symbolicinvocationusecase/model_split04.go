package symbolicinvocationusecase

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type facts struct {
	UserDecisions        int
	AcceptedInstances    int
	RejectedInstances    int
	GeneratedInstances   int
	GoldenMatches        int
	DeterministicReplays int
	Unknowns             int
	Source               SourceCoordinate
	Producer             ProducerBinding
	Resources            ResourceObservation
	Effects              Effects
}
