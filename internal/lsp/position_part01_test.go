package lsp

import (
	"testing"
)

func TestUTF16ConversionHandlesAstralAndEOF(t *testing.T) {
	source := "a😀b"
	if got, err := OffsetToPosition(source, len(source)); err != nil || got != (Position{Character: 4}) {
		t.Fatalf("EOF position = %#v, error = %v", got, err)
	}
	if got, err := PositionToOffset(source, Position{Character: 3}); err != nil || got != len("a😀") {
		t.Fatalf("after astral offset = %d, error = %v", got, err)
	}
	if _, err := PositionToOffset(source, Position{Character: 2}); err == nil {
		t.Fatal("accepted a position inside an astral code point")
	}
	if _, err := OffsetToPosition(source, 2); err == nil {
		t.Fatal("accepted a byte offset inside an astral code point")
	}
}
func TestUTF16ConversionHandlesCRLFBoundaries(t *testing.T) {
	source := "a\r\nb\r\n"
	checks := []struct {
		position Position
		offset   int
	}{
		{Position{Line: 0, Character: 1}, 1},
		{Position{Line: 1, Character: 0}, 3},
		{Position{Line: 1, Character: 1}, 4},
		{Position{Line: 2, Character: 0}, len(source)},
	}
	for _, check := range checks {
		if got, err := PositionToOffset(source, check.position); err != nil || got != check.offset {
			t.Fatalf("position %#v = %d, error = %v, want %d", check.position, got, err, check.offset)
		}
	}
	if _, err := PositionToOffset(source, Position{Character: 2}); err == nil {
		t.Fatal("accepted a position after CR and before LF")
	}
	if _, err := OffsetToPosition(source, 2); err == nil {
		t.Fatal("accepted an offset between CR and LF")
	}
}
func TestUTF16ConversionTreatsInvalidBytesAsBoundaries(t *testing.T) {
	source := string([]byte{'a', 0xff, 'b'})
	for offset, want := range map[int]Position{0: {}, 1: {Character: 1}, 2: {Character: 2}, 3: {Character: 3}} {
		got, err := OffsetToPosition(source, offset)
		if err != nil || got != want {
			t.Fatalf("offset %d = %#v, error = %v, want %#v", offset, got, err, want)
		}
	}
	if got, err := PositionToOffset(source, Position{Character: 2}); err != nil || got != 2 {
		t.Fatalf("invalid-byte position = %d, error = %v", got, err)
	}
}
func TestRangeValidationRejectsReversedAndOverflowRanges(t *testing.T) {
	if _, _, err := ValidateRange("abc", Range{Start: Position{Character: 2}, End: Position{Character: 1}}); err == nil {
		t.Fatal("accepted a reversed range")
	}
	if _, _, err := ValidateRange("abc", Range{Start: Position{Character: 0}, End: Position{Character: 4}}); err == nil {
		t.Fatal("accepted a range beyond the line")
	}
	if _, _, err := ValidateRange("abc", Range{Start: Position{Line: 1}, End: Position{Line: 1}}); err == nil {
		t.Fatal("accepted a nonexistent line")
	}
}
