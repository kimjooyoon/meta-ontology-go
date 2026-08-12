package provenance

import (
	"strings"
	"time"
)

// ContractVersion identifies the reusable provenance storage contract.
const ContractVersion = 1

// ContractSpec describes the input/output boundary that future AST, IR,
// projection, cache, LSP, and CI adapters can implement or verify.
type ContractSpec struct {
	Version       int
	Format        string
	Input         InputContract
	Output        OutputContract
	Adapters      []AdapterContract
	HostingStages []HostingStage
	Hypotheses    []Hypothesis
	NegativeCases []NegativeCase
	Deferred      []DeferredCase
}

// InputContract defines the minimum fields a producer must provide.
type InputContract struct {
	RequiredFields  []string
	FreshnessFields []string
	BindingFields   []string
	IdentityRule    string
}

// OutputContract defines stable serialization, digest, and diagnostics.
type OutputContract struct {
	LineEncoding       string
	RecordHashRule     string
	SnapshotDigestRule string
	OrderingRule       string
	DiagnosticsRule    string
}

// AdapterContract maps an upstream or downstream implementation boundary to
// the evidence fields it must provide and preserve.
type AdapterContract struct {
	Name         string
	Input        string
	Output       string
	MustPreserve string
}

// HostingStage keeps the Go-hosted prototype distinct from future gooo-hosted
// self-hosting. A deferred stage is never a passing implementation result.
type HostingStage struct {
	Name             string
	Host             string
	Status           string
	RequiredEvidence string
}

// Hypothesis is a falsifiable claim tied to an executable fixture or test.
type Hypothesis struct {
	ID            string
	Claim         string
	Fixture       string
	PassCriterion string
	FailCriterion string
}

// NegativeCase records a mutation that must not be accepted as valid evidence.
type NegativeCase struct {
	ID           string
	Mutation     string
	ExpectedKind string
	Preservation string
}

// DeferredCase records a deliberately unimplemented boundary. Deferred work
// must not be reported as a passing implementation result.
type DeferredCase struct {
	ID         string
	Capability string
	Reason     string
}

// Fixture is the minimum deterministic input used by the contract tests.
type Fixture struct {
	Name          string
	SourceHash    string
	Records       []Evidence
	ExpectedOrder []string
}

// CurrentContract returns the current contract and its falsifiable evidence
// plan. Slices are newly allocated so callers can annotate their copy.
func CurrentContract() ContractSpec {
	return ContractSpec{
		Version: ContractVersion,
		Format:  "canonical JSON Lines, schema 1",
		Input: InputContract{
			RequiredFields:  []string{"id", "type", "subject", "generated_by", "freshness"},
			FreshnessFields: []string{"source_hash", "produced_at", "valid_until?"},
			BindingFields:   []string{"repository", "base", "head", "event_ref", "checkout_ref", "run_id", "run_attempt", "workflow", "six jobs(id, conclusion, head_sha)", "policy_digest", "toolchain_digest", "bundle_digest", "predecessors", "evidence_refs", "write_effect"},
			IdentityRule:    "id is unique within one store and is not inferred from display names",
		},
		Output: OutputContract{
			LineEncoding:       "one UTF-8 compact JSON object per LF-terminated line",
			RecordHashRule:     "sha256(canonical record JSON with hash omitted, including optional binding)",
			SnapshotDigestRule: "sha256(sorted canonical JSONL records with LF separators)",
			OrderingRule:       "Read returns records sorted lexicographically by stable id",
			DiagnosticsRule:    "corruption reports path, line, byte offset, kind, and detail",
		},
		Adapters: []AdapterContract{
			{Name: "AST→provenance", Input: "stable source identity, source hash, parse result", Output: "Evidence with generated_by parse activity and Freshness.SourceHash", MustPreserve: "same source hash and stable evidence id"},
			{Name: "IR/BX→provenance", Input: "semantic identity, relation delta, source hash", Output: "append-only Evidence entity", MustPreserve: "identity is not inferred from display-name changes"},
			{Name: "codegen→provenance", Input: "generated artifact digest and generator activity", Output: "evidence subject bound to artifact and source hash", MustPreserve: "artifact evidence cannot validate a different source snapshot"},
			{Name: "LSP→provenance", Input: "ReadOptions and a store path", Output: "sorted Snapshot or typed diagnostic", MustPreserve: "diagnostic line and byte offset"},
			{Name: "cache→provenance", Input: "cached artifact and expected source hash", Output: "fresh Snapshot or FreshnessError", MustPreserve: "stale cache is never returned as fresh"},
			{Name: "CI→provenance", Input: "allowed source hash and verification activity", Output: "pass evidence or deterministic failure", MustPreserve: "unimplemented stages remain deferred"},
		},
		HostingStages: []HostingStage{
			{Name: "go-hosted-provenance", Host: "Go 1.26.5 standard library", Status: "pass for this storage contract", RequiredEvidence: "go test, race, vet, and canonical fixture results"},
			{Name: "gooo-hosted-self-hosting", Host: ".gooo compiler", Status: "deferred", RequiredEvidence: "implemented gooo check, DSL↔IR round-trip, and independent self-hosting evidence"},
		},
		Hypotheses: []Hypothesis{
			{ID: "H1", Claim: "append order does not change the semantic snapshot", Fixture: "minimal-two-records", PassCriterion: "same sorted IDs and snapshot digest", FailCriterion: "different order or digest for a permutation"},
			{ID: "H2", Claim: "a line mutation cannot be read as valid evidence", Fixture: "minimal-two-records", PassCriterion: "Read returns CorruptionError with a precise kind", FailCriterion: "tampered data yields a successful snapshot"},
			{ID: "H3", Claim: "append-only identity prevents destructive duplicate writes", Fixture: "minimal-two-records", PassCriterion: "duplicate append fails and file bytes are unchanged", FailCriterion: "duplicate append mutates or validates"},
			{ID: "H4", Claim: "freshness policy rejects a wrong or expired source", Fixture: "minimal-two-records", PassCriterion: "Read returns FreshnessError when checks are requested", FailCriterion: "stale evidence is returned as fresh"},
			{ID: "H5", Claim: "a protected receipt is bound to one complete run tuple", Fixture: "graph-proof-007-binding", PassCriterion: "missing, internally inconsistent, or expected-tuple-mismatched binding fails closed", FailCriterion: "a protected record validates without its exact tuple"},
			{ID: "H6", Claim: "receipt predecessor claims are append-only and non-replayable", Fixture: "graph-proof-007-binding", PassCriterion: "duplicate predecessor returns ReplayError before any write", FailCriterion: "a predecessor is accepted twice or bytes change"},
		},
		NegativeCases: []NegativeCase{
			{ID: "N1", Mutation: "replace a record hash", ExpectedKind: "hash-mismatch", Preservation: "no successful snapshot"},
			{ID: "N2", Mutation: "append an existing id", ExpectedKind: "duplicate-id", Preservation: "file bytes unchanged"},
			{ID: "N3", Mutation: "write malformed JSON or omit LF", ExpectedKind: "invalid-json or missing-newline", Preservation: "no successful snapshot"},
			{ID: "N4", Mutation: "read with wrong source or after valid_until", ExpectedKind: "source-mismatch or expired", Preservation: "FreshnessError, not success"},
			{ID: "N5", Mutation: "tamper a canonical job head_sha or conclusion", ExpectedKind: "binding-invalid", Preservation: "no successful snapshot"},
			{ID: "N6", Mutation: "supply a different but internally valid expected tuple", ExpectedKind: "binding-mismatch", Preservation: "no successful snapshot"},
			{ID: "N7", Mutation: "append a second receipt with an existing predecessor", ExpectedKind: "replayed-predecessor", Preservation: "file bytes unchanged"},
			{ID: "N8", Mutation: "omit binding or set write_effect to nonzero", ExpectedKind: "binding-invalid", Preservation: "file bytes unchanged"},
		},
		Deferred: []DeferredCase{
			{ID: "D1", Capability: "cross-process append locking", Reason: "Store serializes one in-process instance only; an OS-level lock contract is not implemented"},
			{ID: "D2", Capability: "full RFC 8785 JSON Canonicalization Scheme", Reason: "The profile relies on compact encoding/json with sorted map keys and must not claim full JCS compliance"},
			{ID: "D3", Capability: "AST/IR/BX/codegen/LSP/cache adapters", Reason: "No upstream semantic packages are present in this baseline; adapters need a later integration contract"},
			{ID: "D4", Capability: "gooo-hosted self-hosting stage", Reason: "The gooo check stage is unimplemented and remains deferred"},
			{ID: "D5", Capability: "external provenance provider attestation", Reason: "No external provenance connector or authoritative remote receipt is available in this lane"},
		},
	}
}

// MinimalFixture returns two records with the same source and deterministic
// freshness windows. The input is intentionally returned in non-sorted order
// so ordering invariance is observable rather than assumed.
func MinimalFixture() Fixture {
	sourceHash := strings.Repeat("1", 64)
	produced := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)
	validUntil := produced.Add(2 * time.Hour)
	return Fixture{
		Name:          "minimal-two-records",
		SourceHash:    sourceHash,
		Records:       []Evidence{fixtureRecord("evidence/b", sourceHash, produced, validUntil), fixtureRecord("evidence/a", sourceHash, produced, validUntil)},
		ExpectedOrder: []string{"evidence/a", "evidence/b"},
	}
}

func fixtureRecord(id, sourceHash string, produced, validUntil time.Time) Evidence {
	return Evidence{
		ID:          id,
		Type:        "TestResult",
		Subject:     "artifact/minimal",
		GeneratedBy: "activity/verify",
		Attributes:  map[string]string{"fixture": "minimal", "status": "passed"},
		Freshness:   NewFreshness(sourceHash, produced, validUntil),
	}
}
