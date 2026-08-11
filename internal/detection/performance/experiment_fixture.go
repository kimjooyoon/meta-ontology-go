package performance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Outcome is the falsifiable result of evaluating one host measurement.
type Outcome string

const (
	OutcomePass     Outcome = "pass"
	OutcomeFail     Outcome = "fail"
	OutcomeDeferred Outcome = "deferred"
)

// Fixture is the smallest reproducible input shared by host implementations.
type Fixture struct {
	ID          string `json:"id"`
	Stage       string `json:"stage"`
	Input       string `json:"input"`
	InputDigest string `json:"input_digest"`
}

// NewFixture creates a fixture with a deterministic input digest.
func NewFixture(id, stage, input string) Fixture {
	return Fixture{ID: id, Stage: stage, Input: input, InputDigest: DigestInput(input)}
}

// DigestInput returns the lowercase SHA-256 digest used for fixture identity.
func DigestInput(input string) string {
	digest := sha256.Sum256([]byte(input))
	return hex.EncodeToString(digest[:])
}

// Validate checks that the recorded digest matches the fixture input.
func (f Fixture) Validate() error {
	if f.ID == "" || f.Stage == "" || f.Input == "" {
		return fmt.Errorf("fixture requires ID, stage, and input")
	}
	if f.InputDigest != DigestInput(f.Input) {
		return fmt.Errorf("fixture %q input digest does not match input", f.ID)
	}
	return nil
}

// Hypothesis binds a fixture, contract, and exact deterministic measurements.
type Hypothesis struct {
	ID                  string   `json:"id"`
	Statement           string   `json:"statement"`
	Contract            Contract `json:"contract"`
	Fixture             Fixture  `json:"fixture"`
	Repetitions         uint64   `json:"repetitions"`
	ExpectedOperations  uint64   `json:"expected_operations"`
	ExpectedAllocations uint64   `json:"expected_allocations"`
}

// Validate checks the reusable experiment contract before a host runs it.
func (h Hypothesis) Validate() error {
	if h.ID == "" || h.Statement == "" {
		return fmt.Errorf("hypothesis requires ID and statement")
	}
	if err := h.Contract.Validate(); err != nil {
		return err
	}
	if err := h.Fixture.Validate(); err != nil {
		return err
	}
	if h.Contract.Stage != h.Fixture.Stage {
		return fmt.Errorf("hypothesis contract stage %q does not match fixture stage %q", h.Contract.Stage, h.Fixture.Stage)
	}
	if h.Repetitions == 0 {
		return fmt.Errorf("hypothesis repetitions must be positive")
	}
	return nil
}
