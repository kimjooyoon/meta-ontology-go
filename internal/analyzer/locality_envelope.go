package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const localityEnvelopeSchema = "analyzer-locality/v1"

// LocalityEnvelope mirrors bidir.Locality's Touched/Affected vocabulary. The
// preserved fact set is analyzer-owned compatibility data for the future
// bidirectional adapter; it prevents partial observations from deleting base.
type LocalityEnvelope struct {
	SchemaVersion  string             `json:"schema_version"`
	BaseDigest     string             `json:"base_digest"`
	Touched        []semantic.ID      `json:"touched"`
	Affected       []semantic.ID      `json:"affected"`
	PreservedFacts []semantic.FactKey `json:"preserved_facts"`
	Digest         string             `json:"digest"`
}

// Canonical is order-independent and excludes Digest so it can be hashed.
func (e LocalityEnvelope) Canonical() string {
	var builder strings.Builder
	builder.WriteString(localityEnvelopeSchema)
	builder.WriteByte('\n')
	writeBindingField(&builder, e.SchemaVersion)
	writeBindingField(&builder, e.BaseDigest)
	for _, id := range sortedLocalityIDs(e.Touched) {
		writeBindingField(&builder, "touched")
		writeBindingField(&builder, id.String())
	}
	for _, id := range sortedLocalityIDs(e.Affected) {
		writeBindingField(&builder, "affected")
		writeBindingField(&builder, id.String())
	}
	for _, key := range sortedLocalityFacts(e.PreservedFacts) {
		writeBindingField(&builder, "preserved")
		writeBindingField(&builder, key.Subject.String())
		writeBindingField(&builder, key.Predicate.String())
		writeBindingField(&builder, key.Object.String())
	}
	return builder.String()
}

// StableHash is the deterministic digest of the locality closure envelope.
func (e LocalityEnvelope) StableHash() string {
	return semantic.StableHashString(e.Canonical())
}

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

// ValidateLocalityEnvelope is a read-only, transactional handoff check. It
// rejects missing, relabeled, or tampered locality without changing either IR.
func ValidateLocalityEnvelope(base semantic.IR, result SemanticAdapterResult) error {
	expected, err := LocalityEnvelopeFor(base, result)
	if err != nil {
		return adapterError(AdapterLocalityConfig, "", "", err.Error())
	}
	if result.Locality.Digest != expected.Digest || result.Locality.Canonical() != expected.Canonical() {
		return adapterError(AdapterLocalityConfig, "", "", "locality envelope does not match base closure")
	}
	return nil
}

func validLocalityEnvelope(result SemanticAdapterResult) bool {
	if err := result.Locality.Validate(); err != nil {
		return false
	}
	if result.Locality.BaseDigest != localityBaseDigest(result) {
		return false
	}
	for _, key := range result.Locality.PreservedFacts {
		if !result.IR.Graph.HasFact(key) {
			return false
		}
	}
	touched := localityTouchedFromPreserved(result.Locality.PreservedFacts, result.IR)
	return equalLocalityIDs(touched, result.Locality.Touched) &&
		equalLocalityIDs(localityAffected(result.Locality.PreservedFacts, touched), result.Locality.Affected)
}

func localityTouched(base, result semantic.IR) []semantic.ID {
	touched := make(map[semantic.ID]struct{})
	for _, fact := range result.Graph.DeterministicFacts() {
		if base.Graph.HasFact(fact.Key()) {
			continue
		}
		touched[fact.Subject] = struct{}{}
		touched[fact.Object] = struct{}{}
	}
	return sortedLocalityIDsFromSet(touched)
}

func localityTouchedFromPreserved(preserved []semantic.FactKey, result semantic.IR) []semantic.ID {
	base := make(map[semantic.FactKey]struct{}, len(preserved))
	for _, key := range preserved {
		base[key] = struct{}{}
	}
	touched := make(map[semantic.ID]struct{})
	for _, fact := range result.Graph.DeterministicFacts() {
		if _, ok := base[fact.Key()]; ok {
			continue
		}
		touched[fact.Subject] = struct{}{}
		touched[fact.Object] = struct{}{}
	}
	return sortedLocalityIDsFromSet(touched)
}

func localityAffected(baseFacts []semantic.FactKey, touched []semantic.ID) []semantic.ID {
	closure := make(map[semantic.ID]struct{}, len(touched))
	for _, id := range touched {
		closure[id] = struct{}{}
	}
	for _, fact := range baseFacts {
		_, subjectTouched := closure[fact.Subject]
		_, objectTouched := closure[fact.Object]
		if subjectTouched || objectTouched {
			closure[fact.Subject] = struct{}{}
			closure[fact.Object] = struct{}{}
		}
	}
	return sortedLocalityIDsFromSet(closure)
}

func localityFactKeys(facts []semantic.Fact) []semantic.FactKey {
	keys := make([]semantic.FactKey, 0, len(facts))
	for _, fact := range facts {
		keys = append(keys, fact.Key())
	}
	return sortedLocalityFacts(keys)
}

func localityBaseDigest(result SemanticAdapterResult) string {
	for _, fact := range result.NormalizedDelta.SignatureFacts {
		return fact.Binding.BaseDigest
	}
	for _, candidate := range result.NormalizedDelta.CandidateFacts {
		return candidate.Binding.BaseDigest
	}
	for _, fact := range result.NormalizedDelta.DeferredFacts {
		return fact.Binding.BaseDigest
	}
	for _, observation := range result.NormalizedDelta.DeferredImplementation {
		return observation.BaseDigest
	}
	for _, detail := range result.NormalizedDelta.DeferredDetails {
		return detail.Binding.BaseDigest
	}
	for _, slot := range result.NormalizedDelta.DeferredSlots {
		return slot.BaseDigest
	}
	return ""
}

func validateLocalityIDs(ids []semantic.ID) error {
	seen := make(map[semantic.ID]struct{}, len(ids))
	for _, id := range ids {
		if _, err := semantic.ParseIdentity(id.String()); err != nil {
			return err
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("duplicate locality ID: %s", id)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func validateLocalityFacts(facts []semantic.FactKey) error {
	seen := make(map[semantic.FactKey]struct{}, len(facts))
	for _, key := range facts {
		if _, err := semantic.ParseIdentity(key.Subject.String()); err != nil {
			return err
		}
		if _, err := semantic.ParseIdentity(key.Object.String()); err != nil {
			return err
		}
		if !key.Predicate.Valid() {
			return fmt.Errorf("invalid preserved fact predicate: %s", key.Predicate)
		}
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate preserved fact: %v", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func sortedLocalityIDs(ids []semantic.ID) []semantic.ID {
	copyOf := append([]semantic.ID(nil), ids...)
	sort.Slice(copyOf, func(i, j int) bool { return copyOf[i] < copyOf[j] })
	return copyOf
}

func sortedLocalityIDsFromSet(ids map[semantic.ID]struct{}) []semantic.ID {
	output := make([]semantic.ID, 0, len(ids))
	for id := range ids {
		output = append(output, id)
	}
	return sortedLocalityIDs(output)
}

func sortedLocalityFacts(facts []semantic.FactKey) []semantic.FactKey {
	copyOf := append([]semantic.FactKey(nil), facts...)
	sort.Slice(copyOf, func(i, j int) bool {
		if copyOf[i].Subject != copyOf[j].Subject {
			return copyOf[i].Subject < copyOf[j].Subject
		}
		if copyOf[i].Predicate != copyOf[j].Predicate {
			return copyOf[i].Predicate < copyOf[j].Predicate
		}
		return copyOf[i].Object < copyOf[j].Object
	})
	return copyOf
}

func equalLocalityIDs(left, right []semantic.ID) bool {
	left, right = sortedLocalityIDs(left), sortedLocalityIDs(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
