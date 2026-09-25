package replay

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

var ExpectedStages = []string{
	"READ_SOURCE",
	"PARSE_AST",
	"LOWER_IR",
	"NORMALIZE_IR",
	"REPLAY_IR",
	"SEAL_EFFECTS",
}

type EffectReceipt struct {
	Reads     []string `json:"reads"`
	Writes    int      `json:"writes"`
	Network   int      `json:"network"`
	Processes int      `json:"processes"`
}

type Observation struct {
	Path               string        `json:"path"`
	SourceLines        int           `json:"source_lines"`
	SourceDigest       string        `json:"source_digest"`
	IRVersion          string        `json:"ir_version"`
	Package            string        `json:"package"`
	Namespace          string        `json:"namespace"`
	Nodes              int           `json:"nodes"`
	DeterministicFacts int           `json:"deterministic_facts"`
	CandidateFacts     int           `json:"candidate_facts"`
	Normalized         bool          `json:"normalized"`
	CanonicalReplay    bool          `json:"canonical_replay"`
	SemanticReplay     bool          `json:"semantic_replay"`
	ProvenanceReplay   bool          `json:"provenance_replay"`
	EvidenceReplay     bool          `json:"evidence_replay"`
	SemanticHash       string        `json:"semantic_hash"`
	ProvenanceHash     string        `json:"provenance_hash"`
	EvidenceHash       string        `json:"evidence_hash"`
	Stages             []string      `json:"stages"`
	Effects            EffectReceipt `json:"effects"`
	IR                 semantic.IR   `json:"-"`
}
