// Package proofchoicejudge is intentionally independent from the producer.
// It decodes the receipt wire format and recomputes the verdict from raw
// evidence, rather than importing proofchoicealgebra's evaluator.
package proofchoicejudge

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

const (
	schema           = "gooo/proof-choice-algebra-receipt/v1"
	fixedDenominator = 3
	pass             = "PASS"
	failClosed       = "FAIL_CLOSED"
)

type item struct {
	Kind          string `json:"kind"`
	ID            string `json:"id"`
	Statement     string `json:"statement"`
	Choice        string `json:"choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Numerator     int    `json:"numerator,omitempty"`
	Denominator   int    `json:"denominator,omitempty"`
	Line          int    `json:"line"`
}

type transition struct {
	ClaimID       string `json:"claim_id"`
	From          string `json:"from"`
	To            string `json:"to"`
	Choice        string `json:"choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	Persistent    bool   `json:"persistent"`
	Line          int    `json:"line"`
}

type indicator struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Choice   string `json:"choice"`
	Decision string `json:"decision"`
	Relation string `json:"relation"`
	Value    string `json:"value"`
	Limit    string `json:"limit"`
}

type summary struct {
	Claims                int `json:"claims"`
	Metrics               int `json:"metrics"`
	Items                 int `json:"items"`
	Transitions           int `json:"transitions"`
	PersistentTransitions int `json:"persistent_transitions"`
	ChoicesExplicit       int `json:"choices_explicit"`
	ChoiceCoverageBPS     int `json:"choice_coverage_bps"`
	FixedDenominator      int `json:"fixed_denominator"`
	Unknowns              int `json:"unknowns"`
	Contradictions        int `json:"contradictions"`
	Compositions          int `json:"compositions"`
}

type effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type receipt struct {
	Schema       string       `json:"schema"`
	Decision     string       `json:"decision"`
	Reason       string       `json:"reason"`
	Resolution   string       `json:"resolution"`
	SourcePath   string       `json:"source_path"`
	SourceDigest string       `json:"source_digest"`
	FixedDenom   int          `json:"fixed_denominator"`
	Items        []item       `json:"items"`
	Transitions  []transition `json:"transitions"`
	Indicators   []indicator  `json:"indicators"`
	Summary      summary      `json:"summary"`
	Effects      effects      `json:"effects"`
	Digest       string       `json:"digest"`
}

type Verdict struct {
	Schema         string `json:"schema"`
	Decision       string `json:"decision"`
	Reason         string `json:"reason"`
	ReceiptDigest  string `json:"receipt_digest"`
	ComputedDigest string `json:"computed_digest"`
	DigestMatch    bool   `json:"digest_match"`
	Items          int    `json:"items"`
	Transitions    int    `json:"transitions"`
	Independent    bool   `json:"independent"`
}

func Judge(data []byte) Verdict {
	var input receipt
	result := Verdict{Schema: "gooo/proof-choice-algebra-judge/v1", Independent: true}
	if err := json.Unmarshal(data, &input); err != nil {
		result.Decision, result.Reason = failClosed, "RECEIPT_JSON_UNKNOWN"
		return result
	}
	result.ReceiptDigest = input.Digest
	result.Items, result.Transitions = len(input.Items), len(input.Transitions)
	computed, err := digest(input)
	if err != nil {
		result.Decision, result.Reason = failClosed, "RECEIPT_DIGEST_UNKNOWN"
		return result
	}
	result.ComputedDigest, result.DigestMatch = computed, computed == input.Digest
	reason := validate(input)
	want := pass
	if reason != "" {
		want = failClosed
	}
	if !result.DigestMatch {
		want, reason = failClosed, "RECEIPT_DIGEST_MISMATCH"
	}
	if input.Decision != want && reason == "" {
		want, reason = failClosed, "PRODUCER_DECISION_MISMATCH"
	}
	result.Decision, result.Reason = want, reason
	if result.Decision == pass {
		result.Reason = "PROOF_CHOICES_COMPOSED"
	}
	return result
}

func digest(input receipt) (string, error) {
	input.Digest = ""
	data, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func validate(input receipt) string {
	if input.Schema != schema || input.SourcePath == "" || !validDigest(input.SourceDigest) || input.FixedDenom != fixedDenominator {
		return "RECEIPT_FOUNDATION_UNKNOWN"
	}
	if input.Effects.RepositoryWrites != 0 || input.Effects.MutationAuthority {
		return "READ_ONLY_GUARD_FAILED"
	}
	if len(input.Items) == 0 {
		return "NO_PROOF_CHOICES"
	}
	claims := map[string]string{}
	for _, value := range input.Items {
		if value.ID == "" || value.Statement == "" || !validChoice(value.Choice) {
			return "PROOF_CHOICE_MISSING"
		}
		if unknown(value.Producer, value.Consumer, value.MetaOperation, value.Stage, value.Step, value.Reason) {
			return "UNKNOWN_CONTEXT"
		}
		if value.Kind != "CLAIM" && value.Kind != "METRIC" {
			return "ITEM_KIND_UNKNOWN"
		}
		if value.Kind == "METRIC" && (value.Denominator != fixedDenominator || value.Numerator < 0 || value.Numerator > value.Denominator) {
			return "FIXED_DENOMINATOR_MISMATCH"
		}
		if old, exists := claims[value.ID]; exists && old != value.Choice {
			return "PROOF_CHOICE_CONTRADICTION"
		}
		if value.Kind == "CLAIM" {
			claims[value.ID] = value.Choice
		}
	}
	seen := map[string]bool{}
	for _, value := range input.Transitions {
		choice, exists := claims[value.ClaimID]
		if !exists || !value.Persistent || value.From == "" || value.To == "" || !validChoice(value.Choice) || choice != value.Choice {
			return "PERSISTENT_TRANSITION_MISMATCH"
		}
		if unknown(value.Producer, value.Consumer, value.MetaOperation, value.Stage, value.Step, value.Reason) || unknown(value.From, value.To) {
			return "UNKNOWN_CONTEXT"
		}
		seen[value.ClaimID] = true
	}
	for claimID := range claims {
		if !seen[claimID] {
			return "PERSISTENT_TRANSITION_MISSING"
		}
	}
	return ""
}

func validChoice(value string) bool {
	return value == "FOUNDATION" || value == "COHERENCE" || value == "REGRESSION"
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func unknown(values ...string) bool {
	for _, value := range values {
		if value == "" || strings.EqualFold(value, "UNKNOWN") {
			return true
		}
	}
	return false
}
