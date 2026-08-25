package artifactresolutionexperiment

type CounterexampleSummary struct {
	UnknownEmitterRejections int `json:"unknown_emitter_rejections"`
}

type EffectSummary struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Summary struct {
	Coordinates     Coordinates           `json:"coordinates"`
	Artifacts       ArtifactSummary       `json:"artifacts"`
	Resolution      ResolutionSummary     `json:"resolution"`
	Counterexamples CounterexampleSummary `json:"counterexamples"`
	Effects         EffectSummary         `json:"effects"`
	NotClaimed      int                   `json:"not_claimed"`
	Unknowns        int                   `json:"unknowns"`
}
