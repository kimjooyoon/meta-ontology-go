package improvement

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const (
	SnapshotSchema   = "gooo/language-readiness-snapshot/v1"
	TransitionSchema = "gooo/quantified-improvement-transition/v1"
	SnapshotTotal    = int64(24)

	Satisfied    EvidenceStatus = "SATISFIED"
	NotSatisfied EvidenceStatus = "NOT_SATISFIED"
	Unresolved   EvidenceStatus = "UNRESOLVED"

	Improved        Decision = "IMPROVED"
	NoChange        Decision = "NO_CHANGE"
	Regressed       Decision = "REGRESSED"
	LowerResolution Decision = "LOWER_RESOLUTION"
)

type EvidenceStatus string

type Decision string

type Evidence struct {
	ID     string         `json:"id"`
	Status EvidenceStatus `json:"status"`
}

// Snapshot excludes prose so only registered, replayable evidence can count.
type Snapshot struct {
	ContractSchema string     `json:"contract_schema"`
	RegistryDigest string     `json:"registry_digest"`
	Completed      int64      `json:"completed"`
	Total          int64      `json:"total"`
	BasisPoints    int64      `json:"basis_points"`
	Evidence       []Evidence `json:"evidence"`
}

func validDigest(value string) bool {
	const prefix = "sha256:"
	if !strings.HasPrefix(value, prefix) || len(value) != len(prefix)+sha256.Size*2 {
		return false
	}
	payload := strings.TrimPrefix(value, prefix)
	if payload != strings.ToLower(payload) {
		return false
	}
	_, err := hex.DecodeString(payload)
	return err == nil
}
