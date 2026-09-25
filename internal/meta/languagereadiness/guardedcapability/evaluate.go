package guardedcapability

import (
	"encoding/hex"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedpromotion"
)

func Build(source Source) Receipt {
	coordinates := []guardedpromotion.Coordinate{
		coordinate("current-subject-sha", "FOUNDATION", validSHA(source.CurrentHeadSHA), validSHA(source.CurrentHeadSHA)),
		coordinate("foundation-provenance", "FOUNDATION", source.ArtifactID != 0, foundationExact(source)),
		coordinate("foundation-authorized-report", "FOUNDATION", source.FoundationReport.ReportDigest != "", foundationExact(source)),
		coordinate("foundation-ancestor", "COHERENCE", source.AncestryObserved, source.FoundationAncestor),
		coordinate("guard-implementation-tree", "COHERENCE", source.GuardTreesObserved, source.FoundationGuardTree != "" && source.FoundationGuardTree == source.CurrentGuardTree),
		coordinate("witness-implementation-tree", "COHERENCE", source.WitnessTreesObserved, source.FoundationWitnessTree != "" && source.FoundationWitnessTree == source.CurrentWitnessTree),
		coordinate("observer-write-boundary", "REGRESSION", true, source.RepositoryWrites == 0),
		coordinate("mutation-authority-boundary", "REGRESSION", true, !source.MutationAuthorized),
	}
	summary := summarize(source, coordinates)
	decision, reason, resolution := decide(summary)
	receipt := Receipt{Schema: Schema, Decision: decision, Reason: reason,
		Resolution: resolution, Source: source, Summary: summary, Coordinates: coordinates,
		Indicators: indicators(source, summary, coordinates),
		Proofs:     proofs(coordinates)}
	seal(&receipt)
	return receipt
}

func coordinate(id, proof string, observed, satisfied bool) guardedpromotion.Coordinate {
	item := guardedpromotion.Coordinate{ID: id, ProofChoice: proof}
	switch {
	case !observed:
		item.Status, item.Reason = "UNRESOLVED", "COORDINATE_EVIDENCE_UNKNOWN"
	case satisfied:
		item.Status, item.Reason = "SATISFIED", "COORDINATE_EXACTLY_PROVEN"
	default:
		item.Status, item.Reason = "NOT_SATISFIED", "COORDINATE_EXACTLY_REJECTED"
	}
	return item
}

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
