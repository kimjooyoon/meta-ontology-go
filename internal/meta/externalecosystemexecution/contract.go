package externalecosystemexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const (
	SchemaVersion             = "external-ecosystem-execution-report/v1"
	ObservationSchema         = "external-ecosystem-execution-observation/v1"
	ContractVersion           = "external-ecosystem-execution/v1"
	DenominatorVersion        = "external-ecosystem-execution-denominator/v1"
	ReferenceContractVersion  = "external-ecosystem-conformance/v1"
	ExpectedReferenceDecision = "REFERENCE_ONLY"
	ExpectedReferenceURL      = "https://github.com/cosmos72/gomacro"
	ExpectedCommit            = "cf0d4bf32da393dbda97e3572f216731013ffa55"
	ExpectedTree              = "8cc240a53dd29432ad83620b20fd8a0a05674c6d"
	ExpectedModuleGo          = "1.23.0"
	ExpectedGoVersion         = "go1.27.0"
	DecisionConfirmed         = "EXECUTION_CONFIRMED"
	DecisionFailClosed        = "FAIL_CLOSED"
)

type Criterion struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

var denominator = []Criterion{
	{"reference-binding-exact", "driver"},
	{"pinned-commit-exact", "driver"},
	{"pinned-tree-exact", "driver"},
	{"go-1.27.0-exact", "driver"},
	{"external-run-1-passed", "outcome"},
	{"external-run-2-passed", "outcome"},
	{"normalized-replay-equal", "outcome"},
	{"repository-write-boundaries-exact", "guardrail"},
}

func Criteria() []Criterion { return append([]Criterion(nil), denominator...) }

func Digest(value any) string {
	b, _ := json.Marshal(value)
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DenominatorDigest() string { return Digest(denominator) }
