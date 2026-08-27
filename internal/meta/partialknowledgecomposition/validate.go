package partialknowledgecomposition

import (
	"fmt"
	"regexp"
	"strings"
)

var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

var fixedCaseIDs = []string{
	"exact-pair", "direct-unknown", "dependency-blocked",
	"invariant-preservation", "mixed-unknown-and-blocked",
}

func ValidateInput(input Input) error {
	if input.Repository == "" || !headPattern.MatchString(input.HeadSHA) {
		return fmt.Errorf("partial-knowledge identity is malformed")
	}
	if input.SourcePath != SourcePath || input.Fixture.SourcePath != SourcePath ||
		input.Fixture.Schema != FixtureSchema || input.Fixture.FixedDenominator != FixedDenominator ||
		len(input.Fixture.Cases) != FixedDenominator {
		return fmt.Errorf("partial-knowledge fixture identity or denominator is malformed")
	}
	if !strings.Contains(string(input.Source), "package partialknowledgecomposition") ||
		!strings.Contains(string(input.Source), "entity DirectUnknown") ||
		!strings.Contains(string(input.Source), "entity DependencyBlocked") {
		return fmt.Errorf("partial-knowledge Gooo source is not the declared value vocabulary")
	}
	for index, current := range input.Fixture.Cases {
		if current.ID != fixedCaseIDs[index] || current.Producer != Producer || current.Consumer != Consumer ||
			current.MetaOperation == "" || !validProofChoice(current.ProofChoice) {
			return fmt.Errorf("case %d metadata is not bound to the producer contract", index+1)
		}
		if err := validateOperand(current.Left); err != nil {
			return fmt.Errorf("case %q left operand: %w", current.ID, err)
		}
		if err := validateOperand(current.Right); err != nil {
			return fmt.Errorf("case %q right operand: %w", current.ID, err)
		}
		if !validResultState(current.ExpectedState) || current.ExpectedDecision == "" || current.ExpectedReason == "" {
			return fmt.Errorf("case %q expected outcome is malformed", current.ID)
		}
		value := Compose(current.Left, current.Right)
		decision, reason, _ := classify(value)
		if current.ExpectedState != value.State || current.ExpectedDecision != decision || current.ExpectedReason != reason {
			return fmt.Errorf("case %q expected outcome does not match composition calculus", current.ID)
		}
	}
	return nil
}

func validateOperand(operand Operand) error {
	if operand.Operation == "" || !validState(operand.State) {
		return fmt.Errorf("operation or state is malformed")
	}
	switch operand.State {
	case StateExact:
		if operand.BlockedDependency != "" || len(operand.Invariants) != 0 {
			return fmt.Errorf("exact operand carries unresolved evidence")
		}
	case StateDependencyBlocked:
		if operand.BlockedDependency == "" || len(operand.Invariants) != 0 {
			return fmt.Errorf("dependency-blocked operand lacks one dependency")
		}
	case StateInvariantOnly:
		if operand.BlockedDependency != "" || len(operand.Invariants) == 0 {
			return fmt.Errorf("invariant-only operand lacks invariant evidence")
		}
	case StateDirectUnknown:
		if operand.BlockedDependency != "" || len(operand.Invariants) != 0 {
			return fmt.Errorf("direct-unknown operand carries another cause")
		}
	}
	return nil
}
