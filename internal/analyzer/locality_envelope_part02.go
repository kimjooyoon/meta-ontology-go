package analyzer

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// Validate checks schema, digest, IDs, closure shape, and duplicate entries.
func (e LocalityEnvelope) Validate() error {
	if e.SchemaVersion != localityEnvelopeSchema || !validDigest(e.BaseDigest) {
		return fmt.Errorf("locality envelope binding is incomplete")
	}
	if e.Digest == "" || e.Digest != e.StableHash() {
		return fmt.Errorf("locality envelope digest is invalid")
	}
	if err := validateLocalityIDs(e.Touched); err != nil {
		return err
	}
	if err := validateLocalityIDs(e.Affected); err != nil {
		return err
	}
	if err := validateLocalityFacts(e.PreservedFacts); err != nil {
		return err
	}
	affected := make(map[semantic.ID]struct{}, len(e.Affected))
	for _, id := range e.Affected {
		affected[id] = struct{}{}
	}
	for _, id := range e.Touched {
		if _, ok := affected[id]; !ok {
			return fmt.Errorf("touched ID is outside affected closure: %s", id)
		}
	}
	return nil
}

// LocalityEnvelopeFor computes one-hop affected closure from authoritative
// deterministic facts only. Candidates and deferred observations are not
// locality mutations and therefore cannot enter Touched or Affected.
func LocalityEnvelopeFor(base semantic.IR, result SemanticAdapterResult) (LocalityEnvelope, error) {
	normalizedBase, err := base.Normalized()
	if err != nil {
		return LocalityEnvelope{}, err
	}
	if err := result.IR.Validate(); err != nil {
		return LocalityEnvelope{}, err
	}
	baseFacts := normalizedBase.Graph.DeterministicFacts()
	for _, fact := range baseFacts {
		if !result.IR.Graph.HasFact(fact.Key()) {
			return LocalityEnvelope{}, fmt.Errorf("partial observation removed base fact %v", fact.Key())
		}
	}
	touched := localityTouched(normalizedBase, result.IR)
	envelope := LocalityEnvelope{
		SchemaVersion: localityEnvelopeSchema, BaseDigest: normalizedBase.StableHash(),
		Touched: touched, Affected: localityAffected(localityFactKeys(baseFacts), touched),
		PreservedFacts: localityFactKeys(baseFacts),
	}
	envelope.Digest = envelope.StableHash()
	return envelope, nil
}
