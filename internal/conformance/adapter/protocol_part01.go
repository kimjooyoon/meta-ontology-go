package adapter

import (
	"encoding/json"
)

const (
	ProtocolSchema = "gooo/codegen-adapter/v1"
	EvidenceSchema = "gooo/evidence/v1"
)

// Status is the semantic result of one fixture operation.
type Status string

const (
	StatusPass     Status = "pass"
	StatusFail     Status = "fail"
	StatusDeferred Status = "deferred"
	StatusNotRun   Status = "not-run"
)

// Operation identifies a supported adapter boundary.
type Operation string

const (
	OperationParseAST        Operation = "parse-ast"
	OperationLowerIR         Operation = "lower-ir"
	OperationGenerate        Operation = "generate"
	OperationLiftBX          Operation = "lift-bx"
	OperationResolveLSP      Operation = "resolve-lsp"
	OperationCacheKey        Operation = "cache-key"
	OperationEmitEvidence    Operation = "emit-evidence"
	OperationCompareEvidence Operation = "compare-evidence"
)

// Request is the canonical runner-to-adapter input.
type Request struct {
	Schema    string      `json:"schema"`
	Fixture   string      `json:"fixture"`
	Operation Operation   `json:"operation"`
	RunID     string      `json:"run_id"`
	Input     Input       `json:"input"`
	Contract  Contract    `json:"contract"`
	Options   Options     `json:"options"`
	Expected  Expectation `json:"expected"`
}

// Input contains one authoritative source view and optional previous Go.
type Input struct {
	DSL        string          `json:"dsl,omitempty"`
	IR         json.RawMessage `json:"ir,omitempty"`
	PreviousGo string          `json:"previous_go,omitempty"`
	SourceURI  string          `json:"source_uri,omitempty"`
}

// Contract names the versioned boundaries an adapter must honor.
type Contract struct {
	AST        string `json:"ast"`
	IR         string `json:"ir"`
	Generator  string `json:"generator"`
	Marker     string `json:"marker"`
	PolicyHash string `json:"policy_sha256"`
}

// Options controls only runner behavior; adapters must not mutate them.
type Options struct {
	CanonicalOutput bool `json:"canonical_output"`
	AllowMigration  bool `json:"allow_migration"`
}
