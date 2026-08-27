package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

const (
	expectedSchema       = "gooo/meta.phase-separation-witness/v1"
	expectedPass         = "PASS"
	expectedUnknown      = "UNKNOWN"
	expectedExact        = "EXACT"
	expectedLower        = "LOWER_RESOLUTION"
	expectedReason       = "PHASE_SEPARATION_WITNESS_EXACT"
	expectedUnknownFault = "UNKNOWN_SOURCE_SYNTAX"
	expectedToolchain    = "go1.27.0"
)

type coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type caseResult struct {
	Name            string `json:"name"`
	Class           string `json:"class"`
	Expected        string `json:"expected"`
	Actual          string `json:"actual"`
	Reason          string `json:"reason"`
	Passed          bool   `json:"passed"`
	TransitionCount int    `json:"transition_count"`
}

type transition struct {
	ID            string `json:"id"`
	FromPhase     string `json:"from_phase"`
	ToPhase       string `json:"to_phase"`
	FromClaim     string `json:"from_claim"`
	ToClaim       string `json:"to_claim"`
	FromState     string `json:"from_state"`
	ToState       string `json:"to_state"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Preserved     bool   `json:"preserved"`
}

type indicator struct {
	ID            string `json:"id"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Numerator     int    `json:"numerator"`
	Denominator   int    `json:"denominator"`
	Satisfied     bool   `json:"satisfied"`
}

type view struct {
	Audience      string `json:"audience"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Satisfied     int    `json:"satisfied"`
	Total         int    `json:"total"`
	BasisPoints   int    `json:"basis_points"`
}

type proof struct {
	Choice         string `json:"choice"`
	Claim          string `json:"claim"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type summary struct {
	CleanCasesPassed          int `json:"clean_cases_passed"`
	CleanCasesTotal           int `json:"clean_cases_total"`
	LeakageCasesCaught        int `json:"leakage_cases_caught"`
	LeakageCasesTotal         int `json:"leakage_cases_total"`
	ClaimTransitionsPreserved int `json:"claim_transitions_preserved"`
	ClaimTransitionsTotal     int `json:"claim_transitions_total"`
	IndicatorsSatisfied        int `json:"indicators_satisfied"`
	IndicatorsTotal            int `json:"indicators_total"`
	UnknownCases               int `json:"unknown_cases"`
	RepositoryWrites           int `json:"repository_writes"`
}

type authority struct {
	Execution bool `json:"execution"`
	Mutation  bool `json:"mutation"`
	Promotion bool `json:"promotion"`
}

// These receipt types intentionally live in the adjudicator instead of being
// imported from the producer package. The consumer is a separate trust root.
type receipt struct {
	Schema           string        `json:"schema"`
	Decision         string        `json:"decision"`
	Reason           string        `json:"reason"`
	Resolution       string        `json:"resolution"`
	HeadSHA          string        `json:"head_sha"`
	Toolchain        string        `json:"toolchain"`
	SourcePath       string        `json:"source_path"`
	SourceDigest     string        `json:"source_digest"`
	LeakSourcePath   string        `json:"leak_source_path"`
	LeakSourceDigest string        `json:"leak_source_digest"`
	Producer         string        `json:"producer"`
	Consumer         string        `json:"consumer"`
	MetaOperation    string        `json:"meta_operation"`
	ProofChoice      string        `json:"proof_choice"`
	Cases            []caseResult  `json:"cases"`
	Transitions      []transition  `json:"claim_transitions"`
	Indicators       []indicator    `json:"indicators"`
	Views            []view         `json:"views"`
	Proofs           []proof        `json:"proofs"`
	Summary          summary       `json:"summary"`
	Authority        authority     `json:"authority"`
	Coordinate       coordinate    `json:"coordinate"`
	Digest           string        `json:"digest"`
}

func main() {
	receiptPath := flag.String("receipt", "", "receipt to adjudicate")
	expectedHead := flag.String("expected-head", "", "expected exact source commit")
	mode := flag.String("mode", "proven", "proven or unknown")
	flag.Parse()
	if err := run(*receiptPath, *expectedHead, *mode); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("independent adjudicator: %s\n", *mode)
}

func run(path, expectedHead, mode string) error {
	if path == "" || expectedHead == "" {
		return fmt.Errorf("receipt and expected-head are required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read receipt: %w", err)
	}
	var got receipt
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&got); err != nil {
		return fmt.Errorf("decode receipt: %w", err)
	}
	if got.Digest != digestReceipt(got) {
		return fmt.Errorf("receipt digest mismatch")
	}
	if got.HeadSHA != expectedHead || got.Schema != expectedSchema || got.Toolchain != expectedToolchain {
		return fmt.Errorf("receipt identity is not exact")
	}
	switch mode {
	case "proven":
		return checkProven(got)
	case "unknown":
		return checkUnknown(got)
	default:
		return fmt.Errorf("unknown adjudication mode %q", mode)
	}
}

func checkProven(got receipt) error {
	if got.Decision != expectedPass || got.Reason != expectedReason || got.Resolution != expectedExact || got.Coordinate != (coordinate{"EXECUTION", "ADJUDICATE", expectedReason}) {
		return fmt.Errorf("proven decision or coordinate is not exact")
	}
	if got.Producer == "" || got.Consumer == "" || got.MetaOperation == "" || got.ProofChoice == "" || got.SourcePath == "" || got.LeakSourcePath == "" {
		return fmt.Errorf("producer, consumer, operation, proof, or source binding is missing")
	}
	if len(got.Cases) != 6 || len(got.Transitions) != 2 || len(got.Indicators) != 12 || len(got.Views) != 3 || len(got.Proofs) != 3 {
		return fmt.Errorf("fixed denominator changed")
	}
	expectedCases := []struct {
		name, class, expected, actual, reason string
	}{
		{"clean", "CLEAN", "ACCEPT", "ACCEPT", ""},
		{"value-leak", "LEAKAGE", "REJECT_LEAK", "REJECT_LEAK", "VALUE_CROSSES_PHASE"},
		{"authority-leak", "LEAKAGE", "REJECT_LEAK", "REJECT_LEAK", "AUTHORITY_CROSSES_PHASE"},
		{"evidence-leak", "LEAKAGE", "REJECT_LEAK", "REJECT_LEAK", "EVIDENCE_CROSSES_PHASE"},
		{"phase-skip", "LEAKAGE", "REJECT_LEAK", "REJECT_LEAK", "PHASE_EDGE_SKIPS"},
		{"phase-reverse", "LEAKAGE", "REJECT_LEAK", "REJECT_LEAK", "PHASE_EDGE_REVERSES"},
	}
	for index, want := range expectedCases {
		gotCase := got.Cases[index]
		if gotCase.Name != want.name || gotCase.Class != want.class || gotCase.Expected != want.expected || gotCase.Actual != want.actual || gotCase.Reason != want.reason || !gotCase.Passed {
			return fmt.Errorf("case %s is not independently proven", want.name)
		}
	}
	for index, gotTransition := range got.Transitions {
		if !gotTransition.Preserved || gotTransition.FromState != "DECLARED" || gotTransition.ToState != "PRESERVED" || gotTransition.MetaOperation != got.MetaOperation || gotTransition.ProofChoice != got.ProofChoice {
			return fmt.Errorf("claim transition %d is not preserved", index)
		}
	}
	for index, gotIndicator := range got.Indicators {
		if gotIndicator.ID != fmt.Sprintf("PHASE-%02d", index+1) || gotIndicator.Numerator != 1 || gotIndicator.Denominator != 1 || !gotIndicator.Satisfied {
			return fmt.Errorf("indicator %d is not exact", index)
		}
	}
	for index, gotView := range got.Views {
		want := []struct{ audience string; total int }{{"PRODUCER", 3}, {"CONSUMER", 9}, {"GOVERNOR", 12}}[index]
		if gotView.Audience != want.audience || gotView.Satisfied != want.total || gotView.Total != want.total || gotView.BasisPoints != 10000 {
			return fmt.Errorf("view %d is not exact", index)
		}
	}
	for _, gotProof := range got.Proofs {
		if gotProof.Choice == "" || gotProof.MetaOperation != got.MetaOperation || gotProof.EvidenceDigest == "" || !gotProof.Passed {
			return fmt.Errorf("proof is not independently discharged")
		}
	}
	if got.Summary != (summary{1, 1, 5, 5, 2, 2, 12, 12, 0, 0}) || got.Authority != (authority{}) {
		return fmt.Errorf("summary or authority boundary is not exact")
	}
	return nil
}

func checkUnknown(got receipt) error {
	if got.Decision != expectedUnknown || got.Reason != expectedUnknownFault || got.Resolution != expectedLower || got.Coordinate != (coordinate{"SOURCE", "PARSE", expectedUnknownFault}) {
		return fmt.Errorf("unknown result did not fail closed with a coordinate")
	}
	if len(got.Cases) != 0 || len(got.Transitions) != 0 || len(got.Indicators) != 0 || len(got.Views) != 0 || len(got.Proofs) != 0 || got.Summary != (summary{}) || got.Authority != (authority{}) {
		return fmt.Errorf("unknown result contains executable or authority evidence")
	}
	return nil
}

func digestReceipt(value receipt) string {
	value.Digest = ""
	encoded, _ := json.Marshal(value)
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:])
}
