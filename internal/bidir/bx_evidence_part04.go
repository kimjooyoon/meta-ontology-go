package bidir

import (
	"fmt"
	"strings"
)

func removedCreated(base, after Model, delta FactDelta) bool {
	allowed := make(map[string]struct{}, len(delta.Removed))
	for _, fact := range delta.Removed {
		allowed[fact.SemanticKey()] = struct{}{}
	}
	for _, relation := range base.Relations {
		if _, exists := findRelation(after, relation.Kind, relation.Source, relation.Target); exists {
			continue
		}
		if _, explicit := allowed[relationKey(relation.Kind, relation.Source, relation.Target)]; !explicit {
			return true
		}
	}
	return false
}
func candidatePromoted(base Model, delta FactDelta, model Model) bool {
	for _, fact := range delta.Added.ByLayer(CandidateFact) {
		if _, exists := findRelation(base, fact.Predicate, fact.Subject, fact.Object); exists {
			continue
		}
		if _, exists := findRelation(model, fact.Predicate, fact.Subject, fact.Object); exists {
			return true
		}
	}
	return false
}
func deferredBXSeams() []string {
	return []string{
		"generic gooo:invokes lifting",
		"PROV-O adapter mapping policy",
		"Go-lift/CLI delta atomicity",
		"rejected transaction filesystem/inode observer",
		"three-way merge",
	}
}

// Canonical returns line-oriented golden evidence with deterministic IDs.
func (e BXEvidence) Canonical() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "schema_version=%s\n", e.SchemaVersion)
	fmt.Fprintf(&builder, "fixture=%s\n", e.Fixture)
	fmt.Fprintf(&builder, "base_dsl=%s/%d\n", e.Base.DSL.Hash, e.Base.DSL.Count)
	fmt.Fprintf(&builder, "base_ir=%s/%d\n", e.Base.IR.Hash, e.Base.IR.Count)
	fmt.Fprintf(&builder, "base_go=%s/%d\n", e.Base.Go.Hash, e.Base.Go.Count)
	fmt.Fprintf(&builder, "base_source_map=%s/%d\n", e.Base.SourceMap.Hash, e.Base.SourceMap.Count)
	fmt.Fprintf(&builder, "base_evidence=%s/%d\n", e.Base.Evidence.Hash, e.Base.Evidence.Count)
	fmt.Fprintf(&builder, "base_provenance=%s/%d\n", e.Base.Provenance.Hash, e.Base.Provenance.Count)
	fmt.Fprintf(&builder, "get_put=%s\n", evidenceStatus(e.GetPutPassed))
	fmt.Fprintf(&builder, "put_get=%s\n", evidenceStatus(e.PutGetPassed))
	fmt.Fprintf(&builder, "semantic_equivalence=%s\n", evidenceStatus(e.SemanticEquivalent))
	fmt.Fprintf(&builder, "accepted_relation_adds=%d\n", e.AcceptedRelationAdds)
	writeDeltaCanonical(&builder, "accepted", e.Delta)
	writeDeltaCanonical(&builder, "partial", e.PartialDelta)
	fmt.Fprintf(&builder, "touched=%s\n", joinIDs(e.Locality.Touched))
	fmt.Fprintf(&builder, "affected=%s\n", joinIDs(e.Locality.Affected))
	writeTransactionCanonical(&builder, "accepted", e.AcceptedTransaction)
	writeTransactionCanonical(&builder, "rejected", e.RejectedTransaction)
	fmt.Fprintf(&builder, "partial_conflict=%s\n", conflictStatus(e.PartialConflict.Kind))
	fmt.Fprintf(&builder, "partial_conflict_count=%d\n", e.PartialConflict.Count)
	fmt.Fprintf(&builder, "partial_transactional=%s\n", evidenceStatus(e.PartialConflict.Transactional))
	fmt.Fprintf(&builder, "partial_no_write=%s\n", evidenceStatus(e.PartialConflict.NoWriteObserved))
	fmt.Fprintf(&builder, "partial_removed_created=%t\n", e.PartialConflict.RemovedCreated)
	fmt.Fprintf(&builder, "partial_candidate_promoted=%t\n", e.PartialConflict.CandidatePromoted)
	fmt.Fprintf(&builder, "deferred=%s\n", strings.Join(e.Deferred, ","))
	return builder.String()
}
