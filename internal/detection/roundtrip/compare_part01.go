package roundtrip

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// Verify checks both projection directions, the full round-trip, and any
// supplied generated-source locality witness.
func Verify(observation Observation) Report {
	var report Report
	report.merge(CheckDSLToIR(observation.DSL, observation.IR))
	report.merge(CheckGoToIR(observation.IR, observation.GoIR))
	report.merge(CheckRoundTrip(observation.DSL, observation.GoIR))
	if observation.BeforeGo != nil || observation.AfterGo != nil {
		allowed := append([]semantic.ID(nil), observation.AllowedIDs...)
		if len(allowed) == 0 && hasIR(observation.BeforeIR, observation.AfterIR) {
			allowed = changedLocality(observation.BeforeIR, observation.AfterIR)
		}
		report.merge(CheckLocality(LocalityInput{
			Before:     observation.BeforeGo,
			After:      observation.AfterGo,
			AllowedIDs: allowed,
		}))
	}
	report.normalize()
	return report
}

// CheckDSLToIR checks the authoritative DSL semantic view against its lowered
// IR projection.
func CheckDSLToIR(dsl, lowered semantic.IR) Report {
	return comparePair(RuleDSLToIR, "dsl", dsl, "ir", lowered)
}

// CheckGoToIR checks the canonical IR against facts lifted from Go.
func CheckGoToIR(ir, lifted semantic.IR) Report {
	return comparePair(RuleGoToIR, "ir", ir, "go-ir", lifted)
}

// CheckRoundTrip checks DSL → IR → Go → IR semantic stability.
func CheckRoundTrip(dsl, lifted semantic.IR) Report {
	return comparePair(RuleRoundTrip, "dsl", dsl, "go-ir", lifted)
}

// SemanticDelta computes presentation-insensitive changes between snapshots.
func SemanticDelta(before, after semantic.IR) (Delta, error) {
	left, err := before.Normalized()
	if err != nil {
		return Delta{}, fmt.Errorf("before IR: %w", err)
	}
	right, err := after.Normalized()
	if err != nil {
		return Delta{}, fmt.Errorf("after IR: %w", err)
	}
	delta := Delta{
		MetadataChanged: left.Version != right.Version || left.Package != right.Package || left.Namespace != right.Namespace,
	}
	diffNodes(&delta, left.Graph.Nodes(), right.Graph.Nodes())
	diffFacts(&delta, left.Graph.DeterministicFacts(), right.Graph.DeterministicFacts())
	delta.TouchedIDs = touchedIDs(delta)
	delta.AffectedIDs = affectedIDs(delta, left, right)
	return delta, nil
}
