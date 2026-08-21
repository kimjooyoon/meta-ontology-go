package semantic

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidNode = errors.New("invalid semantic node")
	ErrInvalidSpan = errors.New("invalid source span")
)

// Kind is the PROV-inspired role of a semantic node.
type Kind string

const (
	Entity   Kind = "Entity"
	Activity Kind = "Activity"
	Agent    Kind = "Agent"

	EntityKind   = Entity
	ActivityKind = Activity
	AgentKind    = Agent
)

func (k Kind) String() string {
	return string(k)
}
func (k Kind) Valid() bool {
	switch k {
	case Entity, Activity, Agent:
		return true
	default:
		return false
	}
}

// Position is a source position. Zero means unknown; offsets, lines, and
// columns are otherwise expected to be non-negative.
type Position struct {
	Offset int
	Line   int
	Column int
}

func (p Position) valid() bool {
	return p.Offset >= 0 && p.Line >= 0 && p.Column >= 0
}

// Span records source provenance without coupling the semantic package to a
// lexer or parser implementation.
type Span struct {
	File  string
	Start Position
	End   Position
}

func (s Span) IsZero() bool {
	return s.File == "" && s.Start == (Position{}) && s.End == (Position{})
}
func (s Span) Normalized() Span {
	s.File = strings.TrimSpace(s.File)
	return s
}
func (s Span) Validate() error {
	if !s.Start.valid() || !s.End.valid() {
		return fmt.Errorf("%w: positions must be non-negative", ErrInvalidSpan)
	}
	if s.End.Offset < s.Start.Offset {
		return fmt.Errorf("%w: end offset precedes start offset", ErrInvalidSpan)
	}
	return nil
}
