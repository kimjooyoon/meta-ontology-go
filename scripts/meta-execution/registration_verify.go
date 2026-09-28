package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

var registrationVerifierPackages = []string{
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/conformance",
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/verticalsliceclosureshadow",
}

func registrationVerifyCandidate(workspace, temporary string, request syntaxregistration.Request,
	compiled syntaxregistration.Plan, candidate syntaxregistration.Candidate) (generation.ProcessObservation, *operationError) {
	copyRoot := filepath.Join(temporary, "candidate")
	if err := os.Mkdir(copyRoot, 0700); err != nil {
		return generation.ProcessObservation{}, registrationNativeFailure(err, "prepare-conformance-copy")
	}
	archive := filepath.Join(temporary, "input.tar")
	packed := registrationRun(workspace, []string{"tar", "--dereference", "--exclude=.git", "-cf", "<archive>", "."},
		"tar", "--dereference", "--exclude=.git", "-cf", archive, ".")
	if packed.err != nil {
		return packed.observation, registrationProcessFailure("copy-input", packed)
	}
	unpacked := registrationRun(temporary, []string{"tar", "-xf", "<archive>", "-C", "<candidate>"},
		"tar", "-xf", archive, "-C", copyRoot)
	if unpacked.err != nil {
		return unpacked.observation, registrationProcessFailure("unpack-input", unpacked)
	}
	if err := compiled.ValidateCandidate(os.DirFS(copyRoot), candidate); err != nil {
		return generation.ProcessObservation{}, registrationNativeFailure(err, "recheck-copied-input")
	}
	if candidate.ExecutionBinding.Identity != request.ExecutionIdentity {
		return generation.ProcessObservation{}, newOperationError("OPERATION_INPUT", "compare-execution-identity",
			"REGISTRATION_IDENTITY_SUBSTITUTED", "KNOWN_CONTRADICTION", "restore-exact-execution-identity")
	}
	for _, member := range candidate.Members {
		if !fs.ValidPath(member.Path) || member.Path == "." || member.AfterDigest != digestBytes(member.Content) {
			return generation.ProcessObservation{}, newOperationError("ARTIFACT", "validate-member",
				"REGISTRATION_MEMBER_SUBSTITUTED", "KNOWN_CONTRADICTION", "restore-exact-member")
		}
		path := filepath.Join(copyRoot, filepath.FromSlash(member.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return generation.ProcessObservation{}, registrationNativeFailure(err, "prepare-member")
		}
		if err := os.WriteFile(path, member.Content, 0600); err != nil {
			return generation.ProcessObservation{}, registrationNativeFailure(err, "write-temp-member")
		}
	}
	args := []string{"test", "-count=1",
		"./internal/meta/languagereadiness/languagesyntax/conformance",
		"./internal/meta/languageassurance/verticalsliceclosureshadow"}
	process := registrationRun(copyRoot, append([]string{"go"}, args...), "go", args...)
	if process.err != nil {
		failure := registrationProcessFailure("verify-generated-candidate", process)
		failure.reason, failure.class, failure.next = "REGISTRATION_NATIVE_CONFORMANCE_REFUTED",
			"KNOWN_CONTRADICTION", "preserve-native-conformance-counterexample"
		return process.observation, failure
	}
	normalized, err := normalizeRegistrationVerifier(process.stdout)
	if err != nil {
		return process.observation, registrationNativeFailure(err, "observe-native-conformance")
	}
	// Only elapsed Go test timing is removed; raw output digests remain intact.
	process.observation.StdoutBytes = len(normalized)
	process.observation.StdoutDigest = digestBytes(normalized)
	return process.observation, nil
}

func normalizeRegistrationVerifier(raw []byte) ([]byte, error) {
	seen := make(map[string]bool)
	for line := range strings.SplitSeq(strings.TrimSpace(string(raw)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 || fields[0] != "ok" || !strings.HasSuffix(fields[2], "s") {
			return nil, fmt.Errorf("unexpected native conformance output: %q", line)
		}
		known := false
		for _, expected := range registrationVerifierPackages {
			known = known || fields[1] == expected
		}
		if !known || seen[fields[1]] {
			return nil, fmt.Errorf("unbound or duplicated verifier package: %q", fields[1])
		}
		seen[fields[1]] = true
	}
	if len(seen) != len(registrationVerifierPackages) {
		return nil, fmt.Errorf("native conformance observed %d of %d packages", len(seen), len(registrationVerifierPackages))
	}
	return []byte("PASS\t" + strings.Join(registrationVerifierPackages, "\nPASS\t") + "\n"), nil
}
