package analyzer

import (
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const semanticNormalizedDeltaSchema = "analyzer-semantic-delta/v1"

// DeltaBinding is the common identity tuple for one source-backed handoff.
// SourceDigest covers the exact generated source bytes; the other fields pin
// the semantic base and the producer contract used to interpret them.
type DeltaBinding struct {
	SourceDigest    string `json:"source_digest"`
	BaseDigest      string `json:"base_digest"`
	PolicyDigest    string `json:"policy_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
}

func (b DeltaBinding) canonical() string {
	var builder strings.Builder
	writeBindingField(&builder, b.SourceDigest)
	writeBindingField(&builder, b.BaseDigest)
	writeBindingField(&builder, b.PolicyDigest)
	writeBindingField(&builder, b.ToolchainDigest)
	return builder.String()
}

func (b DeltaBinding) complete() bool {
	return validDigest(b.SourceDigest) && validDigest(b.BaseDigest) &&
		validDigest(b.PolicyDigest) && validDigest(b.ToolchainDigest)
}

// NormalizedSignatureFact is an authoritative, typed signature fact and its
// semantic evidence. It is the only delta member eligible for IR authority.
type NormalizedSignatureFact struct {
	Binding        DeltaBinding      `json:"binding"`
	SourceRelation Relation          `json:"source_relation"`
	Fact           semantic.Fact     `json:"fact"`
	Evidence       semantic.Evidence `json:"evidence"`
}

func (f NormalizedSignatureFact) canonical() string {
	var builder strings.Builder
	builder.WriteString("signature\n")
	builder.WriteString(f.Binding.canonical())
	writeBindingField(&builder, string(f.SourceRelation))
	builder.WriteString(f.Fact.Canonical())
	builder.WriteString(f.Evidence.Canonical())
	return builder.String()
}

// NormalizedCandidateFact records an unresolved source relation. Facts and
// evidence are populated only when an explicit policy mapped the options;
// the candidate remains separate from deterministic graph facts either way.
type NormalizedCandidateFact struct {
	Binding        DeltaBinding        `json:"binding"`
	SourceRelation Relation            `json:"source_relation"`
	Origin         ObservationOrigin   `json:"origin"`
	Subject        semantic.ID         `json:"subject"`
	Options        []semantic.ID       `json:"options"`
	Facts          []semantic.Fact     `json:"facts"`
	Evidence       []semantic.Evidence `json:"evidence"`
	Span           semantic.Span       `json:"span"`
	Reason         string              `json:"reason"`
}

func (f NormalizedCandidateFact) canonical() string {
	var builder strings.Builder
	builder.WriteString("candidate\n")
	builder.WriteString(f.Binding.canonical())
	writeBindingField(&builder, string(f.SourceRelation))
	writeBindingField(&builder, string(f.Origin))
	writeBindingField(&builder, f.Subject.String())
	for _, option := range f.Options {
		writeBindingField(&builder, option.String())
	}
	for _, fact := range f.Facts {
		builder.WriteString(fact.Canonical())
	}
	for _, evidence := range f.Evidence {
		builder.WriteString(evidence.Canonical())
	}
	writeSemanticSpan(&builder, f.Span)
	writeBindingField(&builder, f.Reason)
	return builder.String()
}

// SemanticNormalizedDelta is the machine-readable handoff boundary. The
// three slices deliberately cannot be confused by a status bit or a string.
type SemanticNormalizedDelta struct {
	SchemaVersion          string                         `json:"schema_version"`
	SignatureFacts         []NormalizedSignatureFact      `json:"signature_facts"`
	CandidateFacts         []NormalizedCandidateFact      `json:"candidate_facts"`
	DeferredImplementation []ImplementationObservation    `json:"deferred_implementation"`
	DeferredDetails        []DeferredImplementationDetail `json:"deferred_details"`
	Digest                 string                         `json:"digest"`
}

// Canonical returns an order-independent representation of the typed delta.
func (d SemanticNormalizedDelta) Canonical() string {
	var builder strings.Builder
	builder.WriteString(d.SchemaVersion)
	builder.WriteByte('\n')
	signatures := append([]NormalizedSignatureFact(nil), d.SignatureFacts...)
	sort.Slice(signatures, func(i, j int) bool { return signatures[i].canonical() < signatures[j].canonical() })
	for _, fact := range signatures {
		builder.WriteString(fact.canonical())
	}
	candidates := append([]NormalizedCandidateFact(nil), d.CandidateFacts...)
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].canonical() < candidates[j].canonical() })
	for _, candidate := range candidates {
		builder.WriteString(candidate.canonical())
	}
	observations := append([]ImplementationObservation(nil), d.DeferredImplementation...)
	sort.Slice(observations, func(i, j int) bool { return observations[i].Canonical() < observations[j].Canonical() })
	for _, observation := range observations {
		builder.WriteString("deferred\n")
		builder.WriteString(observation.Canonical())
	}
	details := append([]DeferredImplementationDetail(nil), d.DeferredDetails...)
	sort.Slice(details, func(i, j int) bool { return details[i].canonical() < details[j].canonical() })
	for _, detail := range details {
		builder.WriteString(detail.canonical())
	}
	return builder.String()
}

// StableHash is the digest used to compare normalized deltas across runs.
func (d SemanticNormalizedDelta) StableHash() string {
	return semantic.StableHashString(d.Canonical())
}

func newSemanticNormalizedDelta(
	input SemanticAdapterInput, baseDigest string, result SemanticAdapterResult,
) (SemanticNormalizedDelta, error) {
	delta := SemanticNormalizedDelta{SchemaVersion: semanticNormalizedDeltaSchema}
	binding := DeltaBinding{
		SourceDigest: input.SourceDigest, BaseDigest: baseDigest,
		PolicyDigest: input.Policy.Digest(), ToolchainDigest: input.ToolchainDigest,
	}
	delta.SignatureFacts = normalizedSignatureFacts(input, result, binding)
	var err error
	delta.CandidateFacts, err = normalizedCandidateFacts(input, result, binding)
	if err != nil {
		return SemanticNormalizedDelta{}, err
	}
	delta.DeferredImplementation = append([]ImplementationObservation(nil), result.ImplementationObservations...)
	delta.DeferredDetails = deferredImplementationDetails(result, binding)
	delta.Digest = delta.StableHash()
	return delta, validateDeltaShape(delta)
}

func normalizedSignatureFacts(
	input SemanticAdapterInput, result SemanticAdapterResult, binding DeltaBinding,
) []NormalizedSignatureFact {
	output := make([]NormalizedSignatureFact, 0)
	for _, sourceFact := range input.Analysis.Delta.Added {
		if sourceFact.Origin != OriginSignature {
			continue
		}
		mapping, ok := input.Policy.lookup(sourceFact.Relation)
		if !ok || !mapping.allowsOrigin(sourceFact.Origin) {
			continue
		}
		mapped, err := mapFact(result.IR.Graph, sourceFact.Subject, sourceFact.Object, mapping, sourceFact.Span)
		if err != nil {
			continue
		}
		evidence, ok := evidenceForFact(result.IR.Evidence(), mapped.Key(), semantic.FactDeterministic)
		if !ok {
			continue
		}
		output = append(output, NormalizedSignatureFact{
			Binding: binding, SourceRelation: sourceFact.Relation, Fact: mapped, Evidence: evidence,
		})
	}
	sort.Slice(output, func(i, j int) bool { return output[i].canonical() < output[j].canonical() })
	return output
}

func normalizedCandidateFacts(
	input SemanticAdapterInput, result SemanticAdapterResult, binding DeltaBinding,
) ([]NormalizedCandidateFact, error) {
	output := make([]NormalizedCandidateFact, 0, len(input.Analysis.Delta.Candidates))
	for _, sourceCandidate := range input.Analysis.Delta.Candidates {
		candidate := NormalizedCandidateFact{
			Binding: binding, SourceRelation: sourceCandidate.Relation,
			Origin: sourceCandidate.Origin, Span: semanticSpan(sourceCandidate.Span),
			Reason: sourceCandidate.Reason,
		}
		subject, err := semantic.ParseIdentity(sourceCandidate.Subject.ID)
		if err != nil {
			return nil, err
		}
		candidate.Subject = subject
		for _, option := range sourceCandidate.Options {
			identity, err := semantic.ParseIdentity(option.ID)
			if err != nil {
				return nil, err
			}
			candidate.Options = append(candidate.Options, identity)
		}
		mapping, mapped := input.Policy.lookup(sourceCandidate.Relation)
		if !mapped || !mapping.allowsOrigin(sourceCandidate.Origin) {
			sort.Slice(candidate.Options, func(i, j int) bool { return candidate.Options[i] < candidate.Options[j] })
			output = append(output, candidate)
			continue
		}
		for _, option := range sourceCandidate.Options {
			fact, err := mapFact(result.IR.Graph, sourceCandidate.Subject, option, mapping, sourceCandidate.Span)
			if err != nil {
				continue
			}
			fact.Status = semantic.FactCandidate
			fact.Reason = sourceCandidate.Reason
			candidate.Facts = append(candidate.Facts, fact)
			if evidence, ok := evidenceForFact(result.IR.Evidence(), fact.Key(), semantic.FactCandidate); ok {
				candidate.Evidence = append(candidate.Evidence, evidence)
			} else if evidence, ok := shadowEvidenceForFact(result.ShadowedCandidateEvidence, fact.Key()); ok {
				candidate.Evidence = append(candidate.Evidence, evidence)
			}
		}
		sortNormalizedCandidate(&candidate)
		output = append(output, candidate)
	}
	sort.Slice(output, func(i, j int) bool { return output[i].canonical() < output[j].canonical() })
	return output, nil
}

func sortNormalizedCandidate(candidate *NormalizedCandidateFact) {
	sort.Slice(candidate.Options, func(i, j int) bool { return candidate.Options[i] < candidate.Options[j] })
	sort.Slice(candidate.Facts, func(i, j int) bool {
		return candidate.Facts[i].Canonical() < candidate.Facts[j].Canonical()
	})
	sort.Slice(candidate.Evidence, func(i, j int) bool {
		return candidate.Evidence[i].Canonical() < candidate.Evidence[j].Canonical()
	})
}

func evidenceForFact(
	records []semantic.Evidence, key semantic.FactKey, status semantic.FactStatus,
) (semantic.Evidence, bool) {
	for _, record := range records {
		if record.Fact == key && record.Status == status {
			return record, true
		}
	}
	return semantic.Evidence{}, false
}

func shadowEvidenceForFact(records []semantic.Evidence, key semantic.FactKey) (semantic.Evidence, bool) {
	return evidenceForFact(records, key, semantic.FactCandidate)
}
