package extractor

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

const callbackPreviewLogicalPath = "cmd/language-readiness-witness/predecessor-selection/pagination_test.go"

func TestPaginationCallbackPreviewCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("pagination callback preview is CI-only")
	}
	root := callbackPreviewRepositoryRoot(t)
	preview, err := PreviewBoundedPaginationCallback(root, callbackPreviewLogicalPath)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Schema != callbackPreviewSchema || preview.State != callbackPreviewStateUnknown || preview.Reason != "PENDING_TYPED_CALLBACK_EFFECTS" || preview.Candidate == nil {
		t.Fatalf("pagination callback preview = %#v", preview)
	}
	if preview.OperationResultAdmission != "FORBIDDEN" || preview.ApplyPermission != "FORBIDDEN" || preview.Candidate.Promotion != callbackPreviewPromotionNone {
		t.Fatalf("pending preview was admitted: %#v", preview)
	}
	if preview.Stage != "CALLBACK_PREVIEW" || preview.Step != "PENDING_EFFECT_REVIEW" || preview.UnknownClass != "PENDING_TYPED_CALLBACK_EFFECTS" || preview.NextOperation != "RESOLVE_TYPED_CALLBACK_EFFECTS" || preview.BlockedBy != "PENDING_TYPED_CALLBACK_EFFECTS" {
		t.Fatalf("preview lifecycle = %#v", preview)
	}
	if len(preview.ContractRecords) != 5 || preview.ContractRecords[0].Entity != "CallbackPreviewInput" || preview.ContractRecords[1].Entity != "BoundedCallbackCandidate" || preview.ContractRecords[2].Entity != "CallbackCaptures" || preview.ContractRecords[3].Entity != "PendingCallbackEffects" || preview.ContractRecords[4].Entity != "CallbackPreviewEvidence" {
		t.Fatalf("contract record flow = %#v", preview.ContractRecords)
	}
	if preview.Candidate.HelperLines <= 0 || preview.Candidate.HelperLines > functionLineLimit || preview.Candidate.ParentFunctionLines <= 0 || preview.Candidate.ParentFunctionLines > functionLineLimit {
		t.Fatalf("candidate capacity = helper:%d parent:%d", preview.Candidate.HelperLines, preview.Candidate.ParentFunctionLines)
	}
	if len(preview.Captures) != 2 || !callbackPreviewHasCapture(preview.Captures, "fixture", "paginationFixture", "pointer-identity") || !callbackPreviewHasCapture(preview.Captures, "requestCount", "atomic.Int32", "pointer-identity") {
		t.Fatalf("capture bindings = %#v", preview.Captures)
	}
	if preview.Candidate.CaptureCount != len(preview.Captures) || preview.Candidate.PendingEffectCount != len(preview.PendingEffects) || len(preview.PendingEffects) == 0 {
		t.Fatalf("preview counters = candidate:%#v effects:%d", preview.Candidate, len(preview.PendingEffects))
	}
	if !callbackPreviewHasEffectKind(preview.PendingEffects, "dynamic-interface-method") || !callbackPreviewHasEffectKind(preview.PendingEffects, "dynamic-interface-argument") || !callbackPreviewHasEffectKind(preview.PendingEffects, "typed-method") || !callbackPreviewHasEffectKind(preview.PendingEffects, "external-function") {
		t.Fatalf("typed pending effects = %#v", preview.PendingEffects)
	}
	if !strings.Contains(preview.Candidate.HelperSource, "*paginationFixture") || !strings.Contains(preview.Candidate.HelperSource, "*atomic.Int32") || !strings.Contains(preview.Candidate.WrapperSource, "&fixture") || !strings.Contains(preview.Candidate.WrapperSource, "&requestCount") {
		t.Fatalf("identity-preserving candidate binding missing: wrapper=%q helper=%q", preview.Candidate.WrapperSource, preview.Candidate.HelperSource)
	}
	typeCheckCallbackPreviewCandidate(t, root, callbackPreviewLogicalPath, preview.Candidate.CandidateSource)
	runCallbackPreviewFixtureSuite(t, root, callbackPreviewLogicalPath, "")
	runCallbackPreviewFixtureSuite(t, root, callbackPreviewLogicalPath, preview.Candidate.CandidateSource)
}

func TestPaginationCallbackPreviewMissingAndMalformedAreNotAdmission(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd/language-readiness-witness/predecessor-selection"), 0o755); err != nil {
		t.Fatal(err)
	}
	logical := callbackPreviewLogicalPath
	path := filepath.Join(root, filepath.FromSlash(logical))
	if err := os.WriteFile(path, []byte("package main\n\nfunc Other() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing, err := PreviewBoundedPaginationCallback(root, logical)
	if err != nil {
		t.Fatal(err)
	}
	if missing.Reason != "CALLBACK_TARGET_MISSING" || missing.Candidate != nil || missing.ApplyPermission != "FORBIDDEN" || missing.OperationResultAdmission != "FORBIDDEN" {
		t.Fatalf("missing callback preview = %#v", missing)
	}
	if err := os.WriteFile(path, []byte("package main\nfunc"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := PreviewBoundedPaginationCallback(root, logical); err == nil {
		t.Fatal("malformed callback source was accepted")
	}
}

func callbackPreviewRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "../../../.."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repository root unavailable: %v", err)
	}
	return root
}

func callbackPreviewHasCapture(captures []CallbackPreviewCapture, name, objectType, mode string) bool {
	for _, capture := range captures {
		if capture.Name == name && capture.ObjectType == objectType && capture.BindingMode == mode && capture.ObjectIdentity != "" {
			return true
		}
	}
	return false
}

func callbackPreviewHasEffectKind(effects []CallbackPreviewEffect, kind string) bool {
	for _, effect := range effects {
		if effect.EffectKind == kind && effect.CallIdentity != "" && effect.Signature != "" {
			return true
		}
	}
	return false
}

func typeCheckCallbackPreviewCandidate(t *testing.T, root, logical, source string) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, logical, []byte(source), parser.ParseComments)
	if err != nil {
		t.Fatalf("candidate parse failed: %v", err)
	}
	files, err := packageTypeFiles(root, logical, fset, file)
	if err != nil {
		t.Fatalf("candidate package files failed: %v", err)
	}
	typeErrors := make([]error, 0)
	configuration := types.Config{Importer: newModuleImporter(root), Error: func(err error) {
		if err != nil {
			typeErrors = append(typeErrors, err)
		}
	}}
	if _, err := configuration.Check(filepath.ToSlash(filepath.Dir(logical)), fset, files, nil); err != nil {
		typeErrors = append(typeErrors, err)
	}
	if len(typeErrors) != 0 {
		sort.Slice(typeErrors, func(left, right int) bool { return typeErrors[left].Error() < typeErrors[right].Error() })
		t.Fatalf("candidate strict type-check failed: %s", typeErrors[0])
	}
}

func runCallbackPreviewFixtureSuite(t *testing.T, root, logical, candidateSource string) {
	t.Helper()
	tempRoot := t.TempDir()
	for _, relative := range []string{"go.mod", "cmd/language-readiness-witness/predecessor-selection", "internal", "examples/causal-ci-selection/pagination-fixtures.json"} {
		if err := copyCallbackPreviewPath(root, tempRoot, relative); err != nil {
			t.Fatalf("copy callback preview fixture tree: %v", err)
		}
	}
	if candidateSource != "" {
		path := filepath.Join(tempRoot, filepath.FromSlash(logical))
		if err := os.WriteFile(path, []byte(candidateSource), 0o644); err != nil {
			t.Fatalf("write generated callback preview: %v", err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "./cmd/language-readiness-witness/predecessor-selection", "-run", "^TestPaginationFixturesExecuteParserAndHTTPClient$")
	command.Dir = tempRoot
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("callback preview fixture suite candidate=%t: %v\n%s", candidateSource != "", err, output)
	}
}

func copyCallbackPreviewPath(root, destination, relative string) error {
	source := filepath.Join(root, filepath.FromSlash(relative))
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("source tree contains symlink %s", path)
		}
		relativePath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relativePath)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
}
