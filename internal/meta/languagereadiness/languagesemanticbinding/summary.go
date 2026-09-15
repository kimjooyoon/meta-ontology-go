package languagesemanticbinding

type Summary struct {
	Coordinates         int `json:"coordinates"`
	BoundCoordinates    int `json:"bound_coordinates"`
	Unresolved          int `json:"unresolved"`
	ReadinessCompleted  int `json:"readiness_completed"`
	ReadinessTotal      int `json:"readiness_total"`
	ReadinessBPS        int `json:"readiness_bps"`
	SemanticSatisfied   int `json:"semantic_satisfied"`
	SemanticTotal       int `json:"semantic_total"`
	Concepts            int `json:"concepts"`
	MetricBindings      int `json:"metric_bindings"`
	Guardrails          int `json:"guardrails"`
	EffectfulStages     int `json:"effectful_stages"`
	RepositoryWrites    int `json:"repository_writes"`
	MutationAuthorities int `json:"mutation_authorities"`
}
