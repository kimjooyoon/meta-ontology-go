package lsp

import (
	"errors"
	"sort"
	"unicode/utf16"
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

func sourceLineStarts(source string) []int {
	starts := []int{0}
	for index := 0; index < len(source); {
		switch source[index] {
		case '\r':
			index++
			if index < len(source) && source[index] == '\n' {
				index++
			}
			starts = append(starts, index)
		case '\n':
			index++
			starts = append(starts, index)
		default:
			_, size := utf8.DecodeRuneInString(source[index:])
			if size == 0 {
				size = 1
			}
			index += size
		}
	}
	return starts
}

func sourceLineEnd(source string, start int) int {
	for index := start; index < len(source); index++ {
		if source[index] == '\r' || source[index] == '\n' {
			return index
		}
	}
	return len(source)
}

func validateByteOffset(source string, offset int) error {
	if offset < 0 || offset > len(source) {
		return ErrInvalidPosition
	}
	for index := 0; index < offset; {
		_, size := utf8.DecodeRuneInString(source[index:])
		if size == 0 {
			size = 1
		}
		if index+size > offset {
			return ErrInvalidPosition
		}
		index += size
	}
	if offset > 0 && offset < len(source) && source[offset-1] == '\r' && source[offset] == '\n' {
		return ErrInvalidPosition
	}
	return nil
}

func utf16Width(value rune) int {
	width := utf16.RuneLen(value)
	if width < 0 {
		return 1
	}
	return width
}

func utf16Length(source string) int {
	units := 0
	for index := 0; index < len(source); {
		value, size := utf8.DecodeRuneInString(source[index:])
		if size == 0 {
			size = 1
		}
		units += utf16Width(value)
		index += size
	}
	return units
}
