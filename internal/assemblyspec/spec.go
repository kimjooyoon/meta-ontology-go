// Package assemblyspec carries source-owned finite body assembly intent through
// syntax, bidirectional lowering and semantic IR without model or runtime state.
package assemblyspec

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

type Spec struct {
	Choices            []Choice            `json:"choices"`
	Cases              []Case              `json:"cases"`
	HoldoutCases       []Case              `json:"holdout_cases,omitempty"`
	ValueCases         []ValueCase         `json:"value_cases,omitempty"`
	ValueHoldoutCases  []ValueCase         `json:"value_holdout_cases,omitempty"`
	MaxAttempts        int                 `json:"max_attempts"`
	Seed               string              `json:"seed,omitempty"`
	Baseline           string              `json:"baseline,omitempty"`
	Picked             []Pick              `json:"picked,omitempty"`
	Search             *Search             `json:"search,omitempty"`
	SearchAlternatives []SearchAlternative `json:"search_alternatives,omitempty"`
	FillPlan           *FillPlan           `json:"fill_plan,omitempty"`
}

// Search declares a bounded IR expression grammar whose candidates Gooo derives
// from the activity's training cases. It contains no executable model output.
type Search struct {
	HoleID        string `json:"hole_id"`
	Grammar       string `json:"grammar"`
	Intent        string `json:"intent"`
	MaxCandidates int    `json:"max_candidates"`
}

// SearchAlternative is a source-declared next grammar/candidate bound. The hole,
// intent and expected behavior stay in the same assembly contract.
type SearchAlternative struct {
	ID            string `json:"id"`
	Grammar       string `json:"grammar"`
	MaxCandidates int    `json:"max_candidates"`
}

// FillPlan declares complete candidate assignments for multiple typed holes.
// Models can rank these assignments but cannot add or rewrite expressions.
type FillPlan struct {
	Intent     string          `json:"intent"`
	Holes      []FillHole      `json:"holes"`
	Candidates []FillCandidate `json:"candidates"`
	Generation *FillGeneration `json:"generation,omitempty"`
}

// FillGeneration declares a closed expression grammar and assignment-space cap
// that Gooo applies to each declared hole using only source-owned training cases.
type FillGeneration struct {
	Grammar        string            `json:"grammar"`
	MaxExpressions int               `json:"max_expressions_per_hole"`
	MaxCandidates  int               `json:"max_candidates"`
	HoleGrammars   []FillHoleGrammar `json:"hole_grammars,omitempty"`
}

type FillHoleGrammar struct {
	HoleID         string `json:"hole_id"`
	Grammar        string `json:"grammar"`
	MaxExpressions int    `json:"max_expressions"`
}

type FillHole struct {
	ID string `json:"id"`
}

type FillCandidate struct {
	ID    string        `json:"id"`
	Fills []HoleFilling `json:"fills"`
}

type HoleFilling struct {
	HoleID     string `json:"hole_id"`
	Expression string `json:"expression"`
}

// Pick records an implementation relative to the immutable planning baseline.
type Pick struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type Choice struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Occurrence  int    `json:"occurrence"`
	Intent      string `json:"intent"`
	Alternative string `json:"alternative_name,omitempty"`
}

type Case struct {
	Input    int64 `json:"input"`
	Expected int64 `json:"expected"`
}

// ValueCase carries canonical JSON: positional inputs and a named record result.
type ValueCase struct {
	Inputs   string `json:"inputs"`
	Expected string `json:"expected"`
}

func (s Spec) Validate() error {
	if err := s.validateSearchAlternatives(); err != nil {
		return err
	}
	if len(s.Cases)+len(s.HoldoutCases)+len(s.ValueCases)+len(s.ValueHoldoutCases) < 1 ||
		len(s.Cases)+len(s.HoldoutCases)+len(s.ValueCases)+len(s.ValueHoldoutCases) > 128 || len(s.Seed) > 512 || !utf8.ValidString(s.Seed) {
		return fmt.Errorf("assembly requires 1..16 choices, 1..128 cases and 1..64 attempts")
	}
	if s.FillPlan != nil {
		return s.validateFillPlan()
	}
	if s.MaxAttempts < 1 || s.MaxAttempts > 64 {
		return fmt.Errorf("assembly requires 1..16 choices, 1..128 cases and 1..64 attempts")
	}
	if s.Search != nil {
		if len(s.Cases) == 0 || len(s.Choices) != 0 || len(s.ValueCases)+len(s.ValueHoldoutCases) != 0 ||
			s.Baseline != "" || len(s.Picked) != 0 || s.Seed != "" ||
			!identifier(s.Search.HoleID) || (s.Search.Grammar != "integer-offset-constant/v1" && s.Search.Grammar != "integer-hole-residual/v1") ||
			!boundedText(s.Search.Intent, 2000) || s.Search.MaxCandidates < 2 || s.Search.MaxCandidates > 16 ||
			s.MaxAttempts > s.Search.MaxCandidates {
			return fmt.Errorf("IR search assembly requires a hole, supported grammar, intent, 2..16 candidates, bounded cases and attempts, and no path choices/checkpoint")
		}
		training := make(map[int64]bool, len(s.Cases))
		for _, c := range s.Cases {
			training[c.Input] = true
		}
		for _, c := range s.HoldoutCases {
			if training[c.Input] {
				return fmt.Errorf("IR search holdout input %d also appears in training cases", c.Input)
			}
		}
		return nil
	}
	if len(s.HoldoutCases) != 0 || len(s.Choices) < 1 || len(s.Choices) > 16 {
		return fmt.Errorf("assembly requires 1..16 choices and holdout cases are available only to IR search")
	}
	var seen [16]string
	for i, c := range s.Choices {
		duplicate := false
		for _, id := range seen[:i] {
			duplicate = duplicate || id == c.ID
		}
		if !boundedText(c.ID, 128) || duplicate || !boundedText(c.Intent, 512) ||
			c.Occurrence < 0 || c.Occurrence > 127 {
			return fmt.Errorf("assembly choice %q needs a unique ID, intent and occurrence in 0..127", c.ID)
		}
		seen[i] = c.ID
		switch c.Kind {
		case "operand_order", "branch_layout", "root_order":
			if c.Alternative != "" {
				return fmt.Errorf("assembly layout choice %q cannot name an alternative local", c.ID)
			}
		case "local_reference", "assignment_target":
			if !boundedText(c.Alternative, 128) {
				return fmt.Errorf("assembly name choice %q requires an alternative local", c.ID)
			}
		case "field_value", "field_update":
			if !boundedText(c.Alternative, 512) || len(s.Cases) != 0 || len(s.ValueCases) == 0 || len(s.ValueHoldoutCases) != 0 || len(s.Choices) > 6 {
				return fmt.Errorf("field assembly requires an alternative expression, 1..6 choices and value_case expectations")
			}
		default:
			return fmt.Errorf("unknown assembly choice kind %q", c.Kind)
		}
	}
	if len(s.ValueCases) != 0 || len(s.ValueHoldoutCases) != 0 {
		if len(s.ValueHoldoutCases) != 0 {
			return fmt.Errorf("holdout_value_case is available only for source_fill plans")
		}
		for _, choice := range s.Choices {
			if choice.Kind != "field_value" && choice.Kind != "field_update" {
				return fmt.Errorf("value_case requires only field_value or field_update choices")
			}
		}
		for _, c := range s.ValueCases {
			input, err := CanonicalValue(c.Inputs)
			expected, next := CanonicalValue(c.Expected)
			if err != nil || next != nil || input != c.Inputs || expected != c.Expected {
				return fmt.Errorf("value_case requires bounded canonical JSON")
			}
		}
	}
	if err := s.validateCheckpoint(); err != nil {
		return err
	}
	return nil
}

func (s Spec) validateFillPlan() error {
	plan := s.FillPlan
	recordCases := len(s.ValueCases) > 0
	if len(s.Cases) == 0 && !recordCases || len(s.Cases) > 0 && recordCases ||
		len(s.Choices) != 0 || s.Search != nil || s.MaxAttempts != 0 || s.Seed != "" ||
		s.Baseline != "" || len(s.Picked) != 0 || !boundedText(plan.Intent, 2000) ||
		len(plan.Holes) < 2 || len(plan.Holes) > 8 {
		return fmt.Errorf("source fill plan requires intent, 2..8 holes and either integer cases or typed record value cases; it cannot mix with other assembly modes")
	}
	if recordCases {
		if len(s.HoldoutCases) != 0 {
			return fmt.Errorf("record source fill uses holdout_value_case, not holdout_case")
		}
		trainingInputs := make(map[string]bool, len(s.ValueCases))
		for _, testCase := range s.ValueCases {
			inputs, err := CanonicalValue(testCase.Inputs)
			if err != nil || inputs != testCase.Inputs {
				return fmt.Errorf("record source fill inputs must be bounded canonical JSON")
			}
			expected, err := CanonicalValue(testCase.Expected)
			if err != nil || expected != testCase.Expected {
				return fmt.Errorf("record source fill expected values must be bounded canonical JSON")
			}
			trainingInputs[testCase.Inputs] = true
		}
		for _, testCase := range s.ValueHoldoutCases {
			inputs, err := CanonicalValue(testCase.Inputs)
			if err != nil || inputs != testCase.Inputs {
				return fmt.Errorf("record source fill holdout inputs must be bounded canonical JSON")
			}
			expected, err := CanonicalValue(testCase.Expected)
			if err != nil || expected != testCase.Expected {
				return fmt.Errorf("record source fill holdout expected values must be bounded canonical JSON")
			}
			if trainingInputs[testCase.Inputs] {
				return fmt.Errorf("record source fill holdout input %s also appears in training cases", testCase.Inputs)
			}
		}
	} else if len(s.ValueHoldoutCases) != 0 {
		return fmt.Errorf("holdout_value_case requires record value_case training examples")
	}
	trainingInputs := make(map[int64]bool, len(s.Cases))
	for _, testCase := range s.Cases {
		trainingInputs[testCase.Input] = true
	}
	for _, testCase := range s.HoldoutCases {
		if trainingInputs[testCase.Input] {
			return fmt.Errorf("source fill holdout input %d also appears in training cases", testCase.Input)
		}
	}
	if plan.Generation == nil {
		if len(plan.Candidates) < 2 || len(plan.Candidates) > 16 {
			return fmt.Errorf("source fill plan requires 2..16 complete candidates or a bounded derive clause")
		}
	} else {
		if len(plan.Candidates) != 0 || plan.Generation.MaxCandidates < 2 || plan.Generation.MaxCandidates > 16 {
			return fmt.Errorf("source fill derive requires 2..16 complete candidates and no manual candidates")
		}
		if len(plan.Generation.HoleGrammars) == 0 {
			if !fillGrammarAllowed(plan.Generation.Grammar, recordCases) || plan.Generation.MaxExpressions < 2 || plan.Generation.MaxExpressions > 16 {
				return fmt.Errorf("source fill derive requires a supported grammar and 2..16 expressions per hole")
			}
		} else {
			if plan.Generation.Grammar != "" || plan.Generation.MaxExpressions != 0 || len(plan.Generation.HoleGrammars) != len(plan.Holes) {
				return fmt.Errorf("source fill per-hole derive requires exactly one grammar per hole and no shared grammar")
			}
			for index, grammar := range plan.Generation.HoleGrammars {
				if grammar.HoleID != plan.Holes[index].ID || !fillGrammarAllowed(grammar.Grammar, recordCases) ||
					grammar.MaxExpressions < 2 || grammar.MaxExpressions > 16 {
					return fmt.Errorf("source fill per-hole derive entries must match hole order and use a supported grammar with 2..16 expressions")
				}
			}
		}
	}
	var holes [8]string
	for index, hole := range plan.Holes {
		if !identifier(hole.ID) {
			return fmt.Errorf("source fill plan hole %q has an invalid id", hole.ID)
		}
		if slices.Contains(holes[:index], hole.ID) {
			return fmt.Errorf("source fill plan hole %q is duplicated", hole.ID)
		}
		holes[index] = hole.ID
	}
	if plan.Generation != nil {
		return nil
	}
	var candidateIDs [16]string
	for index, candidate := range plan.Candidates {
		if !identifier(candidate.ID) || len(candidate.Fills) != len(plan.Holes) {
			return fmt.Errorf("source fill candidate %q must have a valid id and fill every declared hole", candidate.ID)
		}
		if slices.Contains(candidateIDs[:index], candidate.ID) {
			return fmt.Errorf("source fill candidate %q is duplicated", candidate.ID)
		}
		candidateIDs[index] = candidate.ID
		for fillIndex, fill := range candidate.Fills {
			if fill.HoleID != plan.Holes[fillIndex].ID || !boundedText(fill.Expression, 4096) {
				return fmt.Errorf("source fill candidate %q must fill holes once in declaration order with bounded expressions", candidate.ID)
			}
		}
	}
	return nil
}

func supportedFillGrammar(grammar string) bool {
	return grammar == "integer-offset-constant/v1" || grammar == "integer-predicate/v1" ||
		grammar == "integer-predicate-composition/v1" || grammar == "integer-predicate-outside-range/v1" ||
		grammar == "integer-predicate-cutpoint/v1" || grammar == "record-field-predicate/v1" ||
		grammar == "record-field-predicate/v2" ||
		grammar == "record-field-relation/v1" ||
		grammar == "record-field-relation-composition/v1" ||
		grammar == "record-field-relation-composition/v2" ||
		grammar == "record-field-relation-composition/v3" ||
		grammar == "record-field-predicate-composition/v1" ||
		grammar == "record-field-predicate-composition/v2" ||
		grammar == "record-string-literal/v1" || grammar == "record-integer-literal/v1" ||
		grammar == "record-boolean-literal/v1"
}

func fillGrammarAllowed(grammar string, recordCases bool) bool {
	if recordCases {
		return recordFillGrammar(grammar)
	}
	return supportedFillGrammar(grammar) && !recordFillGrammar(grammar)
}

func recordFillGrammar(grammar string) bool {
	return grammar == "record-field-predicate/v1" || grammar == "record-field-predicate/v2" ||
		grammar == "record-field-relation/v1" ||
		grammar == "record-field-relation-composition/v1" ||
		grammar == "record-field-relation-composition/v2" ||
		grammar == "record-field-relation-composition/v3" ||
		grammar == "record-field-predicate-composition/v1" ||
		grammar == "record-field-predicate-composition/v2" ||
		grammar == "record-string-literal/v1" ||
		grammar == "record-integer-literal/v1" || grammar == "record-boolean-literal/v1"
}

func (s Spec) validateCheckpoint() error {
	if s.Baseline == "" && len(s.Picked) == 0 {
		return nil
	}
	if !boundedText(s.Baseline, 128<<10) || len(s.Picked) != len(s.Choices) {
		return fmt.Errorf("assembly checkpoint requires a baseline and one picked label per choice")
	}
	for i, c := range s.Choices {
		first, second := "layout_forward", "layout_reverse"
		switch c.Kind {
		case "local_reference":
			first, second = "reference_first", "reference_second"
		case "assignment_target":
			first, second = "assign_first", "assign_second"
		case "root_order":
			first, second = "schedule_forward", "schedule_reverse"
		case "field_value", "field_update":
			first, second = "value_first", "value_second"
		}
		if s.Picked[i].ID != c.ID || (s.Picked[i].Label != first && s.Picked[i].Label != second) {
			return fmt.Errorf("assembly picked labels must match declared choice order and kind")
		}
	}
	return nil
}

func boundedText(text string, limit int) bool {
	return strings.TrimSpace(text) != "" && len(text) <= limit && utf8.ValidString(text)
}

func identifier(value string) bool {
	if value == "" || !((value[0] >= 'a' && value[0] <= 'z') || (value[0] >= 'A' && value[0] <= 'Z') || value[0] == '_') {
		return false
	}
	for index := 1; index < len(value); index++ {
		char := value[index]
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	return len(value) <= 128
}

func (s Spec) Clone() *Spec {
	clone := s
	clone.Choices = append([]Choice(nil), s.Choices...)
	clone.Cases = append([]Case(nil), s.Cases...)
	clone.ValueCases = append([]ValueCase(nil), s.ValueCases...)
	clone.ValueHoldoutCases = append([]ValueCase(nil), s.ValueHoldoutCases...)
	clone.Picked = append([]Pick(nil), s.Picked...)
	clone.SearchAlternatives = append([]SearchAlternative(nil), s.SearchAlternatives...)
	if s.Search != nil {
		search := *s.Search
		clone.Search = &search
	}
	if s.FillPlan != nil {
		plan := *s.FillPlan
		plan.Holes = append([]FillHole(nil), s.FillPlan.Holes...)
		plan.Candidates = append([]FillCandidate(nil), s.FillPlan.Candidates...)
		if s.FillPlan.Generation != nil {
			generation := *s.FillPlan.Generation
			generation.HoleGrammars = append([]FillHoleGrammar(nil), s.FillPlan.Generation.HoleGrammars...)
			plan.Generation = &generation
		}
		for index := range plan.Candidates {
			plan.Candidates[index].Fills = append([]HoleFilling(nil), s.FillPlan.Candidates[index].Fills...)
		}
		clone.FillPlan = &plan
	}
	clone.HoldoutCases = append([]Case(nil), s.HoldoutCases...)
	return &clone
}

// Canonical preserves declared choice and finite case order; neither is a set.
func (s Spec) Canonical() (string, error) {
	if err := s.Validate(); err != nil {
		return "", err
	}
	raw, err := json.Marshal(s)
	return string(raw), err
}

// DecodeCanonical accepts only the exact normalized carrier emitted by lowering.
func DecodeCanonical(raw string) (*Spec, error) {
	if len(raw) == 0 || len(raw) > 128<<10 {
		return nil, fmt.Errorf("assembly carrier must be bounded canonical JSON")
	}
	var s Spec
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return nil, err
	}
	canonical, err := s.Canonical()
	if err != nil {
		return nil, err
	}
	if canonical != raw {
		return nil, fmt.Errorf("assembly carrier is not canonical")
	}
	return s.Clone(), nil
}
