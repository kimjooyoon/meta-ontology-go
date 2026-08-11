package semantic

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// Stable host identities let evidence from two compiler implementations be
// compared without treating the implementation itself as semantic meaning.
const (
	GoHostedCompilerID   ID = "gooo://host/compiler/go"
	GoooHostedCompilerID ID = "gooo://host/compiler/gooo"
	GoVerifierID         ID = "gooo://host/verifier/go"
	CIVerifierID            = GoVerifierID

	GoHostedCompiler   = GoHostedCompilerID
	GoooHostedCompiler = GoooHostedCompilerID
	GoVerifier         = GoVerifierID
	CIVerifier         = GoVerifierID
)

// EvidenceKind identifies the producer contract for an evidence record.
type EvidenceKind string

const (
	CompilerRunEvidence  EvidenceKind = "compiler-run"
	VerificationEvidence EvidenceKind = "verification"
	ComparisonEvidence   EvidenceKind = "comparison"
)

func (k EvidenceKind) Valid() bool {
	switch k {
	case CompilerRunEvidence, VerificationEvidence, ComparisonEvidence:
		return true
	default:
		return false
	}
}

func (k EvidenceKind) String() string {
	return string(k)
}

// Evidence is an append-only audit record for a semantic fact. ID is stable
// across hosts; Producer records which host emitted the record. Digest binds
// the claim to the source or verification artifact without changing IR
// semantic equivalence.
type Evidence struct {
	ID       ID
	Producer ID
	Kind     EvidenceKind
	Fact     FactKey
	Status   FactStatus
	Digest   string
	Span     Span
}

func NewEvidence(id, producer ID, kind EvidenceKind, fact FactKey, digest string) (Evidence, error) {
	return (Evidence{ID: id, Producer: producer, Kind: kind, Fact: fact, Digest: digest}).Normalized()
}

func (e Evidence) Normalized() (Evidence, error) {
	id, err := ParseIdentity(e.ID.String())
	if err != nil {
		return Evidence{}, fmt.Errorf("%w: id: %v", ErrInvalidEvidence, err)
	}
	producer, err := ParseIdentity(e.Producer.String())
	if err != nil {
		return Evidence{}, fmt.Errorf("%w: producer: %v", ErrInvalidEvidence, err)
	}
	fact, err := normalizeFactKey(e.Fact)
	if err != nil {
		return Evidence{}, fmt.Errorf("%w: fact: %v", ErrInvalidEvidence, err)
	}
	if !e.Kind.Valid() {
		return Evidence{}, fmt.Errorf("%w: unknown kind %q", ErrInvalidEvidence, e.Kind)
	}
	status := e.Status
	if status == 0 {
		status = FactDeterministic
	}
	if status != FactDeterministic && status != FactCandidate {
		return Evidence{}, fmt.Errorf("%w: unknown status %d", ErrInvalidEvidence, status)
	}
	digest, err := normalizeDigest(e.Digest)
	if err != nil {
		return Evidence{}, err
	}
	span := e.Span.Normalized()
	if err := span.Validate(); err != nil {
		return Evidence{}, fmt.Errorf("%w: span: %v", ErrInvalidEvidence, err)
	}
	e.ID = id
	e.Producer = producer
	e.Fact = fact
	e.Status = status
	e.Digest = digest
	e.Span = span
	return e, nil
}

func (e Evidence) Validate() error {
	_, err := e.Normalized()
	return err
}

func (e Evidence) WithSpan(span Span) Evidence {
	e.Span = span
	return e
}

// ValidateAgainst ensures evidence supports a fact already present in the IR.
func (e Evidence) ValidateAgainst(graph Graph) error {
	normalized, err := e.Normalized()
	if err != nil {
		return err
	}
	if normalized.Status == FactCandidate {
		if !graph.HasCandidate(normalized.Fact) && !graph.HasFact(normalized.Fact) {
			return fmt.Errorf("%w: candidate fact is not present", ErrInvalidEvidence)
		}
		// Promotion changes the graph fact status explicitly; it does not
		// reclassify or erase the append-only candidate evidence record.
		return nil
	}
	if !graph.HasFact(normalized.Fact) {
		return fmt.Errorf("%w: deterministic fact is not present", ErrInvalidEvidence)
	}
	return nil
}

// ValidateFresh checks that the evidence digest still names the pinned
// payload used to produce the record. It does not decide whether the claim is
// authoritative; that policy remains with the independent Go verifier.
func (e Evidence) ValidateFresh(payload []byte) error {
	normalized, err := e.Normalized()
	if err != nil {
		return err
	}
	expected := StableHash(payload)
	if normalized.Digest != expected {
		return fmt.Errorf("%w: got %s, want %s", ErrStaleEvidence, normalized.Digest, expected)
	}
	return nil
}

func normalizeDigest(raw string) (string, error) {
	digest := strings.ToLower(strings.TrimSpace(raw))
	if len(digest) != sha256.Size*2 {
		return "", fmt.Errorf("%w: digest must be a SHA-256 hex value", ErrInvalidEvidence)
	}
	if _, err := hex.DecodeString(digest); err != nil {
		return "", fmt.Errorf("%w: digest: %v", ErrInvalidEvidence, err)
	}
	return digest, nil
}

func (e Evidence) Canonical() string {
	if normalized, err := e.Normalized(); err == nil {
		e = normalized
	}
	var b strings.Builder
	b.WriteString("evidence\t")
	writeCanonicalField(&b, e.ID.String())
	writeCanonicalField(&b, e.Producer.String())
	writeCanonicalField(&b, e.Kind.String())
	writeCanonicalField(&b, e.Status.String())
	writeCanonicalField(&b, e.Fact.Subject.String())
	writeCanonicalField(&b, e.Fact.Predicate.String())
	writeCanonicalField(&b, e.Fact.Object.String())
	writeCanonicalField(&b, e.Digest)
	writeCanonicalSpan(&b, e.Span)
	return b.String()
}

// ComparisonCanonical excludes host and source-location metadata. It is the
// cross-host provenance claim used to compare Go-hosted and gooo-hosted runs.
func (e Evidence) ComparisonCanonical() string {
	if normalized, err := e.Normalized(); err == nil {
		e = normalized
	}
	var b strings.Builder
	b.WriteString("evidence-comparison\t")
	writeCanonicalField(&b, e.ID.String())
	writeCanonicalField(&b, e.Kind.String())
	writeCanonicalField(&b, e.Status.String())
	writeCanonicalField(&b, e.Fact.Subject.String())
	writeCanonicalField(&b, e.Fact.Predicate.String())
	writeCanonicalField(&b, e.Fact.Object.String())
	writeCanonicalField(&b, e.Digest)
	return b.String()
}

func (e Evidence) StableHash() string {
	return StableHashString(e.Canonical())
}

func (e Evidence) ProvenanceHash() string {
	return StableHashString(e.ComparisonCanonical())
}

func sortEvidence(evidence []Evidence) {
	sort.Slice(evidence, func(i, j int) bool {
		if evidence[i].ID != evidence[j].ID {
			return evidence[i].ID < evidence[j].ID
		}
		return evidence[i].Canonical() < evidence[j].Canonical()
	})
}
