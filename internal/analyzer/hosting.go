package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

// HostStage identifies which implementation is producing compiler evidence.
type HostStage string

const (
	// StageGoHosted is the implemented reference stage. Go hosts analysis and
	// .gooo remains the authoritative source view.
	StageGoHosted HostStage = "go-hosted"
	// StageGoooHosted is the future self-hosted stage. It is intentionally
	// deferred until an independent comparison contract is implemented.
	StageGoooHosted HostStage = "gooo-hosted"
)

// ContractStatus distinguishes an implemented contract from an explicitly
// deferred future stage. Deferred is never treated as a successful stage.
type ContractStatus string

const (
	ContractImplemented ContractStatus = "implemented"
	ContractDeferred    ContractStatus = "deferred"
)

// ContractRequirement names evidence a host stage must provide before it can
// be promoted.
type ContractRequirement string

const (
	RequirementStableIdentity  ContractRequirement = "stable-identity"
	RequirementDeltaEvidence   ContractRequirement = "delta-evidence"
	RequirementSourceSpans     ContractRequirement = "source-spans"
	RequirementHostComparison  ContractRequirement = "host-comparison"
	RequirementIndependentGate ContractRequirement = "independent-gate"
)

// HostingContract is the comparable contract metadata for one host stage.
type HostingContract struct {
	Stage           HostStage
	Status          ContractStatus
	SourceAuthority string
	Producer        string
	Requirements    []ContractRequirement
}

// ContractFor returns the declared contract for a host stage. Future stages
// are returned as deferred contracts, never as successful implementations.
func ContractFor(stage HostStage) HostingContract {
	switch stage {
	case StageGoHosted:
		return HostingContract{
			Stage:           StageGoHosted,
			Status:          ContractImplemented,
			SourceAuthority: ".gooo",
			Producer:        "go",
			Requirements: []ContractRequirement{
				RequirementStableIdentity,
				RequirementDeltaEvidence,
				RequirementSourceSpans,
			},
		}
	case StageGoooHosted:
		return HostingContract{
			Stage:           StageGoooHosted,
			Status:          ContractDeferred,
			SourceAuthority: ".gooo",
			Producer:        "gooo",
			Requirements: []ContractRequirement{
				RequirementStableIdentity,
				RequirementDeltaEvidence,
				RequirementSourceSpans,
				RequirementHostComparison,
				RequirementIndependentGate,
			},
		}
	default:
		return HostingContract{Stage: stage, Status: ContractDeferred}
	}
}

// Valid reports whether the contract has a known stage and complete metadata.
func (c HostingContract) Valid() bool {
	if !c.Stage.Valid() || !validContractStatus(c.Status) || strings.TrimSpace(c.SourceAuthority) == "" || strings.TrimSpace(c.Producer) == "" {
		return false
	}
	return len(c.Requirements) > 0
}

// PromotionReady reports whether this contract can be treated as implemented.
func (c HostingContract) PromotionReady() bool {
	return c.Valid() && c.Status == ContractImplemented
}

func (s HostStage) Valid() bool {
	return s == StageGoHosted || s == StageGoooHosted
}

func validContractStatus(status ContractStatus) bool {
	return status == ContractImplemented || status == ContractDeferred
}

// EvidenceKind identifies which analyzer view a record preserves.
type EvidenceKind string

const (
	EvidenceKindFact           EvidenceKind = "fact"
	EvidenceKindCandidate      EvidenceKind = "candidate"
	EvidenceKindImplementation EvidenceKind = "implementation"
)

// EvidenceStatus identifies the confidence state of one evidence record.
type EvidenceStatus string

const (
	EvidenceStatusDeterministic  EvidenceStatus = "deterministic"
	EvidenceStatusCandidate      EvidenceStatus = "candidate"
	EvidenceStatusImplementation EvidenceStatus = "implementation"
)

// EvidenceRecord is a host-neutral, append-only projection of one analyzer
// result. Producer and stage live on EvidenceReport so records can compare
// across Go-hosted and future gooo-hosted runs.
type EvidenceRecord struct {
	Kind      EvidenceKind
	Status    EvidenceStatus
	Subject   Identity
	Relation  Relation
	Object    Identity
	Reference string
	Options   []Identity
	Span      Span
	Reason    string
}

// Valid reports whether the record has the fields required for its kind.
func (e EvidenceRecord) Valid() bool {
	switch e.Kind {
	case EvidenceKindFact:
		return e.Status == EvidenceStatusDeterministic && e.Subject.Valid() && e.Object.Valid() && e.Relation != "" && evidenceSpanValid(e.Span)
	case EvidenceKindCandidate:
		return e.Status == EvidenceStatusCandidate && e.Subject.Valid() && e.Relation != "" && e.Reference != "" && validIdentityOptions(e.Options) && evidenceSpanValid(e.Span)
	case EvidenceKindImplementation:
		return e.Status == EvidenceStatusImplementation && e.Reference != "" && evidenceSpanValid(e.Span)
	default:
		return false
	}
}

// EvidenceReport is the evidence emitted by one host stage. A deferred report
// has no comparison digest and cannot be mistaken for a successful run.
type EvidenceReport struct {
	Contract HostingContract
	Records  []EvidenceRecord
	Reason   string
}

// Complete reports whether the host contract is implemented and every record
// is structurally valid.
func (r EvidenceReport) Complete() bool {
	if !r.Contract.PromotionReady() {
		return false
	}
	for _, record := range r.Records {
		if !record.Valid() {
			return false
		}
	}
	return true
}

// ComparisonCanonical omits host metadata while retaining semantic evidence,
// making equivalent reports comparable across host implementations.
func (r EvidenceReport) ComparisonCanonical() string {
	if !r.Complete() {
		return ""
	}
	records := append([]EvidenceRecord(nil), r.Records...)
	sort.Slice(records, func(i, j int) bool {
		return records[i].comparisonCanonical() < records[j].comparisonCanonical()
	})
	var builder strings.Builder
	builder.WriteString("analyzer-evidence/v1\n")
	for _, record := range records {
		builder.WriteString(record.comparisonCanonical())
		builder.WriteByte('\n')
	}
	return builder.String()
}

// ComparisonDigest returns a stable digest only for a complete report.
func (r EvidenceReport) ComparisonDigest() string {
	canonical := r.ComparisonCanonical()
	if canonical == "" {
		return ""
	}
	digest := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(digest[:])
}

// GoHostedEvidence projects the current analyzer result into implemented
// reference-host evidence.
func (r Result) GoHostedEvidence() EvidenceReport {
	report := EvidenceReport{Contract: ContractFor(StageGoHosted)}
	for _, fact := range r.Delta.Added {
		report.Records = append(report.Records, EvidenceRecord{
			Kind: EvidenceKindFact, Status: EvidenceStatusDeterministic,
			Subject: fact.Subject, Relation: fact.Relation, Object: fact.Object, Span: fact.Span,
		})
	}
	for _, candidate := range r.Delta.Candidates {
		report.Records = append(report.Records, EvidenceRecord{
			Kind: EvidenceKindCandidate, Status: EvidenceStatusCandidate,
			Subject: candidate.Subject, Relation: candidate.Relation, Reference: candidate.Reference,
			Options: append([]Identity(nil), candidate.Options...), Span: candidate.Span, Reason: candidate.Reason,
		})
	}
	for _, detail := range r.Delta.ImplementationDetails {
		report.Records = append(report.Records, EvidenceRecord{
			Kind: EvidenceKindImplementation, Status: EvidenceStatusImplementation,
			Reference: detail.Reference, Span: detail.Span, Reason: detail.Reason,
		})
	}
	sortEvidenceRecords(report.Records)
	return report
}

// GoooHostedEvidence is an explicit deferred report until a gooo-hosted
// analyzer and independent comparison gate exist.
func (r Result) GoooHostedEvidence() EvidenceReport {
	return EvidenceReport{
		Contract: ContractFor(StageGoooHosted),
		Reason:   "gooo-hosted analyzer is not implemented",
	}
}

func sortEvidenceRecords(records []EvidenceRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].comparisonCanonical() < records[j].comparisonCanonical()
	})
}

func (e EvidenceRecord) comparisonCanonical() string {
	var builder strings.Builder
	writeEvidenceField(&builder, string(e.Kind))
	writeEvidenceField(&builder, string(e.Status))
	writeEvidenceField(&builder, e.Subject.Namespace)
	writeEvidenceField(&builder, e.Subject.ID)
	writeEvidenceField(&builder, string(e.Relation))
	writeEvidenceField(&builder, e.Object.Namespace)
	writeEvidenceField(&builder, e.Object.ID)
	writeEvidenceField(&builder, e.Reference)
	options := append([]Identity(nil), e.Options...)
	sort.Slice(options, func(i, j int) bool { return identityLess(options[i], options[j]) })
	for _, option := range options {
		writeEvidenceField(&builder, option.Namespace)
		writeEvidenceField(&builder, option.ID)
	}
	writeEvidenceField(&builder, e.Span.Filename)
	writeEvidenceField(&builder, strconv.Itoa(e.Span.Start.Offset))
	writeEvidenceField(&builder, strconv.Itoa(e.Span.Start.Line))
	writeEvidenceField(&builder, strconv.Itoa(e.Span.Start.Column))
	writeEvidenceField(&builder, strconv.Itoa(e.Span.End.Offset))
	writeEvidenceField(&builder, strconv.Itoa(e.Span.End.Line))
	writeEvidenceField(&builder, strconv.Itoa(e.Span.End.Column))
	writeEvidenceField(&builder, e.Reason)
	return builder.String()
}

func validIdentityOptions(options []Identity) bool {
	if len(options) == 0 {
		return false
	}
	for _, option := range options {
		if !option.Valid() {
			return false
		}
	}
	return true
}

func evidenceSpanValid(span Span) bool {
	return span.Filename != "" && span.Start.Offset >= 0 && span.End.Offset > span.Start.Offset && span.Start.Line > 0 && span.End.Line > 0 && span.Start.Column > 0 && span.End.Column > 0
}

func writeEvidenceField(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
	builder.WriteByte('|')
}
