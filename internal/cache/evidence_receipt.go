package cache

import (
	"errors"
	"fmt"
)

const (
	cacheReceiptSchemaVersion     = "v2"
	benchmarkReceiptSchemaVersion = "v2"
)

var (
	ErrInvalidReceipt = errors.New("invalid cache evidence receipt")
	ErrReceiptReplay  = errors.New("cache evidence receipt replay")
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
	RunID              string                  `json:"run_id"`
	Event              string                  `json:"event"`
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

// Validate rejects missing, zero, duplicated, or malformed evidence.
func (e EvidenceFreshness) Validate() error {
	for _, item := range []struct {
		label string
		value Digest
	}{
		{"base", e.BaseDigest}, {"head", e.HeadDigest}, {"source", e.SourceDigest},
		{"IR", e.IRDigest}, {"policy", e.PolicyDigest}, {"toolchain", e.ToolchainDigest},
		{"target", e.TargetDigest}, {"bundle", e.BundleDigest},
	} {
		if !item.value.Known() {
			return fmt.Errorf("%w: unknown %s digest", ErrInvalidReceipt, item.label)
		}
	}
	if !validCommitSHA(e.BaseSHA) || !validCommitSHA(e.HeadSHA) {
		return fmt.Errorf("%w: missing immutable base/head SHA", ErrInvalidReceipt)
	}
	if e.RunID == "" || e.Event != "pull_request" || e.EventID == "" || e.Attempt == 0 {
		return fmt.Errorf("%w: missing immutable event attempt", ErrInvalidReceipt)
	}
	if err := validateFreshnessJobs(e.Jobs, e.HeadSHA); err != nil {
		return err
	}
	if e.PredecessorDigests == nil || hasDuplicateDigests(e.PredecessorDigests) {
		return fmt.Errorf("%w: missing or replayed predecessors", ErrInvalidReceipt)
	}
	if len(e.EvidenceRefs) == 0 || hasDuplicateEvidenceRefs(e.EvidenceRefs) {
		return fmt.Errorf("%w: missing or duplicated evidence refs", ErrInvalidReceipt)
	}
	for _, digest := range e.PredecessorDigests {
		if !digest.Known() {
			return fmt.Errorf("%w: unknown predecessor digest", ErrInvalidReceipt)
		}
	}
	for _, ref := range e.EvidenceRefs {
		if err := validateKeyComponent("evidence ref", ref.Name, true); err != nil ||
			!knownEvidenceRef(ref.Name) || !ref.Digest.Known() {
			return fmt.Errorf("%w: malformed evidence ref %q", ErrInvalidReceipt, ref.Name)
		}
	}
	if err := validateEvidenceBindings(e); err != nil {
		return err
	}
	return nil
}

// Equal reports whether two evidence bundles identify the same immutable run.
func (e EvidenceFreshness) Equal(other EvidenceFreshness) bool {
	left, right := canonicalEvidence(e), canonicalEvidence(other)
	return left.BaseDigest == right.BaseDigest && left.HeadDigest == right.HeadDigest &&
		left.BaseSHA == right.BaseSHA && left.HeadSHA == right.HeadSHA && left.Event == right.Event &&
		left.RunID == right.RunID && left.EventID == right.EventID && left.Attempt == right.Attempt &&
		freshnessJobsEqual(left.Jobs, right.Jobs) &&
		left.SourceDigest == right.SourceDigest &&
		left.IRDigest == right.IRDigest && left.PolicyDigest == right.PolicyDigest &&
		left.ToolchainDigest == right.ToolchainDigest && left.TargetDigest == right.TargetDigest &&
		left.BundleDigest == right.BundleDigest && digestSliceEqual(left.PredecessorDigests, right.PredecessorDigests) &&
		evidenceRefsEqual(left.EvidenceRefs, right.EvidenceRefs)
}

// Matches reports whether evidence belongs to the requested current base,
// head, run, and predecessor tuple.
func (e EvidenceFreshness) Matches(current EvidenceFreshness) bool {
	return e.Equal(current)
}

// CacheReceipt is an append-only immutable record of one cache decision.
type CacheReceipt struct {
	SchemaVersion         string            `json:"schema_version"`
	CacheKey              Digest            `json:"cache_key"`
	Domain                string            `json:"domain"`
	KeyVersion            string            `json:"key_version"`
	HostStage             HostStage         `json:"host_stage"`
	ArtifactKind          string            `json:"artifact_kind"`
	Projection            string            `json:"projection"`
	SemanticClosureDigest Digest            `json:"semantic_closure_digest"`
	DependencyRoot        Digest            `json:"dependency_root"`
	DirectDependencies    []Digest          `json:"direct_dependencies"`
	PolicySchemaDigest    Digest            `json:"policy_schema_digest"`
	Toolchain             string            `json:"toolchain"`
	Target                string            `json:"target"`
	BuildTagsDigest       Digest            `json:"build_tags_digest"`
	OptionsDigest         Digest            `json:"options_digest"`
	ContentDigest         Digest            `json:"content_digest"`
	Size                  int64             `json:"size"`
	Reconstructable       bool              `json:"reconstructable"`
	EvidenceRefs          []EvidenceRef     `json:"evidence_refs"`
	ProducerHost          string            `json:"producer_host"`
	Status                ReceiptStatus     `json:"status"`
	Evidence              EvidenceFreshness `json:"evidence"`
	ReceiptDigest         Digest            `json:"receipt_digest"`
}

// ReceiptStatus describes the observed cache outcome, never feature proof.
type ReceiptStatus string

const (
	ReceiptHit        ReceiptStatus = "hit"
	ReceiptMiss       ReceiptStatus = "miss"
	ReceiptRecomputed ReceiptStatus = "recomputed"
	ReceiptStale      ReceiptStatus = "stale"
	ReceiptCorrupt    ReceiptStatus = "corrupt"
)

// Validate checks the required cache receipt schema before it is sealed.
func (r CacheReceipt) Validate() error {
	if r.SchemaVersion != cacheReceiptSchemaVersion || !r.CacheKey.Known() ||
		!r.SemanticClosureDigest.Known() || !r.DependencyRoot.Known() || r.DirectDependencies == nil ||
		!r.PolicySchemaDigest.Known() || r.Toolchain == "" || r.Target == "" || !r.BuildTagsDigest.Known() ||
		!r.OptionsDigest.Known() ||
		r.ProducerHost == "" || !validReceiptStatus(r.Status) || r.Size < 0 {
		return fmt.Errorf("%w: required cache receipt field missing", ErrInvalidReceipt)
	}
	for _, field := range []struct{ label, value string }{
		{"domain", r.Domain}, {"key version", r.KeyVersion}, {"artifact kind", r.ArtifactKind},
		{"projection", r.Projection}, {"toolchain", r.Toolchain}, {"target", r.Target},
		{"producer host", r.ProducerHost},
	} {
		if err := validateKeyComponent(field.label, field.value, true); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidReceipt, err)
		}
	}
	if !r.HostStage.Valid() {
		return fmt.Errorf("%w: invalid host stage", ErrInvalidReceipt)
	}
	if r.hasArtifact() && (!r.ContentDigest.Known() || !r.Reconstructable) {
		return fmt.Errorf("%w: artifact receipt is not reconstructable", ErrInvalidReceipt)
	}
	if !r.hasArtifact() && (r.ContentDigest != "" || r.Size != 0 || r.Reconstructable) {
		return fmt.Errorf("%w: non-artifact receipt has content", ErrInvalidReceipt)
	}
	if err := validateEvidenceRefs(r.EvidenceRefs); err != nil {
		return err
	}
	if err := r.Evidence.Validate(); err != nil {
		return err
	}
	if !evidenceRefsEqual(r.EvidenceRefs, r.Evidence.EvidenceRefs) {
		return fmt.Errorf("%w: evidence refs disagree", ErrInvalidReceipt)
	}
	for _, digest := range r.DirectDependencies {
		if !digest.Known() {
			return fmt.Errorf("%w: unknown direct dependency", ErrInvalidReceipt)
		}
	}
	return nil
}

// ValidateForKey prevents metadata-only artifact or projection aliases from
// being accepted as a different content-addressed projection.
func (r CacheReceipt) ValidateForKey(key Key) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if err := validateFullKey(key); err != nil {
		return err
	}
	if r.CacheKey != key.Digest || r.Domain != key.Domain || r.KeyVersion != key.Version ||
		r.HostStage != key.HostStage || r.ArtifactKind != key.ArtifactKind || r.Projection != key.Projection ||
		r.SemanticClosureDigest != key.SemanticClosureDigest || r.DependencyRoot != key.DependencyRoot ||
		r.PolicySchemaDigest != key.PolicySchemaDigest || r.Toolchain != key.Toolchain ||
		r.OptionsDigest != key.OptionsDigest ||
		r.Target != key.Target || r.BuildTagsDigest != key.BuildTagsDigest {
		return fmt.Errorf("%w: receipt identity differs from key", ErrInvalidReceipt)
	}
	return nil
}

// ValidateForData binds an artifact receipt to the exact bytes it describes.
func (r CacheReceipt) ValidateForData(data []byte) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if !r.hasArtifact() || r.ContentDigest != HashBytes(data) || r.Size != int64(len(data)) {
		return fmt.Errorf("%w: receipt content differs from artifact", ErrInvalidReceipt)
	}
	return nil
}

// Seal validates and computes the immutable receipt digest.
func (r CacheReceipt) Seal() (CacheReceipt, error) {
	if err := r.Validate(); err != nil {
		return CacheReceipt{}, err
	}
	r.Evidence = canonicalEvidence(r.Evidence)
	r.EvidenceRefs = append([]EvidenceRef(nil), r.Evidence.EvidenceRefs...)
	provided := r.ReceiptDigest
	r.ReceiptDigest = ""
	digest, err := DigestOf(r)
	if err != nil {
		return CacheReceipt{}, fmt.Errorf("%w: hash receipt: %v", ErrInvalidReceipt, err)
	}
	if provided != "" && provided != digest {
		return CacheReceipt{}, fmt.Errorf("%w: receipt digest mismatch", ErrInvalidReceipt)
	}
	r.ReceiptDigest = digest
	return r, nil
}

func validReceiptStatus(status ReceiptStatus) bool {
	return status == ReceiptHit || status == ReceiptMiss || status == ReceiptRecomputed ||
		status == ReceiptStale || status == ReceiptCorrupt
}

func (r CacheReceipt) hasArtifact() bool {
	return r.Status == ReceiptHit || r.Status == ReceiptRecomputed
}

func validateEvidenceRefs(refs []EvidenceRef) error {
	if len(refs) == 0 || hasDuplicateEvidenceRefs(refs) {
		return fmt.Errorf("%w: missing or duplicated evidence refs", ErrInvalidReceipt)
	}
	for _, ref := range refs {
		if err := validateKeyComponent("evidence ref", ref.Name, true); err != nil ||
			!knownEvidenceRef(ref.Name) || !ref.Digest.Known() {
			return fmt.Errorf("%w: malformed evidence ref %q", ErrInvalidReceipt, ref.Name)
		}
	}
	return nil
}

func knownEvidenceRef(name string) bool {
	switch name {
	case "source", "ir", "policy", "toolchain", "target", "bundle":
		return true
	default:
		return false
	}
}

func validateEvidenceBindings(e EvidenceFreshness) error {
	bound := make(map[string]Digest, len(e.EvidenceRefs))
	for _, ref := range e.EvidenceRefs {
		bound[ref.Name] = ref.Digest
	}
	for _, required := range []struct {
		name   string
		digest Digest
	}{
		{"policy", e.PolicyDigest}, {"toolchain", e.ToolchainDigest},
	} {
		if bound[required.name] != required.digest {
			return fmt.Errorf("%w: evidence ref %q is not bound", ErrInvalidReceipt, required.name)
		}
	}
	return nil
}
