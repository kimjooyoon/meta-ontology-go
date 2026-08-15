// Package adapter defines the dependency-free fixture protocol used to compare
// independent AST, IR, BX, codegen, LSP, cache, and evidence adapters.
package adapter

import (
	"encoding/json"
	"fmt"
	"strings"
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

// Expectation is owned by the runner oracle, not by an adapter.
type Expectation struct {
	Status      Status `json:"status"`
	FailureCode string `json:"failure_code,omitempty"`
}

// Response is the canonical adapter-to-runner result.
type Response struct {
	Schema            string           `json:"schema"`
	Fixture           string           `json:"fixture"`
	Operation         Operation        `json:"operation"`
	RunID             string           `json:"run_id"`
	Status            Status           `json:"status"`
	Failure           *Failure         `json:"failure,omitempty"`
	PromotionEligible bool             `json:"promotion_eligible"`
	Observed          Observed         `json:"observed"`
	Measurements      Measurements     `json:"measurements"`
	Evidence          EvidenceArtifact `json:"evidence"`
	ProducerClaims    ProducerClaims   `json:"producer_claims,omitempty"`
}

// ProducerClaims are advisory and are never accepted as observer proof.
type ProducerClaims struct {
	NoWrite *bool `json:"no_write,omitempty"`
}

// Failure identifies a deterministic safety or conformance rejection.
type Failure struct {
	Code       string `json:"code"`
	SemanticID string `json:"semantic_id,omitempty"`
	Detail     string `json:"detail,omitempty"`
}

func (r Request) Validate() error {
	if r.Schema != ProtocolSchema {
		return fmt.Errorf("unsupported request schema %q", r.Schema)
	}
	if strings.TrimSpace(r.RunID) == "" {
		return fmt.Errorf("run_id is required")
	}
	if strings.TrimSpace(r.Fixture) == "" {
		return fmt.Errorf("fixture is required")
	}
	if !knownOperation(r.Operation) {
		return fmt.Errorf("unsupported operation %q", r.Operation)
	}
	if err := r.Contract.validate(); err != nil {
		return err
	}
	if err := r.Expected.validate(); err != nil {
		return err
	}
	if err := validateSourceURI(r.Input.SourceURI); err != nil {
		return err
	}
	return validateAuthoritativeInput(r.Operation, r.Input)
}

func (c Contract) validate() error {
	fields := []struct {
		name  string
		value string
	}{
		{"ast contract", c.AST}, {"ir contract", c.IR},
		{"generator contract", c.Generator}, {"marker contract", c.Marker},
		{"policy digest", c.PolicyHash},
	}
	for _, field := range fields {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s is required", field.name)
		}
	}
	return nil
}

func (e Expectation) validate() error {
	if !validStatus(e.Status) {
		return fmt.Errorf("unsupported expected status %q", e.Status)
	}
	if e.Status != StatusFail && e.FailureCode != "" {
		return fmt.Errorf("failure code requires expected fail status")
	}
	return nil
}

func validateAuthoritativeInput(operation Operation, input Input) error {
	if operation == OperationParseAST && strings.TrimSpace(input.DSL) == "" {
		return fmt.Errorf("parse-ast requires DSL input")
	}
	if operation == OperationLowerIR && strings.TrimSpace(input.DSL) == "" && !hasJSON(input.IR) {
		return fmt.Errorf("lower-ir requires DSL or IR input")
	}
	if operation == OperationGenerate && !hasJSON(input.IR) {
		return fmt.Errorf("generate requires IR input")
	}
	return nil
}

func hasJSON(value json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(value))
	return trimmed != "" && trimmed != "null"
}

func validStatus(status Status) bool {
	return status == StatusPass || status == StatusFail || status == StatusDeferred || status == StatusNotRun
}

func knownOperation(operation Operation) bool {
	switch operation {
	case OperationParseAST, OperationLowerIR, OperationGenerate, OperationLiftBX,
		OperationResolveLSP, OperationCacheKey, OperationEmitEvidence, OperationCompareEvidence:
		return true
	default:
		return false
	}
}
