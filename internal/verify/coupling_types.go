package verify

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const (
	CouplingRegistrySchemaVersion = "gooo/code-semantic-coupling-registry/v1"
	CouplingEnvelopeSchemaVersion = "gooo/code-semantic-coupling-envelope/v1"
	CouplingEvidenceSchemaVersion = "gooo/code-semantic-coupling-evidence/v1"
	CouplingContractDigest        = "2b7f664f7b8c5ccbe0f40f63ccee50514df68a78f3209533d7edc29949a8ad6d"
	CouplingSemanticSourceHead    = "bafef7cd6802647ab943654708fab0629c77be1b"
	CouplingFailureCodePrefix     = "CI-COUPLING-001#"
)

const (
	CouplingApplicable    = "APPLICABLE"
	CouplingNotApplicable = "NOT_APPLICABLE"
	CouplingCurrent       = "CURRENT"

	CouplingDecisionPass       = "PASS"
	CouplingDecisionFailClosed = "FAIL_CLOSED"
	CouplingDecisionUnknown    = "UNKNOWN"

	CouplingEnforcementAllow = "ALLOW"
	CouplingEnforcementBlock = "BLOCK"

	CouplingDomainFeature     = "FEATURE_LANE"
	CouplingDomainDependency  = "DEPENDENCY_LOCAL"
	CouplingDomainIntegrity   = "REPOSITORY_INTEGRITY"
	CouplingOwnerUnavailable  = "observer"
	CouplingObserverAvailable = "AVAILABLE"
)

// CouplingRule is an immutable rule binding in the docs-owned registry.
type CouplingRule struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

// CouplingSurface binds a stable code surface to stable semantic identity.
// Paths, labels, and names are locators only; IDs establish identity.
type CouplingSurface struct {
	SurfaceID              string   `json:"surface_id"`
	CodeSymbolID           string   `json:"code_symbol_id"`
	SemanticOwnerID        string   `json:"semantic_owner_id"`
	ScopeID                string   `json:"scope_id"`
	CodePathPatterns       []string `json:"code_path_patterns"`
	SourceMapID            string   `json:"source_map_id"`
	SourceMapBindingDigest string   `json:"source_map_binding_digest"`
	SemanticSourceIDs      []string `json:"semantic_source_ids"`
	SemanticSourcePaths    []string `json:"semantic_source_paths"`
	ProfileID              string   `json:"profile_id"`
	ProfileVersion         string   `json:"profile_version"`
	ProfileDigest          string   `json:"profile_digest"`
	ToolchainDigest        string   `json:"toolchain_digest"`
	RuleDigests            []string `json:"rule_digests"`
	Applicability          string   `json:"applicability"`
}

// CouplingRegistry is versioned and digestable. Callers must bind its Digest
// to the exact snapshot envelope; it is never inferred from changed paths.
type CouplingRegistry struct {
	Schema   string            `json:"schema"`
	Version  string            `json:"version"`
	Surfaces []CouplingSurface `json:"surfaces"`
}

// CouplingEnvelope is the exact CI and semantic input binding.
type CouplingEnvelope struct {
	Schema             string `json:"schema"`
	ContractDigest     string `json:"contract_digest"`
	Repository         string `json:"repository"`
	Event              string `json:"event"`
	Ref                string `json:"ref"`
	EventRef           string `json:"event_ref"`
	CheckoutRef        string `json:"checkout_ref"`
	BaseRef            string `json:"base_ref"`
	BaseSHA            string `json:"base_sha"`
	HeadSHA            string `json:"head_sha"`
	WorkflowSHA        string `json:"workflow_sha"`
	PRNumber           int64  `json:"pr_number"`
	RunID              int64  `json:"run_id"`
	RunAttempt         int64  `json:"run_attempt"`
	CatalogDigest      string `json:"catalog_digest"`
	PolicyDigest       string `json:"policy_digest"`
	RegistryDigest     string `json:"registry_digest"`
	ProfileDigest      string `json:"profile_digest"`
	ToolchainDigest    string `json:"toolchain_digest"`
	SchemaDigest       string `json:"schema_digest"`
	SnapshotDigest     string `json:"snapshot_digest"`
	SemanticSourceHead string `json:"semantic_source_head"`
	SemanticPathSchema string `json:"semantic_path_schema"`
}

// ChangedCodeSite is source-backed changed-surface input. CodeSymbolID is
// optional only when the registry resolves the path to exactly one surface.
type ChangedCodeSite struct {
	Path                   string `json:"path"`
	CodeSymbolID           string `json:"code_symbol_id"`
	SourceMapBindingDigest string `json:"source_map_binding_digest"`
}

// CouplingReceipt is the input proof for one changed surface. Path is kept as
// a typed semantic value for validation and deliberately omitted from emitted
// JSON; PathDigest and the producer-independent evidence references are the
// serialized proof binding.
type CouplingReceipt struct {
	ReceiptID                   string                   `json:"receipt_id"`
	SurfaceID                   string                   `json:"surface_id"`
	SemanticOwnerID             string                   `json:"semantic_owner_id"`
	CodeSymbolID                string                   `json:"code_symbol_id"`
	EnvelopeDigest              string                   `json:"envelope_digest"`
	SnapshotDigest              string                   `json:"snapshot_digest"`
	RegistryDigest              string                   `json:"registry_digest"`
	ProfileDigest               string                   `json:"profile_digest"`
	ToolchainDigest             string                   `json:"toolchain_digest"`
	RuleDigest                  string                   `json:"rule_digest"`
	SourceMapBindingDigest      string                   `json:"source_map_binding_digest"`
	BeforeIRDigest              string                   `json:"before_ir_digest"`
	AfterIRDigest               string                   `json:"after_ir_digest"`
	AuthoritySourceBeforeDigest string                   `json:"authority_source_before_digest"`
	AuthoritySourceAfterDigest  string                   `json:"authority_source_after_digest"`
	ChangeClaim                 string                   `json:"change_claim"`
	ReceiptKind                 string                   `json:"receipt_kind"`
	CanonicalDelta              string                   `json:"canonical_delta"`
	DeltaDigest                 string                   `json:"delta_digest"`
	PathDigest                  string                   `json:"path_digest"`
	OriginPathIDs               []string                 `json:"origin_path_ids"`
	EvidenceRefs                []string                 `json:"evidence_refs"`
	State                       string                   `json:"state"`
	CanonicalPayload            string                   `json:"canonical_payload"`
	Path                        semantic.InferencePathV1 `json:"-"`
}

// CouplingInput is intentionally data-only. Verification performs no writes.
type CouplingInput struct {
	Schema       string            `json:"schema"`
	Envelope     CouplingEnvelope  `json:"envelope"`
	Registry     CouplingRegistry  `json:"registry"`
	ChangedSites []ChangedCodeSite `json:"changed_sites"`
	Receipts     []CouplingReceipt `json:"receipts"`
}

type CouplingSurfaceResult struct {
	SurfaceID   string   `json:"surface_id"`
	Decision    string   `json:"decision"`
	ReasonCodes []string `json:"reason_codes"`
}

type CouplingFailure struct {
	Code   string `json:"code"`
	Domain string `json:"domain"`
	Owner  string `json:"owner"`
	Retry  bool   `json:"retry"`
	Detail string `json:"detail"`
}

type couplingReceiptPayload struct {
	ReceiptID                   string   `json:"receipt_id"`
	SurfaceID                   string   `json:"surface_id"`
	SemanticOwnerID             string   `json:"semantic_owner_id"`
	CodeSymbolID                string   `json:"code_symbol_id"`
	EnvelopeDigest              string   `json:"envelope_digest"`
	SnapshotDigest              string   `json:"snapshot_digest"`
	RegistryDigest              string   `json:"registry_digest"`
	ProfileDigest               string   `json:"profile_digest"`
	ToolchainDigest             string   `json:"toolchain_digest"`
	RuleDigest                  string   `json:"rule_digest"`
	SourceMapBindingDigest      string   `json:"source_map_binding_digest"`
	BeforeIRDigest              string   `json:"before_ir_digest"`
	AfterIRDigest               string   `json:"after_ir_digest"`
	AuthoritySourceBeforeDigest string   `json:"authority_source_before_digest"`
	AuthoritySourceAfterDigest  string   `json:"authority_source_after_digest"`
	ChangeClaim                 string   `json:"change_claim"`
	ReceiptKind                 string   `json:"receipt_kind"`
	CanonicalDelta              string   `json:"canonical_delta"`
	DeltaDigest                 string   `json:"delta_digest"`
	PathDigest                  string   `json:"path_digest"`
	OriginPathIDs               []string `json:"origin_path_ids"`
	EvidenceRefs                []string `json:"evidence_refs"`
	State                       string   `json:"state"`
}

type CouplingReceiptEvidence struct {
	couplingReceiptPayload
	CanonicalPayload string `json:"canonical_payload"`
}

// CouplingEvidence is the raw, full-vector, producer-independent result.
type CouplingEvidence struct {
	Schema            string                    `json:"schema"`
	Observer          string                    `json:"observer"`
	ObserverAvailable string                    `json:"observer_available"`
	EnvelopeDigest    string                    `json:"envelope_digest"`
	RegistryDigest    string                    `json:"registry_digest"`
	RawDecision       string                    `json:"raw_decision"`
	Enforcement       string                    `json:"enforcement"`
	SurfaceResults    []CouplingSurfaceResult   `json:"surface_results"`
	Receipts          []CouplingReceiptEvidence `json:"receipts"`
	Failures          []CouplingFailure         `json:"failures"`
}

func (e CouplingEnvelope) Canonical() string {
	parts := []string{
		CouplingEnvelopeSchemaVersion, e.ContractDigest, e.Repository, e.Event,
		e.Ref, e.EventRef, e.CheckoutRef, e.BaseRef, e.BaseSHA, e.HeadSHA,
		e.WorkflowSHA, fmt.Sprint(e.PRNumber), fmt.Sprint(e.RunID), fmt.Sprint(e.RunAttempt),
		e.CatalogDigest, e.PolicyDigest, e.RegistryDigest, e.ProfileDigest,
		e.ToolchainDigest, e.SchemaDigest, e.SnapshotDigest, e.SemanticSourceHead,
		e.SemanticPathSchema,
	}
	return strings.Join(parts, "\x00")
}

func (e CouplingEnvelope) TupleDigest() string { return semantic.StableHashString(e.Canonical()) }

func (r CouplingReceipt) payload() couplingReceiptPayload {
	return couplingReceiptPayload{
		ReceiptID: r.ReceiptID, SurfaceID: r.SurfaceID, SemanticOwnerID: r.SemanticOwnerID,
		CodeSymbolID: r.CodeSymbolID, EnvelopeDigest: r.EnvelopeDigest, SnapshotDigest: r.SnapshotDigest,
		RegistryDigest: r.RegistryDigest, ProfileDigest: r.ProfileDigest, ToolchainDigest: r.ToolchainDigest,
		RuleDigest: r.RuleDigest, SourceMapBindingDigest: r.SourceMapBindingDigest,
		BeforeIRDigest: r.BeforeIRDigest, AfterIRDigest: r.AfterIRDigest,
		AuthoritySourceBeforeDigest: r.AuthoritySourceBeforeDigest,
		AuthoritySourceAfterDigest:  r.AuthoritySourceAfterDigest, ChangeClaim: r.ChangeClaim,
		ReceiptKind: r.ReceiptKind, CanonicalDelta: r.CanonicalDelta, DeltaDigest: r.DeltaDigest,
		PathDigest: r.PathDigest, OriginPathIDs: sortedStrings(r.OriginPathIDs),
		EvidenceRefs: sortedStrings(r.EvidenceRefs), State: r.State,
	}
}

func (r CouplingReceipt) ExpectedCanonicalPayload() (string, error) {
	data, err := json.Marshal(r.payload())
	return string(data), err
}

func (r CouplingReceipt) Evidence() CouplingReceiptEvidence {
	payload := r.payload()
	data, _ := json.Marshal(payload)
	return CouplingReceiptEvidence{couplingReceiptPayload: payload, CanonicalPayload: string(data)}
}

func (e CouplingEvidence) CanonicalJSON() ([]byte, error) {
	normalized := e
	normalized.SurfaceResults = append([]CouplingSurfaceResult(nil), e.SurfaceResults...)
	normalized.Receipts = append([]CouplingReceiptEvidence(nil), e.Receipts...)
	normalized.Failures = append([]CouplingFailure(nil), e.Failures...)
	sort.Slice(normalized.SurfaceResults, func(i, j int) bool {
		return normalized.SurfaceResults[i].SurfaceID < normalized.SurfaceResults[j].SurfaceID
	})
	sort.Slice(normalized.Receipts, func(i, j int) bool { return normalized.Receipts[i].SurfaceID < normalized.Receipts[j].SurfaceID })
	sort.Slice(normalized.Failures, func(i, j int) bool {
		if normalized.Failures[i].Code != normalized.Failures[j].Code {
			return normalized.Failures[i].Code < normalized.Failures[j].Code
		}
		return normalized.Failures[i].Detail < normalized.Failures[j].Detail
	})
	for i := range normalized.SurfaceResults {
		normalized.SurfaceResults[i].ReasonCodes = sortedStrings(normalized.SurfaceResults[i].ReasonCodes)
	}
	return json.Marshal(normalized)
}

func (e CouplingEvidence) CanonicalJSONL() ([]byte, error) {
	data, err := e.CanonicalJSON()
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func (e CouplingEvidence) AppendCanonicalJSONL(w io.Writer) error {
	data, err := e.CanonicalJSONL()
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
