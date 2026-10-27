package sourcepartition

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestBoundarySizesMatchExactOutputBytes(t *testing.T) {
	const original = "// 기원\npackage p\nnamespace p\nentity A id \"urn:a\"\n\n// retain\n  entity B id \"urn:b\"\nentity C id \"urn:c\"\n"
	const inline = "package p namespace p entity A id \"urn:a\" entity B id \"urn:b\" entity C id \"urn:c\""
	for _, source := range []string{original, strings.ReplaceAll(original, "\n", "\r\n"),
		strings.TrimSuffix(original, "\n"), inline} {
		layout, err := InspectBoundaries("a.gooo", source)
		if err != nil || len(layout.Boundaries) != 2 {
			t.Fatal("missing declaration boundaries", err)
		}
		for _, boundary := range layout.Boundaries {
			files, err := SplitAt("a.gooo", source, "b.gooo", boundary.Declaration.SourceSpan().Start.Offset)
			if err != nil {
				t.Fatal(err)
			}
			for i, expected := range []Size{boundary.Left, boundary.Right} {
				scanner, lines := bufio.NewScanner(strings.NewReader(files[i].Source)), 0
				for scanner.Scan() {
					lines++
				}
				if err := scanner.Err(); err != nil || expected != (Size{len(files[i].Source), lines}) {
					t.Fatal("predicted size differs from actual partition", expected, lines, err)
				}
			}
			if boundary.Left.Bytes+boundary.Right.Bytes != len(source)+layout.HeaderBytes {
				t.Fatal("header duplication is missing from the size")
			}
		}
	}
}

func TestBoundaryInspectionDoesNotInventCutsOrReturnPrefixes(t *testing.T) {
	const single = "package p\nnamespace p\nentity A id \"urn:a\"\n"
	layout, err := InspectBoundaries("a.gooo", single)
	if err != nil || len(layout.Boundaries) != 0 {
		t.Fatal("single declaration has no usable cut", layout, err)
	}
	for _, source := range []string{"", single + "broken", single + "import q \"q\"\n"} {
		layout, err := InspectBoundaries("a.gooo", source)
		if err == nil || !reflect.DeepEqual(layout, Layout{}) {
			t.Fatal("invalid input returned a candidate prefix", layout, err)
		}
	}
}
