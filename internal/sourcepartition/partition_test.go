package sourcepartition

import (
	"strings"
	"testing"
)

func TestCutRequiresAnExactDeclarationBoundary(t *testing.T) {
	const source = "// 기원\r\npackage p\r\nnamespace p\r\nentity A id \"urn:a\"\r\nentity B id \"urn:b\"\r\n"
	cut := strings.Index(source, "entity B")
	files, err := SplitAt("a.gooo", source, "b.gooo", cut)
	if err != nil || len(files) != 2 {
		t.Fatal(files, err)
	}
	for _, file := range files {
		end := 0
		for _, origin := range file.Origins {
			if origin.OutputStart != end || source[origin.InputStart:origin.InputEnd] != file.Source[origin.OutputStart:origin.OutputEnd] {
				t.Fatal("partition lost exact original bytes")
			}
			end = origin.OutputEnd
		}
		if end != len(file.Source) {
			t.Fatal("partition has output without an origin")
		}
	}
	for _, offset := range []int{-1, 0, strings.Index(source, "entity A"), cut - 1, cut + 1, len(source)} {
		if files, err := SplitAt("a.gooo", source, "b.gooo", offset); err == nil || files != nil {
			t.Fatal("non-boundary cut accepted", offset)
		}
	}
}
