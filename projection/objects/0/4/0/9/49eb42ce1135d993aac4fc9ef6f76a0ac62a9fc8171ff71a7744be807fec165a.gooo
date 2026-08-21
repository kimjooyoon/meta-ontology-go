package coupling

import (
	"errors"
	"sort"
	"unicode/utf8"
)

var (
	ErrInvalidLivePosition = errors.New("coupling: invalid UTF-16 position")
	ErrInvalidLiveRange    = errors.New("coupling: invalid document range")
)

// PositionToOffset converts an LSP UTF-16 position to a UTF-8 byte offset.
// It rejects astral-code-point splits, line overflow, invalid boundaries, and
// the interior of CRLF line endings.
func PositionToOffset(source string, position Position) (int, error) {
	if position.Line < 0 || position.Character < 0 {
		return 0, ErrInvalidLivePosition
	}
	starts := liveLineStarts(source)
	if position.Line >= len(starts) {
		return 0, ErrInvalidLivePosition
	}
	start, end := starts[position.Line], liveLineEnd(source, starts[position.Line])
	offset, units := start, 0
	for offset < end {
		if units == position.Character {
			return offset, nil
		}
		runeValue, size := utf8.DecodeRuneInString(source[offset:end])
		width := liveUTF16Width(runeValue)
		if units+width > position.Character {
			return 0, ErrInvalidLivePosition
		}
		units += width
		offset += size
	}
	if units == position.Character {
		return end, nil
	}
	return 0, ErrInvalidLivePosition
}

// OffsetToPosition converts a UTF-8 byte offset to a zero-based LSP UTF-16
// position.
func OffsetToPosition(source string, offset int) (Position, error) {
	if err := validateLiveOffset(source, offset); err != nil {
		return Position{}, err
	}
	starts := liveLineStarts(source)
	line := max(sort.Search(len(starts), func(index int) bool { return starts[index] > offset })-1, 0)
	return Position{Line: line, Character: liveUTF16Length(source[starts[line]:offset])}, nil
}

// ValidateRange validates an LSP UTF-16 range and returns UTF-8 byte offsets.
func ValidateRange(source string, value Range) (int, int, error) {
	start, err := PositionToOffset(source, value.Start)
	if err != nil {
		return 0, 0, ErrInvalidLiveRange
	}
	end, err := PositionToOffset(source, value.End)
	if err != nil || start > end {
		return 0, 0, ErrInvalidLiveRange
	}
	return start, end, nil
}
