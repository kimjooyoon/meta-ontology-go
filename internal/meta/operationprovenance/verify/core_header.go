package verify

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func verifyHeader(actual receipt, source []byte, ir semantic.IR) error {
	if actual.Schema != receiptSchema || actual.Toolchain != toolchain {
		return fmt.Errorf("receipt schema or toolchain is invalid")
	}
	if actual.Source != digest(source) || actual.Semantic != "sha256:"+ir.StableHash() {
		return fmt.Errorf("raw or canonical semantic source digest is not bound")
	}
	if actual.SourceResolution != "EXACT" && actual.SourceResolution != "LOWER_RESOLUTION" {
		return fmt.Errorf("source resolution is invalid")
	}
	if actual.Observation.Before != actual.Observation.After || len(actual.Observation.Changed) != 0 || actual.Observation.Writes || actual.Observation.Authority == "" {
		return fmt.Errorf("repository write observation is not clean or authority is unresolved")
	}
	return nil
}
