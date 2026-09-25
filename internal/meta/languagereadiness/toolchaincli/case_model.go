package toolchaincli

import cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"

type CaseResult struct {
	Definition         Definition             `json:"definition"`
	Arguments          []string               `json:"arguments"`
	First              cliruntime.Observation `json:"first"`
	Replay             cliruntime.Observation `json:"replay"`
	Status             string                 `json:"status"`
	Reason             string                 `json:"reason"`
	Invocations        int                    `json:"invocations"`
	ExitMatched        bool                   `json:"exit_matched"`
	StdoutMatched      bool                   `json:"stdout_matched"`
	StderrMatched      bool                   `json:"stderr_matched"`
	ReplayMatched      bool                   `json:"replay_matched"`
	StructuredOutputs  int                    `json:"structured_outputs"`
	LanguageOperations int                    `json:"language_operations"`
	DeclaredCommands   int                    `json:"declared_commands"`
	RepositoryWrites   int                    `json:"repository_writes"`
	EvidenceDigest     string                 `json:"evidence_digest"`
}
