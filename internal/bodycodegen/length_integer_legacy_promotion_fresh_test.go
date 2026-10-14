package bodycodegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestLengthIntegerPreservesFreshValidShapes20537(t *testing.T) {
	for _, text := range []string{
		`package p;func Legacy(input string)int64{var n=len(input);var m=n+20537;return int64(m)}`,
		`package p;func Compare(input string)bool{var n=len(input);return n<20537}`,
		`package p;func len(input string)int64{return 20537};func Shadow(input string)int64{return len(input)}`,
	} {
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, "newvalid20537.go", text, parser.AllErrors)
		if err != nil {
			t.Fatal(err)
		}
		before, _ := formatNode(set, file)
		if units := normalizeLengthIntegerContexts("p", file, set); units != 0 {
			t.Fatal("well-typed fresh source gained projection units", units)
		}
		after, _ := formatNode(set, file)
		if before != after {
			t.Fatal("well-typed length or source-declared len projection changed", before, after)
		}
	}
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "foreign20537.go", `package p;func Foreign(input int64,foreign int)int64{return input+foreign+20537}`, parser.AllErrors)
	if err != nil {
		t.Fatal(err)
	}
	if normalizeLengthIntegerContexts("p", file, set) != 0 {
		t.Fatal("arbitrary int was treated as a runtime length")
	}
	if err := typecheck("p", file, set); err == nil {
		t.Fatal("unrelated mixed integer types accepted")
	}
	t.Log("three fresh well-typed ASTs byte-form unchanged;source len helper retained;unrelated int rejected;historical cohort executions0")
}

func TestLengthIntegerExplicitProjectionRemains20537(t *testing.T) {
	source := []byte(lengthContextPrelude20537 + "activity Run(Text)->Integer computes `let n=len(input);let m=n+20537;return int64(m)`\n")
	result, err := Generate("explicit20537.gooo", source, "Run")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result.Source, "var n int64") || strings.Contains(result.Source, "var m int64") ||
		!strings.Contains(result.Source, "return int64(m)") {
		t.Fatal("established explicit conversion and length locals changed", result.Source)
	}
}
