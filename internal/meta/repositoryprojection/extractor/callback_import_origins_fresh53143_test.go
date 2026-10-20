package extractor

import (
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type callbackImportProbe53143 struct {
	base  types.Importer
	calls int
}

func (probe *callbackImportProbe53143) Import(path string) (*types.Package, error) {
	probe.calls++
	return probe.base.Import(path)
}

func callbackOriginFixture53143(t *testing.T) (string, []byte) {
	t.Helper()
	root := t.TempDir()
	source := []byte(`package fixture
import h "net/http"
import "testing"
func TestPaginationFixturesExecuteParserAndHTTPClient(t *testing.T) {
value := 73
callback := h.HandlerFunc(func(_ h.ResponseWriter, _ *h.Request) { value += 2 })
callback(nil, nil)
if value != 75 { t.Fatal(value) }
}
`)
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module callback.origin53143\n\ngo 1.27.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "fixture_test.go"), source, 0o644); err != nil {
		t.Fatal(err)
	}
	return root, source
}

func TestCallbackImportOrigins53143(t *testing.T) {
	root, source := callbackOriginFixture53143(t)
	_, set, file, err := readParsedSource(root, "fixture_test.go")
	if err != nil {
		t.Fatal(err)
	}
	target := callbackPreviewFunction(file, callbackPreviewTarget)
	evidence, err := checkTypes(root, "fixture_test.go", set, file, target)
	if err != nil {
		t.Fatal(err)
	}
	callback, err := callbackPreviewFuncLit(target, evidence.info)
	if err != nil {
		t.Fatal(err)
	}
	captures, err := callbackPreviewCaptures(callback, evidence, set)
	if err != nil {
		t.Fatal(err)
	}
	probe := &callbackImportProbe53143{base: evidence.standard.fallback}
	evidence.standard.fallback = probe
	candidate, err := callbackPreviewFactoryCandidate(root, source, "fixture_test.go", set, file, target, callback, evidence, captures, callbackPreviewEffects(callback, evidence, set))
	if err != nil || probe.calls == 0 {
		t.Fatal("candidate lost original import origin", err, probe.calls)
	}
	prior := probe.calls
	proof, err := proveClosurePreservation(root, "fixture_test.go", source, set, file, target, callback, evidence, captures, candidate)
	if err != nil || probe.calls <= prior {
		t.Fatal("preservation lost operation import origin", err, probe.calls)
	}
	if proof.SourceBodyDigest != proof.CandidateBodyDigest || proof.SourceContextDigest != proof.OutputContextDigest || proof.SemanticAdmission != "UNKNOWN" {
		t.Fatal("source/capture relation changed", proof)
	}
	changed := strings.Replace(string(source), "value := 73", "value := 83", 1)
	if err := os.WriteFile(filepath.Join(root, "fixture_test.go"), []byte(changed), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyCallbackFactoryPreservation(root, "fixture_test.go", candidate); err == nil {
		t.Fatal("independent reload accepted stale source")
	}
	t.Log("new73/75 callback:original/candidate/preservation share operation import origin;fresh source reload rejects83;models0")
}

func TestCallbackImportScopeChange53143(t *testing.T) {
	root, source := callbackOriginFixture53143(t)
	_, set, file, err := readParsedSource(root, "fixture_test.go")
	if err != nil {
		t.Fatal(err)
	}
	target := callbackPreviewFunction(file, callbackPreviewTarget)
	evidence, err := checkTypes(root, "fixture_test.go", set, file, target)
	if err != nil {
		t.Fatal(err)
	}
	callback, err := callbackPreviewFuncLit(target, evidence.info)
	if err != nil {
		t.Fatal(err)
	}
	captures, err := callbackPreviewCaptures(callback, evidence, set)
	if err != nil {
		t.Fatal(err)
	}
	probe := &callbackImportProbe53143{base: evidence.standard.fallback}
	evidence.standard.fallback = probe
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module callback.changed53143\n\ngo 1.27.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := callbackPreviewFactoryCandidate(root, source, "fixture_test.go", set, file, target, callback, evidence, captures, callbackPreviewEffects(callback, evidence, set)); err != nil {
		t.Fatal(err)
	}
	if evidence.standard.fallback == probe || probe.calls != 0 {
		t.Fatal("changed module kept previous origin reader", probe.calls)
	}
	t.Log("new module change resets standard origin reader;no global or project AST reuse")
}
