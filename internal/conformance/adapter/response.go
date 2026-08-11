package adapter

import (
	"fmt"
	"strings"
)

// Normalized validates a response and sorts all protocol sets by stable keys.
func (r Response) Normalized() (Response, error) {
	if r.Schema != ProtocolSchema {
		return Response{}, fmt.Errorf("unsupported response schema %q", r.Schema)
	}
	if strings.TrimSpace(r.Fixture) == "" {
		return Response{}, fmt.Errorf("response fixture is required")
	}
	if !knownOperation(r.Operation) {
		return Response{}, fmt.Errorf("unsupported response operation %q", r.Operation)
	}
	if !validStatus(r.Status) {
		return Response{}, fmt.Errorf("unsupported response status %q", r.Status)
	}
	if err := validateFailure(r.Status, r.Failure); err != nil {
		return Response{}, err
	}
	if r.PromotionEligible && r.Status != StatusPass {
		return Response{}, fmt.Errorf("only pass responses can be promotion eligible")
	}
	observed, err := r.Observed.normalized()
	if err != nil {
		return Response{}, fmt.Errorf("observed output: %w", err)
	}
	if err := r.Measurements.validate(); err != nil {
		return Response{}, err
	}
	evidence, err := r.Evidence.normalized()
	if err != nil {
		return Response{}, fmt.Errorf("evidence: %w", err)
	}
	r.Observed, r.Evidence = observed, evidence
	return r, nil
}

func validateFailure(status Status, failure *Failure) error {
	if status == StatusFail && failure == nil {
		return fmt.Errorf("fail response requires failure details")
	}
	if failure == nil {
		return nil
	}
	if strings.TrimSpace(failure.Code) == "" {
		return fmt.Errorf("failure code is required")
	}
	if failure.NoWrite && status == StatusPass {
		return fmt.Errorf("pass response cannot carry no-write failure")
	}
	return nil
}
