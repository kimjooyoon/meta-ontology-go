package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
)

const (
	latticeSchema = "gooo/meta-semantic-resolution-lattice/v1"
	exact         = "exact_operation"
	invariantOnly = "invariant_only"
)

type observation struct {
	Required          int    `json:"required"`
	Observed          int    `json:"observed"`
	Reason            string `json:"reason"`
	RepositoryWrites  int    `json:"repository_writes"`
	MutationAuthority bool   `json:"mutation_authority"`
}

type unknownValue struct {
	Stage  string `json:"stage"`
	Step   int    `json:"step"`
	Reason string `json:"reason"`
}

type transition struct {
	FromResolution    string        `json:"from_resolution"`
	ToResolution      string        `json:"to_resolution,omitempty"`
	Decision          string        `json:"decision"`
	Reason            string        `json:"reason"`
	Unknown           *unknownValue `json:"unknown,omitempty"`
	RepositoryWrites  int           `json:"repository_writes"`
	MutationAuthority bool          `json:"mutation_authority"`
}

type latticeCase struct {
	ID          string      `json:"id"`
	Decision    string      `json:"decision"`
	Observation observation `json:"observation"`
	Transition  transition  `json:"transition"`
	ClaimID     string      `json:"claim_id"`
}

type claim struct {
	ID          string `json:"id"`
	State       string `json:"state"`
	BeforeState string `json:"before_state"`
	AfterState  string `json:"after_state"`
	Preserved   bool   `json:"preserved"`
}

type metric struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	Numerator     int    `json:"numerator"`
	Denominator   int    `json:"denominator"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Proof         string `json:"proof"`
}

type receipt struct {
	Schema            string `json:"schema"`
	Source            string `json:"source"`
	SourceSHA256      string `json:"source_sha256"`
	RepositoryWrites  int    `json:"repository_writes"`
	MutationAuthority bool   `json:"mutation_authority"`
	CaseDenominator   int    `json:"case_denominator"`
	Counts            struct {
		CasesTotal int `json:"cases_total"`
		Pass       int `json:"pass"`
		FailClosed int `json:"fail_closed"`
		Unknown    int `json:"unknown"`
	} `json:"counts"`
	Cases   []latticeCase `json:"cases"`
	Claims  []claim       `json:"claims"`
	Metrics []metric      `json:"metrics"`
}

func main() {
	source := flag.String("source", "examples/semantic-resolution-lattice/main.gooo", "Gooo source")
	receipt := flag.String("receipt", "examples/semantic-resolution-lattice/receipt.json", "receipt")
	check := flag.Bool("check", false, "require a valid receipt")
	flag.Parse()
	if !*check {
		fatal(errors.New("-check is required for the independent adjudicator"))
	}
	if err := validate(*source, *receipt); err != nil {
		fatal(err)
	}
	fmt.Println("semantic resolution lattice: PASS")
}

func validate(sourcePath, receiptPath string) error {
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	var got receipt
	data, err := os.ReadFile(receiptPath)
	if err != nil {
		return fmt.Errorf("read receipt: %w", err)
	}
	if err := json.Unmarshal(data, &got); err != nil {
		return fmt.Errorf("decode receipt: %w", err)
	}
	digest := sha256.Sum256(source)
	if got.Schema != latticeSchema || got.Source != sourcePath || got.SourceSHA256 != "sha256:"+hex.EncodeToString(digest[:]) {
		return errors.New("receipt identity is not bound to the source")
	}
	if got.RepositoryWrites != 0 || got.MutationAuthority || got.CaseDenominator != 4 || got.Counts.CasesTotal != 4 {
		return errors.New("effect or denominator guardrail failed")
	}
	if got.Counts.Pass != 1 || got.Counts.FailClosed != 2 || got.Counts.Unknown != 1 || len(got.Cases) != 4 {
		return errors.New("case counts are not the fixed contract")
	}
	if !hasGoooRelations(string(source)) {
		return errors.New("Gooo source relations are missing")
	}
	for _, item := range got.Cases {
		if err := validateCase(item); err != nil {
			return fmt.Errorf("case %s: %w", item.ID, err)
		}
	}
	if err := validateClaims(got.Claims, got.Cases); err != nil {
		return err
	}
	return validateMetrics(got.Metrics)
}

func adjudicate(input observation) transition {
	result := transition{FromResolution: exact, RepositoryWrites: input.RepositoryWrites, MutationAuthority: input.MutationAuthority}
	switch {
	case input.RepositoryWrites != 0:
		result.Decision, result.Reason = "FAIL_CLOSED", "REPOSITORY_WRITE_EFFECT"
	case input.MutationAuthority:
		result.Decision, result.Reason = "FAIL_CLOSED", "MUTATION_AUTHORITY_PRESENT"
	case input.Required <= 0 || input.Observed < 0 || input.Observed > input.Required:
		result.Decision, result.Reason = "FAIL_CLOSED", "OBSERVATION_CARDINALITY_INVALID"
	case input.Observed == input.Required:
		result.ToResolution, result.Decision, result.Reason = exact, "PASS", "OBSERVATION_COMPLETE"
	default:
		result.ToResolution, result.Decision, result.Reason = invariantOnly, "LOWER_RESOLUTION", "PARTIAL_OBSERVATION"
		result.Unknown = &unknownValue{Stage: "PARTIAL_OBSERVATION", Step: 1, Reason: input.Reason}
	}
	return result
}

func validateCase(item latticeCase) error {
	if item.ID == "" || item.ClaimID == "" || item.Transition.FromResolution != exact {
		return errors.New("invalid case identity")
	}
	if !reflect.DeepEqual(adjudicate(item.Observation), item.Transition) {
		return errors.New("independent transition replay disagrees")
	}
	wantDecision := item.Transition.Decision
	if wantDecision == "LOWER_RESOLUTION" {
		wantDecision = "UNKNOWN"
	}
	if item.Decision != wantDecision {
		return errors.New("case decision is not derived from transition")
	}
	return nil
}

func validateClaims(claims []claim, cases []latticeCase) error {
	if len(claims) != 4 {
		return errors.New("claim denominator is not fixed")
	}
	expected := map[string]string{
		"claim-exact-observation":            "DISCHARGED",
		"claim-invariant-fallback":           "OPEN",
		"claim-exact-under-missing-evidence": "REFUTED",
		"claim-write-free-descent":           "DISCHARGED",
	}
	seen := map[string]bool{}
	for _, item := range claims {
		want, known := expected[item.ID]
		if !known || seen[item.ID] || item.State != want || item.State != item.BeforeState || item.State != item.AfterState || !item.Preserved {
			return errors.New("claim state was not preserved")
		}
		seen[item.ID] = true
	}
	for _, item := range cases {
		if !seen[item.ClaimID] {
			return errors.New("case claim is not in the preserved ledger")
		}
	}
	return nil
}

func validateMetrics(metrics []metric) error {
	if len(metrics) != 5 {
		return errors.New("metric cardinality is invalid")
	}
	expected := map[string]struct {
		numerator                    int
		class, unit, relation, proof string
	}{
		"gooo.metric.meta-resolution-lattice.exact-observation.count.v1":  {1, "outcome", "cases", "greater_or_equal", "FOUNDATION"},
		"gooo.metric.meta-resolution-lattice.invariant-descent.count.v1":  {1, "driver", "cases", "greater_or_equal", "COHERENCE"},
		"gooo.metric.meta-resolution-lattice.claim-preservation.count.v1": {4, "driver", "cases", "greater_or_equal", "REGRESSION"},
		"gooo.metric.meta-resolution-lattice.replay.count.v1":             {4, "driver", "cases", "greater_or_equal", "REGRESSION"},
		"gooo.metric.meta-resolution-lattice.write-guardrail.v1":          {0, "guardrail", "repository_writes", "less_or_equal", "FOUNDATION"},
	}
	proofs := map[string]bool{}
	seen := map[string]bool{}
	for _, item := range metrics {
		want, known := expected[item.ID]
		if !known || seen[item.ID] || item.Numerator != want.numerator || item.Denominator != 4 || item.Numerator < 0 || item.Numerator > 4 || item.Class != want.class || item.Unit != want.unit || item.Relation != want.relation || item.Proof != want.proof || item.Producer == "" || item.Consumer == "" || item.MetaOperation == "" {
			return errors.New("metric is not fixed and provenance-bound")
		}
		seen[item.ID] = true
		proofs[item.Proof] = true
	}
	if len(proofs) != 3 || len(seen) != len(expected) {
		return errors.New("metric proof trilemma is incomplete")
	}
	return nil
}

func hasGoooRelations(source string) bool {
	return contains(source, "activity ObservePartialObservation") && contains(source, "activity DescendToInvariantOnly") && contains(source, "activity AdjudicateReceipt")
}

func contains(source, part string) bool {
	return len(source) >= len(part) && (source == part || len(source) > len(part) && containsAt(source, part))
}

func containsAt(source, part string) bool {
	for index := 0; index+len(part) <= len(source); index++ {
		if source[index:index+len(part)] == part {
			return true
		}
	}
	return false
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
