package policycompilation

import (
	"bytes"
	"strings"
	"testing"
)

func TestGoHumanGuardProducesThreeDeclaredGuardsAndCanonicalOutput(t *testing.T) {
	original := bytes.Clone(goHumanGuardOriginal)
	raw, err := ProposeGoErrorGuard("human-guard.gooo", goHumanGuardFixture(original, false), original)
	if err != nil || raw.State != "PROPOSED" || raw.GuardedCalls != 3 {
		t.Fatalf("raw human guard: state=%s calls=%d error=%v", raw.State, raw.GuardedCalls, err)
	}
	if strings.Count(raw.CandidateSource, "if _, err := fmt.Fprintf(stdout,") != 3 ||
		!strings.HasPrefix(raw.CandidateSource, string(original[:raw.EditStart])) ||
		!strings.HasSuffix(raw.CandidateSource, string(original[raw.EditEnd:])) {
		t.Fatal("human guard call count or source boundary differs")
	}
	program := goHumanGuardFixture(original, true)
	canonical, err := ProposeGoErrorGuardPipeline("human-pipeline.gooo", program, original)
	if err != nil || canonical.Guard == nil || canonical.Guard.GuardedCalls != 3 ||
		canonical.Rendering == nil || !canonical.Rendering.FixedPoint || !canonical.Rendering.OutsideEditPreserved {
		t.Fatalf("canonical human guard: %#v error=%v", canonical, err)
	}
	replay, err := ProposeGoErrorGuardPipeline("human-pipeline.gooo", program, original)
	if err != nil || canonical.CandidateDigest != replay.CandidateDigest ||
		canonical.Guard.CandidateDigest == canonical.CandidateDigest ||
		canonical.Admission.State != "UNKNOWN" || canonical.MutationAuthority != 0 ||
		canonical.PromotionAuthority != 0 || !bytes.Equal(original, goHumanGuardOriginal) {
		t.Fatalf("human guard replay, authority or input differs: error=%v", err)
	}
}

func TestGoHumanGuardContradictionsAndInvalidProfilesDoNotPropose(t *testing.T) {
	program := string(goHumanGuardFixture(goHumanGuardOriginal, false))
	cases := []struct {
		name    string
		program string
		source  []byte
		reason  string
	}{
		{"source", program, append(bytes.Clone(goHumanGuardOriginal), '\n'), "GO_SOURCE_PIN_MISMATCH"},
		{"handler", strings.Replace(program, "handler="+DigestBytes([]byte(goHumanGuardHandlerBlock)), "handler="+DigestBytes([]byte("different")), 1), goHumanGuardOriginal, "GO_HANDLER_PIN_MISMATCH"},
		{"count", strings.Replace(program, "writes=3", "writes=2", 1), goHumanGuardOriginal, "GO_HUMAN_GUARD_WRITE_COUNT_MISMATCH"},
		{"noncanonical-count", strings.Replace(program, "writes=3", "writes=03", 1), goHumanGuardOriginal, "GOOO_GUARD_PROFILE_INVALID"},
		{"zero-count", strings.Replace(program, "writes=3", "writes=0", 1), goHumanGuardOriginal, "GOOO_GUARD_PROFILE_INVALID"},
		{"implicit-v2", strings.Replace(program, "go-error-guard:v3", "go-error-guard:v2", 1), goHumanGuardOriginal, "GOOO_GUARD_PROFILE_INVALID"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			proposal, err := ProposeGoErrorGuard("invalid-human.gooo", []byte(test.program), test.source)
			if err == nil || proposal.State != "REFUTED" || proposal.Reason != test.reason ||
				proposal.CandidateSource != "" || proposal.GuardedCalls != 0 ||
				proposal.MutationAuthority != 0 || proposal.PromotionAuthority != 0 {
				t.Fatalf("invalid human guard: %#v error=%v", proposal, err)
			}
		})
	}
}

func TestGoHumanGuardUnsupportedShapesRetainCausalUnknown(t *testing.T) {
	cases := []struct {
		name string
		from string
		to   string
	}{
		{"mode", "if !jsonMode {", "if jsonMode {"},
		{"branch-binding", "if discovery != nil {", "if discovery := discovery; discovery != nil {"},
		{"concurrent-write", "fmt.Fprintf(stdout, \"generated:", "go fmt.Fprintf(stdout, \"generated:"},
		{"writer-type", "stdout io.Writer", "stdout interfaceWriter"},
		{"writer-escape", "filepath.Join(options.outputDir, generatedFileName)", "fmt.Sprint(stdout)"},
		{"handler-shadow", "stdout io.Writer) int", "stdout io.Writer, writeJSONReport func(io.Writer, any) error) int"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			source := []byte(strings.Replace(string(goHumanGuardOriginal), test.from, test.to, 1))
			proposal, err := ProposeGoErrorGuard("unsupported-human.gooo", goHumanGuardFixture(source, false), source)
			if err == nil || proposal.State != "UNKNOWN" || proposal.CandidateSource != "" ||
				proposal.Pending == nil || proposal.Pending.UnknownClass != "UNBOUNDED" ||
				proposal.Pending.Stage == "" || proposal.Pending.Step == "" || proposal.Pending.Reason == "" ||
				proposal.Pending.NextOperation == "" || proposal.Pending.BlockedBy == nil {
				t.Fatalf("unsupported human guard: %#v error=%v", proposal, err)
			}
		})
	}
}

func TestGoHumanGuardMissingFunctionRemainsDirectMissing(t *testing.T) {
	program := bytes.Replace(goHumanGuardFixture(goHumanGuardOriginal, false),
		[]byte("function=reportGenerateSuccess"), []byte("function=missing"), 1)
	proposal, err := ProposeGoErrorGuard("missing-human.gooo", program, goHumanGuardOriginal)
	if err == nil || proposal.State != "UNKNOWN" || proposal.Pending == nil ||
		proposal.Pending.UnknownClass != "DIRECT_MISSING" || len(proposal.Pending.BlockedBy) != 0 {
		t.Fatalf("missing human guard: %#v error=%v", proposal, err)
	}
}
