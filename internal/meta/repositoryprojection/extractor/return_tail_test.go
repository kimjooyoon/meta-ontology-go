package extractor

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReturnTailSafetyMatrix(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		positive bool
	}{
		{name: "positive terminal error tail", source: returnTailFixture("func F(values map[string]struct{}) error {\n", "\tif len(values) != 0 {\n\t\treturn nil\n\t}\n\treturn nil\n"), positive: true},
		{name: "positive early return is preserved", source: returnTailFixture("func F(values map[string]struct{}) error {\n", "\tif len(values) == 0 {\n\t\treturn errorSentinel()\n\t}\n\tif len(values) != 0 {\n\t\treturn nil\n\t}\n\treturn nil\n"), positive: true},
		{name: "named result", source: returnTailFixture("func F(values map[string]struct{}) (err error) {\n", "\tif len(values) != 0 {\n\t\treturn err\n\t}\n\treturn err\n"), positive: false},
		{name: "method", source: returnTailFixture("func (T) F(values map[string]struct{}) error {\n", "\tif len(values) != 0 {\n\t\treturn nil\n\t}\n\treturn nil\n"), positive: false},
		{name: "go statement", source: returnTailFixture("func F(values map[string]struct{}) error {\n", "\tgo func() {}()\n\treturn nil\n"), positive: false},
		{name: "defer statement", source: returnTailFixture("func F(values map[string]struct{}) error {\n", "\tdefer func() {}()\n\treturn nil\n"), positive: false},
		{name: "escaping branch", source: returnTailFixture("func F(values map[string]struct{}) error {\n", "\tgoto done\n\tdone:\n\treturn nil\n"), positive: false},
		{name: "address escape stale pointer", source: returnTailPrefixBindingFixture("func F(values map[string]struct{}) error {\n", "\terr := error(nil)\n\tp := &err\n\t_ = p\n", "\t*p = errorSentinel()\n\treturn err\n"), positive: false},
		{name: "closure capture stale copy", source: returnTailPrefixBindingFixture("func F(values map[string]struct{}) error {\n", "\terr := error(nil)\n\tset := func() { err = errorSentinel() }\n\t_ = set\n", "\tset()\n\treturn err\n"), positive: false},
		{name: "false helper capacity proof", source: returnTailFixture("func F(values map[string]struct{}) error {\n", "\tif len(values) != 0 {\n"+strings.Repeat("\t\t_ = 1\n", 70)+"\t\treturn nil\n\t}\n\treturn nil\n"), positive: false},
	}
	if len(cases) != 10 {
		t.Fatalf("safety matrix denominator=%d, want 10", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(tc.source), 0o644); err != nil {
				t.Fatal(err)
			}
			result, err := ExtractWithResult(root, "x.go")
			if tc.positive {
				if err != nil {
					t.Fatalf("positive case failed: %v", err)
				}
				if len(result.Evidence) != 1 || result.Evidence[0].Strategy != returnTailStrategy {
					t.Fatalf("strategy evidence=%+v", result.Evidence)
				}
				if len(result.Evidence[0].Obligations) != len(returnTailObligations) {
					t.Fatalf("obligations=%+v", result.Evidence[0].Obligations)
				}
				if result.Evidence[0].BeforeFunctionLines <= functionLineLimit || result.Evidence[0].AfterFunctionLines > functionLineLimit ||
					result.Evidence[0].RenderedHelperLines > functionLineLimit || result.Evidence[0].RenderedOuterHelperLines > functionLineLimit {
					t.Fatalf("capacity evidence=%+v", result.Evidence[0])
				}
				for path, data := range result.Generated {
					if extractionLines(data) > functionLineLimit {
						t.Fatalf("generated unit %s exceeds capacity: %d lines", path, extractionLines(data))
					}
				}
				if !strings.Contains(string(result.Generated["x.go"]), "return FExtractedReturnTail") {
					t.Fatal("outer function did not use a return-valued helper")
				}
				if tc.name == "positive early return is preserved" && !strings.Contains(string(result.Generated["x.go"]), "return errorSentinel()") {
					t.Fatal("outer early return was not preserved")
				}
				return
			}
			var failure Failure
			if !errors.As(err, &failure) {
				t.Fatalf("negative case error=%v", err)
			}
			if failure.Reason != "NO_SAFE_DECLARATION_CAPACITY" && failure.Reason != "METHOD_SUFFIX_DECOMPOSITION_UNSAFE" {
				t.Fatalf("negative case reason=%s error=%v", failure.Reason, err)
			}
		})
	}
}

func returnTailFixture(header, tail string) string {
	return "package p\n\n" + header + strings.Repeat("\t_ = 1\n", 72) + tail + "}\n\nfunc errorSentinel() error { return &sentinelError{} }\n\ntype sentinelError struct{ error }\n\ntype T struct{}\n"
}

func returnTailPrefixBindingFixture(header, prefix, tail string) string {
	return "package p\n\n" + header + prefix + strings.Repeat("\t_ = 1\n", 72) + tail + "}\n\nfunc errorSentinel() error { return &sentinelError{} }\n\ntype sentinelError struct{ error }\n\ntype T struct{}\n"
}
