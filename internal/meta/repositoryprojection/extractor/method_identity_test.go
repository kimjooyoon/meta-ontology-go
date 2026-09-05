package extractor

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMethodIdentityIgnoresReceiverSpelling(t *testing.T) {
	first := methodIdentityFromSource(t, `package p
type T struct{}
func (original *T) M() {}
`)
	second := methodIdentityFromSource(t, `package p
type T struct{}
func ( /* receiver */ renamed *T ) M() {}
`)
	if first != "method:T:M" || second != first {
		t.Fatalf("receiver spelling changed method identity: first=%q second=%q", first, second)
	}
}

func TestMethodIdentityIgnoresGenericReceiverBinders(t *testing.T) {
	cases := []struct {
		name   string
		first  string
		second string
		want   string
	}{
		{
			name: "single binder",
			first: `package p
type Single[A any] struct{}
func (receiver Single[First]) M() {}
`,
			second: `package p
type Single[A any] struct{}
func (renamed Single[Second]) M() {}
`,
			want: "method:Single:M",
		},
		{
			name: "multiple binders",
			first: `package p
type Pair[A, B any] struct{}
func (receiver Pair[First, Second]) M() {}
`,
			second: `package p
type Pair[A, B any] struct{}
func (renamed Pair[Left, Right]) M() {}
`,
			want: "method:Pair:M",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			first := methodIdentityFromSource(t, testCase.first)
			second := methodIdentityFromSource(t, testCase.second)
			if first != testCase.want || second != first {
				t.Fatalf("generic receiver binders changed method identity: first=%q second=%q want=%q", first, second, testCase.want)
			}
		})
	}
}

func TestMethodIdentityDifferentBasesDoNotCollide(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", `package p
type A struct{}
type B struct{}
func (a A) M() {}
func (b B) M() {}
`, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	declarations, fallbackUsed, err := candidates(fset, file)
	if err != nil || fallbackUsed {
		t.Fatalf("different receiver bases = declarations=%#v fallback=%t err=%v", declarations, fallbackUsed, err)
	}
	want := map[string]bool{"method:A:M": false, "method:B:M": false}
	for _, declaration := range declarations {
		if _, ok := want[declaration.identity]; ok {
			want[declaration.identity] = true
		}
	}
	for identity, found := range want {
		if !found {
			t.Fatalf("missing distinct method identity %q in %#v", identity, declarations)
		}
	}
}

func TestMethodIdentityPointerValueCollisionUsesExtractPath(t *testing.T) {
	for _, source := range []string{
		`package p
type T struct{}
func (value T) M() {}
func (pointer *T) M() {}
`,
		`package p
type T struct{}
func (renamed T) M() {}
// spacing and comments must not create a second declaration identity.
func (other *T) M() {}
`,
	} {
		t.Run(strings.Split(source, "\n")[2], func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			_, _, err := Extract(root, "x.go")
			var failure Failure
			if !errors.As(err, &failure) || failure.Reason != "DECLARATION_IDENTITY_COLLISION" || len(failure.Diagnostics) != 1 || failure.Diagnostics[0] != "method:T:M" {
				t.Fatalf("Extract collision = %v, want stable method:T:M collision", err)
			}
		})
	}
}

func TestMethodIdentityMalformedReceiverFailsClosed(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(`package p
func (receiver pkg.T) M() {}
`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := Extract(root, "x.go")
	var failure Failure
	if !errors.As(err, &failure) || failure.Reason != "UNSUPPORTED_RECEIVER" || len(failure.Diagnostics) != 1 || strings.Contains(failure.Diagnostics[0], "func-at:") {
		t.Fatalf("malformed receiver = %v, want fail-closed unsupported receiver", err)
	}
}

func methodIdentityFromSource(t *testing.T, source string) string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv == nil {
			continue
		}
		identity, movable := identityOf(fset, function)
		if !movable {
			t.Fatalf("method receiver was not supported: %#v", function.Recv)
		}
		return identity
	}
	t.Fatal("source has no method declaration")
	return ""
}
