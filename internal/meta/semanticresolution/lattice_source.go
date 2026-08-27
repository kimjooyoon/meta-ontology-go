package semanticresolution

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const latticeCaseProgramPrefix = "resolution-lattice.case;"

type sourceCase struct {
	ID                 string
	Observation        PartialObservation
	ClaimID            string
	ExpectedClaimState ClaimState
}

// CasesFromGoooSource exposes the source-derived case producer used by the
// lattice receipt. The machine-readable value programs in the Gooo source,
// rather than a Go fixture, are the producer input.
func CasesFromGoooSource(sourcePath, source string) ([]LatticeCase, error) {
	cases, _, _, err := casesFromGoooSource(sourcePath, source)
	return cases, err
}

// ClaimsFromGoooSource exposes the source-derived claim producer for metric
// provenance and downstream consumers.
func ClaimsFromGoooSource(sourcePath, source string) ([]ClaimRecord, error) {
	_, claims, _, err := casesFromGoooSource(sourcePath, source)
	return claims, err
}

func casesFromGoooSource(sourcePath, source string) ([]LatticeCase, []ClaimRecord, string, error) {
	sourceCases, err := parseGoooCases(sourcePath, source)
	if err != nil {
		return nil, nil, "", err
	}
	cases := make([]LatticeCase, 0, len(sourceCases))
	claims := make([]ClaimRecord, 0, len(sourceCases))
	for _, item := range sourceCases {
		transition := ResolvePartialObservation(item.Observation)
		cases = append(cases, LatticeCase{
			ID: item.ID, Decision: decisionForTransition(transition),
			Observation: item.Observation, Transition: transition, ClaimID: item.ClaimID,
		})
		claims = append(claims, claimFromSource(item, transition))
	}
	return cases, claims, semanticDigest(sourceCases), nil
}

func parseGoooCases(sourcePath, source string) ([]sourceCase, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, source)
	if file == nil || diagnostics.HasErrors() {
		return nil, fmt.Errorf("parse Gooo lattice source: %d syntax diagnostics", len(diagnostics))
	}
	result := make([]sourceCase, 0, LatticeCaseDenominator)
	seen := make(map[string]bool)
	for _, declaration := range file.Decls {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || !strings.HasPrefix(activity.ValueProgram, latticeCaseProgramPrefix) {
			continue
		}
		item, err := parseCaseProgram(activity.ValueProgram)
		if err != nil {
			return nil, fmt.Errorf("activity %s: %w", activity.Name, err)
		}
		if seen[item.ID] {
			return nil, fmt.Errorf("duplicate lattice case %q", item.ID)
		}
		seen[item.ID] = true
		result = append(result, item)
	}
	if len(result) != LatticeCaseDenominator {
		return nil, fmt.Errorf("Gooo lattice case denominator = %d, want %d", len(result), LatticeCaseDenominator)
	}
	return result, nil
}

func parseCaseProgram(program string) (sourceCase, error) {
	fields := make(map[string]string)
	for _, field := range strings.Split(strings.TrimPrefix(program, latticeCaseProgramPrefix), ";") {
		key, value, ok := strings.Cut(field, "=")
		if !ok || strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			return sourceCase{}, fmt.Errorf("malformed case field %q", field)
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if _, exists := fields[key]; exists {
			return sourceCase{}, fmt.Errorf("duplicate case field %q", key)
		}
		fields[key] = value
	}
	for _, key := range []string{"id", "required", "observed", "reason", "repository_writes", "mutation_authority", "claim_id", "claim_state"} {
		if fields[key] == "" {
			return sourceCase{}, fmt.Errorf("missing case field %q", key)
		}
	}
	required, err := strconv.Atoi(fields["required"])
	if err != nil {
		return sourceCase{}, fmt.Errorf("required: %w", err)
	}
	observed, err := strconv.Atoi(fields["observed"])
	if err != nil {
		return sourceCase{}, fmt.Errorf("observed: %w", err)
	}
	writes, err := strconv.Atoi(fields["repository_writes"])
	if err != nil {
		return sourceCase{}, fmt.Errorf("repository_writes: %w", err)
	}
	authority, err := strconv.ParseBool(fields["mutation_authority"])
	if err != nil {
		return sourceCase{}, fmt.Errorf("mutation_authority: %w", err)
	}
	state := ClaimState(fields["claim_state"])
	if !validClaimState(state) {
		return sourceCase{}, fmt.Errorf("invalid claim_state %q", state)
	}
	return sourceCase{
		ID: fields["id"],
		Observation: PartialObservation{
			Required: required, Observed: observed, Reason: fields["reason"],
			RepositoryWrites: writes, MutationAuthority: authority,
		},
		ClaimID: fields["claim_id"], ExpectedClaimState: state,
	}, nil
}

func claimFromSource(item sourceCase, transition LatticeTransition) ClaimRecord {
	before := item.ExpectedClaimState
	after := before
	if transition.Decision == DecisionPass && before == ClaimOpen {
		after = ClaimDischarged
	}
	return ClaimRecord{ID: item.ClaimID, State: after, BeforeState: before, AfterState: after, Preserved: before == after}
}

func decisionForTransition(transition LatticeTransition) string {
	if transition.Decision == DecisionLowerResolution {
		return DecisionUnknown
	}
	return transition.Decision
}

func semanticDigest(sourceCases []sourceCase) string {
	var canonical strings.Builder
	for _, item := range sourceCases {
		fmt.Fprintf(&canonical, "%s|%d|%d|%s|%d|%t|%s|%s\n", item.ID,
			item.Observation.Required, item.Observation.Observed, item.Observation.Reason,
			item.Observation.RepositoryWrites, item.Observation.MutationAuthority,
			item.ClaimID, item.ExpectedClaimState)
	}
	digest := sha256.Sum256([]byte(canonical.String()))
	return "sha256:" + hex.EncodeToString(digest[:])
}
