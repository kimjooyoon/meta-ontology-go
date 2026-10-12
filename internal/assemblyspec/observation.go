package assemblyspec

import (
	"fmt"
	"go/token"
)

// Observation declares how to distinguish source-owned candidates that satisfy
// the initial examples. Inputs are questions; only a separate oracle supplies answers.
type Observation struct {
	Inputs                 []int64 `json:"inputs"`
	MaxCandidates          int     `json:"max_candidates"`
	MaxRounds              int     `json:"max_rounds"`
	OracleActivity         string  `json:"oracle_activity,omitempty"`
	ReuseProbeOutputs      bool    `json:"reuse_probe_outputs,omitempty"`
	ResolveUniqueCandidate bool    `json:"resolve_unique_candidate,omitempty"`
}

func (s Spec) validateObservation() error {
	o := s.Observation
	if o == nil {
		return nil
	}
	if s.Search != nil || s.FillPlan != nil || len(s.ValueCases)+len(s.ValueHoldoutCases) != 0 || len(s.Cases) == 0 {
		return fmt.Errorf("observe requires Integer typed choices and initial case declarations")
	}
	if len(o.Inputs) == 0 || len(o.Inputs) > 32 || o.MaxCandidates < 1 || o.MaxCandidates > 64 ||
		o.MaxRounds < 1 || o.MaxRounds > 8 {
		return fmt.Errorf("observe requires 1..32 inputs, max_candidates 1..64 and max_rounds 1..8")
	}
	if o.OracleActivity != "" && (len(o.OracleActivity) > 128 || !token.IsIdentifier(o.OracleActivity)) {
		return fmt.Errorf("observe oracle must name a Gooo activity")
	}
	seen := make(map[int64]bool, len(o.Inputs))
	for _, input := range o.Inputs {
		if seen[input] {
			return fmt.Errorf("duplicate observe input %d", input)
		}
		seen[input] = true
		for _, held := range s.HoldoutCases {
			if input == held.Input {
				return fmt.Errorf("observe input %d is reserved by holdout_case", input)
			}
		}
	}
	return nil
}
