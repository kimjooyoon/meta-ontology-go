package adapter

import (
	"fmt"
	"strings"
)

// Expectation is owned by the runner oracle, not by an adapter.
type Expectation struct {
	Status      Status `json:"status"`
	FailureCode string `json:"failure_code,omitempty"`
}

// Response is the canonical adapter-to-runner result.
type Response struct {
	Schema            string           `json:"schema"`
	Fixture           string           `json:"fixture"`
	Operation         Operation        `json:"operation"`
	RunID             string           `json:"run_id"`
	Status            Status           `json:"status"`
	Failure           *Failure         `json:"failure,omitempty"`
	PromotionEligible bool             `json:"promotion_eligible"`
	Observed          Observed         `json:"observed"`
	Measurements      Measurements     `json:"measurements"`
	Evidence          EvidenceArtifact `json:"evidence"`
	ProducerClaims    ProducerClaims   `json:"producer_claims,omitempty"`
}

// ProducerClaims are advisory and are never accepted as observer proof.
type ProducerClaims struct {
	NoWrite *bool `json:"no_write,omitempty"`
}

// Failure identifies a deterministic safety or conformance rejection.
type Failure struct {
	Code       string `json:"code"`
	SemanticID string `json:"semantic_id,omitempty"`
	Detail     string `json:"detail,omitempty"`
}

func (r Request) Validate() error {
	if r.Schema != ProtocolSchema {
		return fmt.Errorf("unsupported request schema %q", r.Schema)
	}
	if strings.TrimSpace(r.RunID) == "" {
		return fmt.Errorf("run_id is required")
	}
	if strings.TrimSpace(r.Fixture) == "" {
		return fmt.Errorf("fixture is required")
	}
	if !knownOperation(r.Operation) {
		return fmt.Errorf("unsupported operation %q", r.Operation)
	}
	if err := r.Contract.validate(); err != nil {
		return err
	}
	if err := r.Expected.validate(); err != nil {
		return err
	}
	if err := validateSourceURI(r.Input.SourceURI); err != nil {
		return err
	}
	return validateAuthoritativeInput(r.Operation, r.Input)
}
