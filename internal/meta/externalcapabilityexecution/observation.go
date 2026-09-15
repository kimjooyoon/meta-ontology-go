package externalcapabilityexecution

type Observation struct {
	Schema                   string          `json:"schema"`
	SubjectSHA               string          `json:"subject_sha"`
	Available                bool            `json:"available"`
	Reference                Reference       `json:"reference"`
	Parent                   ParentReport    `json:"parent"`
	Runs                     []CapabilityRun `json:"runs"`
	ReplayExact              bool            `json:"replay_exact"`
	RepositoryWrites         int             `json:"repository_writes"`
	ExternalRepositoryWrites int             `json:"external_repository_writes"`
	ExternalExecutions       int             `json:"external_executions"`
	OfficialMutationCount    int             `json:"official_mutation_count"`
	PromotionCount           int             `json:"promotion_count"`
	UnknownEvents            []string        `json:"unknown_events"`
	ObservationDigest        string          `json:"observation_digest"`
}

type ObserverOptions struct {
	SubjectSHA   string
	SourceRoot   string
	ExternalRoot string
}
