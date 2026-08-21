package cache

import (
	"errors"
)

const (
	cacheReceiptSchemaVersion     = "v2"
	benchmarkReceiptSchemaVersion = "v2"
)

var (
	ErrInvalidReceipt   = errors.New("invalid cache evidence receipt")
	ErrReceiptReplay    = errors.New("cache evidence receipt replay")
	ErrUnsafeReceiptLog = errors.New("unsafe cache receipt log")
)

// EvidenceRef identifies immutable external evidence included in a receipt.
type EvidenceRef struct {
	Name   string `json:"name"`
	Digest Digest `json:"digest"`
}

// EvidenceFreshness is separate from ProjectionKey and records the exact
// source, verifier, policy, and CI evidence used for a cache decision.
type EvidenceFreshness struct {
	BaseDigest         Digest                  `json:"base_digest"`
	HeadDigest         Digest                  `json:"head_digest"`
	BaseSHA            string                  `json:"base_sha"`
	HeadSHA            string                  `json:"head_sha"`
	CheckoutRef        string                  `json:"checkout_ref"`
	RunID              string                  `json:"run_id"`
	Event              string                  `json:"event"`
	EventRef           string                  `json:"event_ref"`
	EventID            string                  `json:"event_id"`
	Attempt            uint64                  `json:"attempt"`
	Jobs               map[string]FreshnessJob `json:"jobs"`
	PredecessorDigests []Digest                `json:"predecessor_digests"`
	SourceDigest       Digest                  `json:"source_digest"`
	IRDigest           Digest                  `json:"ir_digest"`
	PolicyDigest       Digest                  `json:"policy_digest"`
	ToolchainDigest    Digest                  `json:"toolchain_digest"`
	TargetDigest       Digest                  `json:"target_digest"`
	BundleDigest       Digest                  `json:"bundle_digest"`
	EvidenceRefs       []EvidenceRef           `json:"evidence_refs"`
}
