package syntaxregistration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecutionIdentityPinsThreeBinariesAndLoweredInput(t *testing.T) {
	data, request := fixture(t)
	observed, err := ObserveExecutionIdentity()
	if err != nil {
		t.Fatal(err)
	}
	if err := validateExecutionIdentity(request.ExecutionIdentity, observed); err != nil {
		t.Fatal(err)
	}
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := plan.Generate(data)
	if err != nil {
		t.Fatal(err)
	}
	binding := candidate.ExecutionBinding
	if binding.Identity != observed || binding.ActivityID != plan.binding.inputActivity ||
		binding.InputID != "gooo://syntax-registration/request" || binding.OutputID != plan.binding.inputOutput {
		t.Fatalf("candidate lost the lowered identity binding: %#v", binding)
	}
	candidate.ExecutionBinding.Identity.CompilerDigest = "sha256:" + strings.Repeat("0", 64)
	requireFailure(t, plan.ValidateCandidate(data, candidate), "REFUTED", "")
}

func TestVersionOnlyOrChangedExecutionIdentityCannotGenerate(t *testing.T) {
	data, request := fixture(t)
	for _, part := range []string{"missing", "generator", "driver", "compiler", "goos", "goarch", "malformed"} {
		t.Run(part, func(t *testing.T) {
			changed := request
			state, class := "UNKNOWN", "STALE"
			switch part {
			case "missing":
				changed.ExecutionIdentity = ExecutionIdentity{}
				class = "DIRECT_MISSING"
			case "generator":
				changed.ExecutionIdentity.ExecutableDigest = "sha256:" + strings.Repeat("0", 64)
			case "driver":
				changed.ExecutionIdentity.GoCommandDigest = "sha256:" + strings.Repeat("0", 64)
			case "compiler":
				changed.ExecutionIdentity.CompilerDigest = "sha256:" + strings.Repeat("0", 64)
			case "goos":
				changed.ExecutionIdentity.GOOS = "different-os"
			case "goarch":
				changed.ExecutionIdentity.GOARCH = "different-arch"
			case "malformed":
				changed.ExecutionIdentity.CompilerDigest = "sha256:not-a-digest"
				state, class = "REFUTED", ""
			}
			_, err := Compile(data, changed)
			requireFailure(t, err, state, class)
		})
	}
	plan, err := Compile(data, request)
	if err != nil {
		t.Fatal(err)
	}
	plan.request.ExecutionIdentity.GoCommandDigest = "sha256:" + strings.Repeat("0", 64)
	candidate, err := plan.Generate(data)
	requireFailure(t, err, "UNKNOWN", "STALE")
	if candidate.Emitted != 0 || candidate.ApplyAuthorized || candidate.PromotionAllowed {
		t.Fatal("changed execution identity emitted a candidate or authority")
	}
}

func TestExecutionBinaryDigestTracksBytesNotVersionLabel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "compiler")
	if err := os.WriteFile(path, []byte("same-version-first-binary"), 0600); err != nil {
		t.Fatal(err)
	}
	first, err := hashExecutionFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("same-version-second-binary"), 0600); err != nil {
		t.Fatal(err)
	}
	second, err := hashExecutionFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !executionDigestValid(first) || !executionDigestValid(second) {
		t.Fatal("binary content change was hidden by an identity label")
	}
	if _, err := hashExecutionFile(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("missing execution binary acquired an identity")
	}
}

func TestRegistrationContractRejectsUnpinnedInputRoute(t *testing.T) {
	for _, mutation := range []struct{ before, after string }{
		{"PinRegistrationExecutionIdentity(RegistrationRequest)", "PinRegistrationExecutionIdentity(RegistrationCandidate)"},
		{"RegisterSyntaxCapability(PinnedRegistrationInput)", "RegisterSyntaxCapability(RegistrationRequest)"},
		{"GenerateSyntaxCorpus(PinnedRegistrationInput)", "GenerateSyntaxCorpus(RegistrationRequest)"},
		{"syntax.register:v2", "syntax.register:v1"},
	} {
		source := bytes.Replace(contractSource, []byte(mutation.before), []byte(mutation.after), 1)
		if _, err := bindContractSource(source); err == nil {
			t.Fatalf("unbound execution route accepted: %s", mutation.before)
		}
	}
}
