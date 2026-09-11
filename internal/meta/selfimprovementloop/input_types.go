package selfimprovementloop

type Input struct {
	Schema          string                 `json:"schema"`
	Scenario        string                 `json:"scenario"`
	SourceDigest    string                 `json:"source_digest"`
	ToolchainDigest string                 `json:"toolchain_digest"`
	Baseline        BaselineObservation    `json:"baseline"`
	Target          TargetDeclaration      `json:"target"`
	Scope           ScopePin               `json:"scope"`
	Transformation  TransformationProposal `json:"transformation"`
	Prediction      EffectPrediction       `json:"prediction"`
	Counterexample  CounterexampleResult   `json:"counterexample"`
	CI              CIResult               `json:"ci"`
	Receipt         ReceiptInput           `json:"receipt"`
	Pair            ExactPair              `json:"pair"`
	Human           HumanDecision          `json:"human"`
}

type BaselineObservation struct {
	Present bool   `json:"present"`
	Metric  string `json:"metric"`
	Value   int64  `json:"value"`
}

type TargetDeclaration struct {
	Present bool   `json:"present"`
	Metric  string `json:"metric"`
	Value   int64  `json:"value"`
}

type ScopePin struct {
	Paths []string `json:"paths"`
}

type TransformationProposal struct {
	Present            bool   `json:"present"`
	Patch              string `json:"patch"`
	OutputMode         string `json:"output_mode"`
	RepositoryMutation bool   `json:"repository_mutation"`
}

type EffectPrediction struct {
	Present bool   `json:"present"`
	Metric  string `json:"metric"`
	Before  int64  `json:"before"`
	After   int64  `json:"after"`
}
