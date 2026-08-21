package cache

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSemanticProjectionKeyRejectsMismatchedIdentity(t *testing.T) {
	ir := loadBillingIR(t, func(source string) string { return source })
	spec := projectionSpec()
	spec.SemanticClosureDigest = HashBytes([]byte("source-presentation"))
	if _, err := NewSemanticProjectionKey(ir, spec); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("mismatched semantic digest = %v, want ErrInvalidKey", err)
	}
	spec = projectionSpec()
	spec.Domain = "payments"
	if _, err := NewSemanticProjectionKey(ir, spec); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("mismatched semantic namespace = %v, want ErrInvalidKey", err)
	}
}
func TestSemanticProjectionKeyUsesNormalizedNamespace(t *testing.T) {
	ir := loadBillingIR(t, func(source string) string {
		return strings.Replace(source, "namespace billing", "namespace   billing ", 1)
	})
	spec := projectionSpec()
	spec.Domain = ""
	spec.Namespace = ""
	spec.SemanticClosureDigest = ""
	key, err := NewSemanticProjectionKey(ir, spec)
	if err != nil {
		t.Fatal(err)
	}
	if key.Domain != "billing" || key.Namespace != "billing" {
		t.Fatalf("normalized namespace leaked presentation whitespace: %+v", key)
	}
}
func TestSemanticDigestRejectsInvalidIR(t *testing.T) {
	invalid := semantic.NewIR("billing", "billing")
	invalid.Namespace = "invalid namespace"
	if _, err := SemanticDigest(invalid); !errors.Is(err, semantic.ErrInvalidNamespace) {
		t.Fatalf("invalid semantic IR = %v, want ErrInvalidNamespace", err)
	}
}
func loadBillingIR(t *testing.T, mutate func(string) string) semantic.IR {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	path := filepath.Join(root, "examples", "billing", "main.gooo")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read billing fixture: %v", err)
	}
	file, diagnostics := syntax.ParseFile(path, mutate(string(raw)))
	if err := diagnostics.Error(); err != nil {
		t.Fatalf("billing fixture diagnostics: %v", err)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatalf("lower billing fixture: %v", err)
	}
	return ir
}
