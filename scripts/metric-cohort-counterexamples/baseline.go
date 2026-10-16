package main

import (
	"fmt"
	"path/filepath"

	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

func verifyBaseline(cfg config, f fixture, result *report) error {
	result.DeclaredCohort, result.ExpectedCount = f.Contract, len(f.Contract)
	digest, err := artifact.Digest(f.Contract)
	if err != nil {
		return err
	}
	result.DeclaredCohortDigest = digest
	if len(f.Contract) == 0 || len(f.Receipt.Expected) != len(f.Contract) {
		return fmt.Errorf("original producer expected set differs from the source-declared cohort")
	}
	for index, spec := range f.Contract {
		row := f.Receipt.Expected[index]
		if row.Operation != spec.Subject || row.CarrierOperation != spec.Carrier || row.ProofChoice != spec.Family {
			return fmt.Errorf("original producer changed a source-declared expected identity")
		}
	}
	baseline, err := readiness.BuildWithConceptOperationEvidence(f.Concept, f.Raw, f.Inputs,
		cfg.Root, filepath.Join(cfg.Scratch, "baseline"), cfg.Origin.Repository, cfg.Origin.SubjectSHA)
	if err != nil {
		return err
	}
	result.ReadinessTotal = baseline.Snapshot.Summary.Total
	return verifyPositive(baseline)
}

func verifyPositive(baseline readiness.Receipt) error {
	if baseline.Snapshot.Summary.Total != 24 || len(baseline.Snapshot.Obligations) != 24 ||
		baseline.Snapshot.RepositoryWrites != 0 || baseline.FixedPoint.Decision != "NO_CHANGE" ||
		baseline.ProposalPromotionDigest != "" || baseline.GuardedCapabilityDigest != "" {
		return fmt.Errorf("original readiness scope or authority boundary changed")
	}
	found := 0
	for _, obligation := range baseline.Snapshot.Obligations {
		if obligation.ID == "META-CONCEPT-GOVERNED-REFACTORING" && obligation.Status == "SATISFIED" {
			found++
		}
	}
	if found != 1 {
		return fmt.Errorf("original produced receipt did not satisfy its one readiness obligation")
	}
	return nil
}
