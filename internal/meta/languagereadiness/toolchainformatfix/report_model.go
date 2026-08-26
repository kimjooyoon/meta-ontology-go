package toolchainformatfix

import cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"

type Source struct {
	ExpectedHeadSHA  string `json:"expected_head_sha"`
	GoVersion        string `json:"go_version"`
	ConceptDigest    string `json:"concept_digest"`
	RegistryDigest   string `json:"registry_digest"`
	ExecutableDigest string `json:"executable_digest"`
	ObservationKnown bool   `json:"observation_known"`
}

type CaseResult struct {
	Definition       Definition             `json:"definition"`
	Arguments        []string               `json:"arguments"`
	First            cliruntime.Observation `json:"first"`
	Replay           cliruntime.Observation `json:"replay"`
	Status           string                 `json:"status"`
	Reason           string                 `json:"reason"`
	Invocations      int                    `json:"invocations"`
	ExitMatched      bool                   `json:"exit_matched"`
	OutputMatched    bool                   `json:"output_matched"`
	ReplayMatched    bool                   `json:"replay_matched"`
	StructuredOutput int                    `json:"structured_output"`
	StructuredPlan   int                    `json:"structured_plan"`
	RepositoryWrites int                    `json:"repository_writes"`
	EvidenceDigest   string                 `json:"evidence_digest"`
}

type Summary struct {
	Satisfied, Total, ReadinessBPS, PositivePaths, GuardrailRejections int
	Executed, Invocations, StructuredOutputs, StructuredPlans          int
	InMemoryApplications, FixedPoints, ReplayMatches, BinaryBindings   int
	Unresolved, ExitMismatches, OutputMismatches, ReplayMismatches     int
	RepositoryWrites, DirectWrites, RegistryDrift                      int
}
