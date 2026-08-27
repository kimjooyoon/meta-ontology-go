package causalci

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

func decodeInput(raw []byte) (Input, error) {
	var input Input
	if err := json.Unmarshal(raw, &input); err != nil {
		return Input{}, fmt.Errorf("decode input: %w", err)
	}
	if err := validateInput(input); err != nil {
		return Input{}, err
	}
	return input, nil
}

func validateInput(input Input) error {
	if input.Schema != InputSchema || input.SourcePath == "" || filepath.Ext(input.SourcePath) != ".gooo" {
		return fmt.Errorf("%s: %s", ReasonMalformedInput, "source authority")
	}
	if input.Operation.Producer == "" || input.Operation.Consumer == "" || input.Operation.MetaOperation == "" || input.Operation.ProofChoice == "" || input.Operation.MutationAuthority || !input.Operation.ReadOnly {
		return fmt.Errorf("%s: %s", ReasonMalformedInput, "operation boundary")
	}
	if input.Policy.Schema != PolicySchema || input.Policy.FullSuiteID != "full-suite" || len(input.Policy.Checks) != FixedCheckDenominator {
		return fmt.Errorf("%s: fixed check policy", ReasonMalformedInput)
	}
	for index, check := range input.Policy.Checks {
		if check.ID != requiredCheckIDs[index] || check.Ordinal != index+1 || check.Scope == "" || check.Description == "" {
			return fmt.Errorf("%s: check catalog ordinal %d", ReasonMalformedInput, index+1)
		}
	}
	if len(input.Cases) != ScenarioDenominator {
		return fmt.Errorf("%s: scenario denominator %d", ReasonMalformedInput, ScenarioDenominator)
	}
	for index, expected := range requiredScenarioIDs() {
		if input.Cases[index].ID != expected {
			return fmt.Errorf("%s: scenario %d", ReasonMalformedInput, index+1)
		}
		if err := validateCaseShape(input.Cases[index]); err != nil {
			return err
		}
	}
	if len(input.ClaimTransitions) == 0 {
		return fmt.Errorf("%s: empty claim transition ledger", ReasonMalformedInput)
	}
	for index, transition := range input.ClaimTransitions {
		if transition.Sequence != index+1 || transition.ClaimID == "" || transition.Before == "" || transition.After == "" || transition.Event == "" || transition.EvidenceDigest == "" || !coordinateKnown(transition.Coordinate) {
			return fmt.Errorf("%s: claim transition %d", ReasonMalformedInput, index+1)
		}
	}
	return nil
}

func validateCaseShape(value Case) error {
	if len(value.ChangedFiles) == 0 || len(value.Claims) == 0 || len(value.ImpactEdges) == 0 {
		return fmt.Errorf("%s: incomplete case %q", ReasonMalformedInput, value.ID)
	}
	claims := make(map[string]struct{}, len(value.Claims))
	for _, claim := range value.Claims {
		if !strings.HasPrefix(claim.ID, "claim:") || claim.Question == "" || claim.State == "" {
			return fmt.Errorf("%s: claim in %q", ReasonMalformedInput, value.ID)
		}
		if _, exists := claims[claim.ID]; exists {
			return fmt.Errorf("%s: duplicate claim %q", ReasonMalformedInput, claim.ID)
		}
		claims[claim.ID] = struct{}{}
	}
	seenEdges := map[string]struct{}{}
	for _, edge := range value.ImpactEdges {
		if edge.ID == "" || edge.From == "" || edge.To == "" || edge.Kind == "" || edge.Reason == "" || !coordinateKnown(edge.Coordinate) {
			return fmt.Errorf("%s: edge in %q", ReasonMalformedInput, value.ID)
		}
		if _, exists := seenEdges[edge.ID]; exists {
			return fmt.Errorf("%s: duplicate edge %q", ReasonMalformedInput, edge.ID)
		}
		seenEdges[edge.ID] = struct{}{}
	}
	return nil
}

func coordinateKnown(value Coordinate) bool {
	return value.Stage != "" && value.Step != "" && value.Reason != ""
}
