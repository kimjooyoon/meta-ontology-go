package query

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidNode     = errors.New("invalid query node")
	ErrUnknownEndpoint = errors.New("unknown query endpoint")
)

// FactSelection chooses which stored fact layer a read may inspect. The zero
// value is deliberately inclusive for compatibility with the original API.
type FactSelection uint8

const (
	SelectAll FactSelection = iota
	SelectDeterministic
	SelectCandidate
)

func (selection FactSelection) normalized() (FactSelection, error) {
	if selection > SelectCandidate {
		return 0, fmt.Errorf("%w: unknown fact selection %d", ErrInvalidQuery, selection)
	}
	return selection, nil
}

func (selection FactSelection) includes(status FactStatus) bool {
	switch selection {
	case SelectDeterministic:
		return status == FactDeterministic
	case SelectCandidate:
		return status == FactCandidate
	default:
		return status == FactDeterministic || status == FactCandidate
	}
}

// MatchOptions controls the fact layer visible to an exact read.
type MatchOptions struct {
	Selection FactSelection
}

func (options MatchOptions) normalized() (MatchOptions, error) {
	selection, err := options.Selection.normalized()
	if err != nil {
		return MatchOptions{}, err
	}
	options.Selection = selection
	return options, nil
}
