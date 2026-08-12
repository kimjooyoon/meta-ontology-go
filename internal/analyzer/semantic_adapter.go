package analyzer

import (
	"encoding/hex"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// SemanticAdapterInput supplies an analyzer result and an immutable semantic
// base. Base is normalized into a private transaction before any additions.
type SemanticAdapterInput struct {
	Base            semantic.IR
	Analysis        Result
	Policy          MappingPolicy
	Producer        semantic.ID
	EvidenceKind    semantic.EvidenceKind
	SourceDigest    string
	ToolchainDigest string
}

// SemanticAdapterResult keeps unmapped observations and implementation detail
// outside the authoritative semantic graph. Candidates are added only through
// Graph.AddCandidate when their explicit mapping and endpoints are valid.
type SemanticAdapterResult struct {
	IR                              semantic.IR
	SourceDigest                    string
	PolicyDigest                    string
	ToolchainDigest                 string
	BindingDigest                   string
	ImplementationObservationDigest string
	DeferredFacts                   []Fact
	DeferredCandidates              []Candidate
	// ShadowedCandidateEvidence retains a mapped candidate observation when
	// the same FactKey is already deterministic in the base graph. The
	// evidence is intentionally not added to IR: candidate evidence cannot
	// stand in for authoritative evidence, and the semantic IR rejects a
	// candidate evidence record without a candidate fact. Keeping it here
	// preserves the historical observation without changing authority.
	ShadowedCandidateEvidence  []semantic.Evidence
	ImplementationDetails      []ImplementationDetail
	ImplementationObservations []ImplementationObservation
}

// AdaptSemantic performs a transactional, explicit mapping. The input IR is
// never mutated, including when a later observation fails preflight.
func AdaptSemantic(input SemanticAdapterInput) (SemanticAdapterResult, error) {
	if err := input.Policy.Validate(); err != nil {
		return SemanticAdapterResult{}, err
	}
	base, err := input.Base.Normalized()
	if err != nil {
		return SemanticAdapterResult{}, err
	}
	if err := validateObservations(input.Analysis); err != nil {
		return SemanticAdapterResult{}, err
	}
	transaction := SemanticAdapterResult{
		IR: base, SourceDigest: input.SourceDigest, PolicyDigest: input.Policy.Digest(),
		ToolchainDigest:       input.ToolchainDigest,
		DeferredCandidates:    copyCandidates(input.Analysis.Delta.Candidates),
		ImplementationDetails: copyDetails(input.Analysis.Delta.ImplementationDetails),
	}
	transaction.ImplementationObservations = collectImplementationObservations(input.Analysis, base, input)
	transaction.ImplementationObservationDigest = implementationObservationDigest(transaction.ImplementationObservations)
	if err := addRegisteredNodes(&transaction.IR, input.Analysis.Registrations); err != nil {
		return SemanticAdapterResult{}, err
	}
	if hasMappedObservation(input.Analysis, input.Policy) {
		if err := validateEvidenceConfig(input); err != nil {
			return SemanticAdapterResult{}, err
		}
	}
	if err := adaptFacts(&transaction, input); err != nil {
		return SemanticAdapterResult{}, err
	}
	if err := adaptCandidates(&transaction, input); err != nil {
		return SemanticAdapterResult{}, err
	}
	if err := transaction.IR.Validate(); err != nil {
		return SemanticAdapterResult{}, err
	}
	transaction.BindingDigest = semanticAdapterBindingDigest(transaction)
	return transaction, nil
}

func validateObservations(result Result) error {
	for _, fact := range result.Delta.Added {
		if !knownAnalyzerRelation(fact.Relation) {
			return adapterError(AdapterUnknownRelation, fact.Relation, "", "fact relation is not analyzer vocabulary")
		}
	}
	for _, candidate := range result.Delta.Candidates {
		if !knownAnalyzerRelation(candidate.Relation) {
			return adapterError(AdapterUnknownRelation, candidate.Relation, "", "candidate relation is not analyzer vocabulary")
		}
	}
	return nil
}

func hasMappedObservation(result Result, policy MappingPolicy) bool {
	for _, fact := range result.Delta.Added {
		if _, ok := policy.lookup(fact.Relation); ok {
			return true
		}
	}
	for _, candidate := range result.Delta.Candidates {
		if _, ok := policy.lookup(candidate.Relation); ok && len(candidate.Options) > 0 {
			return true
		}
	}
	return false
}

func adaptFacts(result *SemanticAdapterResult, input SemanticAdapterInput) error {
	for _, fact := range input.Analysis.Delta.Added {
		mapping, ok := input.Policy.lookup(fact.Relation)
		if !ok {
			result.DeferredFacts = append(result.DeferredFacts, fact)
			continue
		}
		if !mapping.allowsOrigin(fact.Origin) {
			result.DeferredFacts = append(result.DeferredFacts, fact)
			continue
		}
		mapped, err := mapFact(result.IR.Graph, fact.Subject, fact.Object, mapping, fact.Span)
		if err != nil {
			return err
		}
		if err := result.IR.AddFact(mapped); err != nil {
			return err
		}
		evidence, err := mappedEvidence(input, fact.Relation, mapped, semantic.FactDeterministic)
		if err != nil {
			return err
		}
		if err := result.IR.AddEvidence(evidence); err != nil {
			return err
		}
	}
	return nil
}

func adaptCandidates(result *SemanticAdapterResult, input SemanticAdapterInput) error {
	for _, candidate := range input.Analysis.Delta.Candidates {
		mapping, ok := input.Policy.lookup(candidate.Relation)
		if !ok {
			continue
		}
		for _, option := range candidate.Options {
			mapped, err := mapFact(result.IR.Graph, candidate.Subject, option, mapping, candidate.Span)
			if err != nil {
				return err
			}
			mapped.Status = semantic.FactCandidate
			mapped.Reason = candidate.Reason
			evidence, err := mappedEvidence(input, candidate.Relation, mapped, semantic.FactCandidate)
			if err != nil {
				return err
			}
			if result.IR.Graph.HasFact(mapped.Key()) {
				result.ShadowedCandidateEvidence = append(result.ShadowedCandidateEvidence, evidence)
				continue
			}
			if err := result.IR.AddCandidate(mapped); err != nil {
				return err
			}
			if err := result.IR.AddEvidence(evidence); err != nil {
				return err
			}
		}
	}
	sort.Slice(result.ShadowedCandidateEvidence, func(i, j int) bool {
		return result.ShadowedCandidateEvidence[i].Canonical() < result.ShadowedCandidateEvidence[j].Canonical()
	})
	return nil
}

func copyCandidates(candidates []Candidate) []Candidate {
	copyOf := append([]Candidate(nil), candidates...)
	for index := range copyOf {
		copyOf[index].Options = append([]Identity(nil), copyOf[index].Options...)
	}
	return copyOf
}

func copyDetails(details []ImplementationDetail) []ImplementationDetail {
	return append([]ImplementationDetail(nil), details...)
}

func validateEvidenceConfig(input SemanticAdapterInput) error {
	if _, err := semantic.ParseIdentity(input.Producer.String()); err != nil {
		return adapterError(AdapterEvidenceConfig, "", input.Producer.String(), "producer is not a semantic identity")
	}
	if !input.EvidenceKind.Valid() {
		return adapterError(AdapterEvidenceConfig, "", "", "evidence kind is invalid")
	}
	digest := strings.TrimSpace(input.SourceDigest)
	if len(digest) != 64 {
		return adapterError(AdapterEvidenceConfig, "", "", "source digest must be SHA-256")
	}
	if _, err := hex.DecodeString(digest); err != nil {
		return adapterError(AdapterEvidenceConfig, "", "", "source digest is not hexadecimal")
	}
	return nil
}
