package lsp

import (
	"errors"
	"sort"
	"unicode/utf8"
)

var (
	ErrInvalidPosition = errors.New("lsp: invalid UTF-16 position")
	ErrInvalidRange    = errors.New("lsp: invalid document range")
)

// PositionToOffset converts an LSP UTF-16 position to a UTF-8 byte offset.
// It rejects negative coordinates, astral-code-point splits, and line overflow.
func PositionToOffset(source string, position Position) (int, error) {
	if position.Line < 0 || position.Character < 0 {
		return 0, ErrInvalidPosition
	}
	starts := sourceLineStarts(source)
	if position.Line >= len(starts) {
		return 0, ErrInvalidPosition
	}
	start, end := starts[position.Line], sourceLineEnd(source, starts[position.Line])
	offset, units := start, 0
	for offset < end {
		if units == position.Character {
			return offset, nil
		}
		runeValue, size := utf8.DecodeRuneInString(source[offset:end])
		width := utf16Width(runeValue)
		if units+width > position.Character {
			return 0, ErrInvalidPosition
		}
		units += width
		offset += size
	}
	if units == position.Character {
		return end, nil
	}
	return 0, ErrInvalidPosition
}

// OffsetToPosition converts a UTF-8 byte offset to a zero-based LSP position.
// Invalid UTF-8 bytes are deterministic one-unit replacement boundaries.
func OffsetToPosition(source string, offset int) (Position, error) {
	if err := validateByteOffset(source, offset); err != nil {
		return Position{}, err
	}
	starts := sourceLineStarts(source)
	line := sort.Search(len(starts), func(index int) bool { return starts[index] > offset }) - 1
	if line < 0 {
		line = 0
	}
	start := starts[line]
	return Position{Line: line, Character: utf16Length(source[start:offset])}, nil
}
func ValidateRange(source string, value Range) (int, int, error) {
	start, err := PositionToOffset(source, value.Start)
	if err != nil {
		return 0, 0, ErrInvalidRange
	}
	end, err := PositionToOffset(source, value.End)
	if err != nil || start > end {
		return 0, 0, ErrInvalidRange
	}
	return start, end, nil
}
