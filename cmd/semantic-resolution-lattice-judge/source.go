package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

const sourceCasePrefix = "resolution-lattice.case;"

type declaredCase struct {
	ID              string
	Observation     observation
	ClaimID         string
	ClaimPriorState string
}

func parseGoooCases(source string) ([]declaredCase, error) {
	requiredActivities := map[string]bool{
		"ObservePartialObservation": false,
		"DescendToInvariantOnly":    false,
		"PreserveClaimState":        false,
		"EmitResolutionReceipt":     false,
		"AdjudicateReceipt":         false,
	}
	result := make([]declaredCase, 0, 4)
	seen := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(source))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		activity, ok, err := parseActivityLine(line)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		if _, expected := requiredActivities[activity.name]; expected {
			requiredActivities[activity.name] = true
		}
		if !strings.HasPrefix(activity.program, sourceCasePrefix) {
			continue
		}
		item, err := parseDeclaredCase(activity.program)
		if err != nil {
			return nil, fmt.Errorf("activity %s: %w", activity.name, err)
		}
		if seen[item.ID] {
			return nil, fmt.Errorf("duplicate lattice case %q", item.ID)
		}
		seen[item.ID] = true
		result = append(result, item)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan Gooo source: %w", err)
	}
	for name, present := range requiredActivities {
		if !present {
			return nil, fmt.Errorf("required Gooo activity %q is missing", name)
		}
	}
	if len(result) != 4 {
		return nil, fmt.Errorf("Gooo lattice case denominator = %d, want 4", len(result))
	}
	return result, nil
}

type parsedActivity struct {
	name    string
	program string
}

func parseActivityLine(line string) (parsedActivity, bool, error) {
	if !strings.HasPrefix(line, "activity ") {
		return parsedActivity{}, false, nil
	}
	rest := strings.TrimSpace(strings.TrimPrefix(line, "activity "))
	open := strings.IndexByte(rest, '(')
	if open <= 0 {
		return parsedActivity{}, false, fmt.Errorf("activity name or parameter list is malformed")
	}
	name := strings.TrimSpace(rest[:open])
	closeOffset := strings.IndexByte(rest[open+1:], ')')
	if closeOffset < 0 {
		return parsedActivity{}, false, fmt.Errorf("activity %s has no closing parameter list", name)
	}
	after := strings.TrimSpace(rest[open+1+closeOffset+1:])
	if !strings.HasPrefix(after, "-> ") {
		return parsedActivity{}, false, fmt.Errorf("activity %s has no result arrow", name)
	}
	after = strings.TrimSpace(strings.TrimPrefix(after, "->"))
	resultEnd := strings.IndexAny(after, " \t")
	if resultEnd < 0 {
		return parsedActivity{name: name}, true, nil
	}
	if resultEnd == 0 {
		return parsedActivity{}, false, fmt.Errorf("activity %s has no result type", name)
	}
	after = strings.TrimSpace(after[resultEnd:])
	if !strings.HasPrefix(after, "computes ") {
		return parsedActivity{name: name}, true, nil
	}
	rawProgram := strings.TrimSpace(strings.TrimPrefix(after, "computes "))
	program, err := strconv.Unquote(rawProgram)
	if err != nil {
		return parsedActivity{}, false, fmt.Errorf("activity %s value program: %w", name, err)
	}
	return parsedActivity{name: name, program: program}, true, nil
}

func parseDeclaredCase(program string) (declaredCase, error) {
	fields := map[string]string{}
	for _, field := range strings.Split(strings.TrimPrefix(program, sourceCasePrefix), ";") {
		key, value, ok := strings.Cut(field, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || key == "" || value == "" {
			return declaredCase{}, fmt.Errorf("malformed case field %q", field)
		}
		if _, exists := fields[key]; exists {
			return declaredCase{}, fmt.Errorf("duplicate case field %q", key)
		}
		fields[key] = value
	}
	keys := []string{"id", "required", "observed", "reason", "repository_writes", "mutation_authority", "claim_id", "claim_prior_state"}
	for _, key := range keys {
		if fields[key] == "" {
			return declaredCase{}, fmt.Errorf("missing case field %q", key)
		}
	}
	if len(fields) != len(keys) {
		for key := range fields {
			found := false
			for _, expected := range keys {
				if key == expected {
					found = true
					break
				}
			}
			if !found {
				return declaredCase{}, fmt.Errorf("unexpected case field %q", key)
			}
		}
		return declaredCase{}, fmt.Errorf("case field set is invalid")
	}
	required, err := strconv.Atoi(fields["required"])
	if err != nil {
		return declaredCase{}, fmt.Errorf("required: %w", err)
	}
	observed, err := strconv.Atoi(fields["observed"])
	if err != nil {
		return declaredCase{}, fmt.Errorf("observed: %w", err)
	}
	writes, err := strconv.Atoi(fields["repository_writes"])
	if err != nil {
		return declaredCase{}, fmt.Errorf("repository_writes: %w", err)
	}
	authority, err := strconv.ParseBool(fields["mutation_authority"])
	if err != nil {
		return declaredCase{}, fmt.Errorf("mutation_authority: %w", err)
	}
	state := fields["claim_prior_state"]
	if state != "OPEN" && state != "DISCHARGED" && state != "REFUTED" {
		return declaredCase{}, fmt.Errorf("invalid claim_prior_state %q", state)
	}
	return declaredCase{
		ID: fields["id"],
		Observation: observation{
			Required: required, Observed: observed, Reason: fields["reason"],
			RepositoryWrites: writes, MutationAuthority: authority,
		},
		ClaimID: fields["claim_id"], ClaimPriorState: state,
	}, nil
}
