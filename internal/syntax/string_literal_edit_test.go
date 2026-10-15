package syntax

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

func TestLiteralEditPreservesSourceAndEscapesReplacement(t *testing.T) {
	for _, literal := range []string{`"old"`, "`old`"} {
		source := []byte("before " + literal + " after")
		span := Span{}
		span.Start.Offset, span.End.Offset = 7, 7+len(literal)
		replacement := "\"\nactivity Inject(Integer) -> Integer\\\x00"
		got, err := ReplaceStringLiteral(source, span, replacement)
		if err != nil || string(got) != "before "+strconv.Quote(replacement)+" after" {
			t.Fatalf("literal edit = %q, %v", got, err)
		}
		if string(source) != "before "+literal+" after" {
			t.Fatal("source was mutated")
		}
	}
}

func TestLiteralEditRejectsInvalidRangesAndTokens(t *testing.T) {
	for _, bounds := range [][2]int{{-1, 3}, {0, 0}, {2, 1}, {0, 4}, {1, 2}} {
		span := Span{}
		span.Start.Offset, span.End.Offset = bounds[0], bounds[1]
		if _, err := ReplaceStringLiteral([]byte(`"x"`), span, "new"); err == nil {
			t.Fatalf("accepted invalid literal span %v", bounds)
		}
	}
}

func TestLiteralReplacementSizeChecksBeforeAllocation(t *testing.T) {
	for _, sizes := range [][3]int{{math.MaxInt, 0, 1}, {-1, 0, 0}, {1, 2, 0}, {1, -1, 0}, {1, 0, -1}} {
		if _, err := replacementSourceSize(sizes[0], sizes[1], sizes[2]); err == nil {
			t.Fatalf("accepted invalid size %v", sizes)
		}
	}
	for _, sizes := range [][4]int{{math.MaxInt, 1, 1, math.MaxInt}, {7, 3, 5, 9}, {3, 3, 0, 0}} {
		got, err := replacementSourceSize(sizes[0], sizes[1], sizes[2])
		if err != nil || got != sizes[3] {
			t.Fatalf("size %v = %d, %v", sizes, got, err)
		}
	}
}

func TestAssemblyIntegerBoundsRetainOnlyValidValues(t *testing.T) {
	source := `package bounds
namespace bounds
entity Integer id "gooo://bounds/integer"
activity Adjust(Integer) -> Integer computes "return input" assembling {
 choice "offset" operand_order at "OCCURRENCE" intent "Keep the input"
 case "17" -> "17"
 attempts "ATTEMPTS"
}`
	for _, value := range []string{"-9223372036854775808", "-1", "9223372036854775807"} {
		text := strings.NewReplacer("OCCURRENCE", value, "ATTEMPTS", value).Replace(source)
		file, diagnostics := ParseFile("bounds.gooo", text)
		if file == nil || !diagnostics.HasErrors() {
			t.Fatalf("invalid integers accepted: %s", value)
		}
		assembly := file.Declarations[1].(*ActivityDecl).Assembly.Spec
		if assembly.MaxAttempts != 0 || assembly.Choices[0].Occurrence != 0 {
			t.Fatalf("invalid integers entered the AST: %+v", assembly)
		}
	}
	for _, pair := range [][2]string{{"0", "1"}, {"127", "64"}} {
		text := strings.NewReplacer("OCCURRENCE", pair[0], "ATTEMPTS", pair[1]).Replace(source)
		if _, diagnostics := ParseFile("bounds.gooo", text); diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
	}
}
