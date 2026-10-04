// Package assemblyspec carries source-owned finite body assembly intent through
// syntax, bidirectional lowering and semantic IR without model or runtime state.
package assemblyspec

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

type Spec struct {
	Choices     []Choice `json:"choices"`
	Cases       []Case   `json:"cases"`
	MaxAttempts int      `json:"max_attempts"`
	Seed        string   `json:"seed,omitempty"`
	Baseline    string   `json:"baseline,omitempty"`
	Picked      []Pick   `json:"picked,omitempty"`
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

func (s Spec) Validate() error {
	if len(s.Choices) < 1 || len(s.Choices) > 16 || len(s.Cases) < 1 || len(s.Cases) > 128 ||
		s.MaxAttempts < 1 || s.MaxAttempts > 64 || len(s.Seed) > 512 || !utf8.ValidString(s.Seed) {
		return fmt.Errorf("assembly requires 1..16 choices, 1..128 cases and 1..64 attempts")
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
		default:
			return fmt.Errorf("unknown assembly choice kind %q", c.Kind)
		}
	}
	if err := s.validateCheckpoint(); err != nil {
		return err
	}
	return nil
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

func (s Spec) Clone() *Spec {
	clone := s
	clone.Choices = append([]Choice(nil), s.Choices...)
	clone.Cases = append([]Case(nil), s.Cases...)
	clone.Picked = append([]Pick(nil), s.Picked...)
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
