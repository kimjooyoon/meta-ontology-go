package syntaxregistration

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func nativeCommand(t *testing.T, root string, args ...string) []byte {
	t.Helper()
	context, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	command := exec.CommandContext(context, args[0], args[1:]...)
	command.Dir = root
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "GIT_") && !strings.HasPrefix(value, "GOWORK=") &&
			!strings.HasPrefix(value, "GOTOOLCHAIN=") {
			command.Env = append(command.Env, value)
		}
	}
	command.Env = append(command.Env, "GOWORK=off", "GOTOOLCHAIN=local")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("native command %v failed: %v\n%s", args, err, output)
	}
	return output
}

func TestNativeNineMemberCandidatePassesExistingConformance(t *testing.T) {
	runNativeRegistrationConformance(t, "example")
}

func TestNativeInternalMetaSourceRegistrationPassesExistingConformance(t *testing.T) {
	runNativeRegistrationConformance(t, "internal-meta-source")
}

func TestNativeRegisteredMetaSourcePromotionPassesExistingConformance(t *testing.T) {
	runNativeRegistrationConformance(t, "registered-meta-source")
}

func nativeRegistrationInput(t *testing.T, mode string) (Request, []byte) {
	t.Helper()
	switch mode {
	case "internal-meta-source":
		data, request := internalMetaFixture(t)
		return request, data[request.Case.Path].Data
	case "registered-meta-source":
		data, request := promotedMetaFixture(t)
		return request, data[request.Case.Path].Data
	case "example":
		_, request := fixture(t)
		return request, []byte(fixtureSource)
	default:
		t.Fatalf("unsupported native registration case: %s", mode)
		return Request{}, nil
	}
}

func prepareNativeRegistrationSource(t *testing.T, snapshot string, request Request, source []byte) {
	t.Helper()
	inputPath := filepath.Join(snapshot, filepath.FromSlash(request.Case.Path))
	if request.PromoteMetaSource {
		observed, err := os.ReadFile(inputPath)
		if err != nil || !bytes.Equal(observed, source) {
			t.Fatalf("registered source differs from the pinned fixture: %v", err)
		}
		return
	}
	if err := os.MkdirAll(filepath.Dir(inputPath), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inputPath, source, 0600); err != nil {
		t.Fatal(err)
	}
}

func runNativeRegistrationConformance(t *testing.T, mode string) {
	t.Helper()
	if os.Getenv("GOOO_SYNTAX_REGISTRATION_E2E") != "1" || os.Getenv("CI") != "true" {
		t.Skip("native candidate application runs only in its dedicated GitHub Actions job")
	}
	started := time.Now()
	original, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	temporary := t.TempDir()
	snapshot := filepath.Join(temporary, "project")
	if err := os.Mkdir(snapshot, 0700); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(temporary, "source.tar")
	view := os.Getenv("GOOO_SYNTAX_REGISTRATION_SOURCE_VIEW")
	if view == "" {
		view = "canonical"
	}
	switch view {
	case "canonical":
		nativeCommand(t, original, "git", "archive", "--format=tar", "--output="+archive, "HEAD")
	case "projected":
		// Preserve the actual projected input, including created units. Dereference
		// logical-workspace links so applying the copy cannot write through them.
		nativeCommand(t, original, "tar", "--dereference", "--exclude=.git", "-cf", archive, ".")
	default:
		t.Fatalf("unsupported native source view: %s", view)
	}
	nativeCommand(t, original, "tar", "-xf", archive, "-C", snapshot)
	request, source := nativeRegistrationInput(t, mode)
	prepareNativeRegistrationSource(t, snapshot, request, source)
	repository := os.DirFS(snapshot)
	pin(t, repository, &request)
	plan, err := Compile(repository, request)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := plan.Generate(repository)
	if err != nil {
		t.Fatal(err)
	}
	if candidate.RequiredArtifacts != 9 || len(candidate.Artifacts) != 9 ||
		candidate.Required != candidate.Emitted || candidate.Emitted != len(candidate.Members) {
		t.Fatal("native candidate lost a semantic role or physical member")
	}
	if candidate.ExecutionBinding.Identity != request.ExecutionIdentity ||
		candidate.ExecutionBinding.ActivityID == "" || candidate.ExecutionBinding.OutputID == "" {
		t.Fatal("native candidate lost its observed execution identity binding")
	}
	if view == "canonical" && candidate.Required != RequiredMembers {
		t.Fatalf("canonical acceptance requires exactly nine files, got %d", candidate.Required)
	}
	replayed, err := plan.Generate(repository)
	if err != nil {
		t.Fatal(err)
	}
	if digestValue(candidate) != digestValue(replayed) {
		t.Fatal("native candidate replay did not match")
	}
	if err := plan.ValidateCandidate(repository, candidate); err != nil {
		t.Fatal(err)
	}
	for _, member := range candidate.Members {
		target := filepath.Join(snapshot, filepath.FromSlash(member.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, member.Content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	output := nativeCommand(t, snapshot, "go", "test", "-count=1",
		"./internal/meta/languagereadiness/languagesyntax/conformance",
		"./internal/meta/languageassurance/verticalsliceclosureshadow")
	if directory := os.Getenv("GOOO_SYNTAX_REGISTRATION_EVIDENCE_DIR"); directory != "" {
		if mode != "example" {
			directory = filepath.Join(directory, mode)
		}
		if err := os.MkdirAll(directory, 0700); err != nil {
			t.Fatal(err)
		}
		report := map[string]any{"operation": Operation, "candidate_digest": digestValue(candidate),
			"native_conformance": "PASS", "emitted_members": candidate.Emitted, "required_members": candidate.Required,
			"required_artifacts": candidate.RequiredArtifacts, "generated_artifacts": len(candidate.Artifacts),
			"source_view": view, "case_id": request.Case.ID, "source_path": request.Case.Path,
			"source_digest": request.SourceDigest, "input_digest": request.SnapshotDigest,
			"request_digest": digestValue(request), "evidence_class": "SYNTHETIC",
			"source_promotion_requested": request.PromoteMetaSource,
			"base_denominator_version":   request.BaseVersion, "candidate_denominator_version": request.BaseVersion + 1,
			"execution_binding":      candidate.ExecutionBinding,
			"manual_follow_up_edits": 0, "replay_comparisons": 1, "repository_writes": 0,
			"apply_scope": "CALLER_OWNED_CI_TEMP_COPY", "semantic_admission": "UNASSESSED",
			"global_planner_admission": "UNASSESSED", "wall_ms": time.Since(started).Milliseconds()}
		raw, _ := json.MarshalIndent(report, "", "  ")
		bundle, _ := json.MarshalIndent(candidate, "", "  ")
		requestDocument, _ := json.MarshalIndent(request, "", "  ")
		for name, data := range map[string][]byte{"candidate.json": bundle, "native-evaluation.json": raw, "native-conformance.txt": output,
			"request.json": requestDocument, "source.gooo": source} {
			if err := os.WriteFile(filepath.Join(directory, name), data, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
}
