package toolchainformatfix

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/formatfix"
	cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

const testHead = "0000000000000000000000000000000000000000"
const testDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

type fakeExecutor struct{ calls int }

func (*fakeExecutor) BinaryDigest() (string, error) { return testDigest, nil }
func (fake *fakeExecutor) Invoke(arguments []string) (cliruntime.Observation, error) {
	fake.calls++
	exit, stdout, stderr := fakeOutput(strings.Join(arguments, "\x00"))
	return cliruntime.Observation{Arguments: append([]string(nil), arguments...), ExitCode: exit,
		Stdout: stdout, Stderr: stderr, TreeBeforeDigest: testDigest,
		TreeAfterDigest: testDigest}, nil
}

func registryFixture(t interface{ Fatal(...any) }) []byte {
	raw, err := os.ReadFile("../../../../examples/toolchain-format-fix/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func fakeOutput(key string) (int, string, string) {
	change := formatfix.Build(unformattedPath, unformatted)
	fixed := formatfix.Build(canonicalPath, canonical)
	switch key {
	case "format\x00" + unformattedPath:
		return 0, canonical, ""
	case "format\x00--json\x00" + unformattedPath:
		raw, _ := json.Marshal(map[string]any{"schema": "gooo/format-report/v1", "command": "format",
			"status": "formatted", "file": unformattedPath, "changed": true, "source": canonical,
			"source_digest": change.SourceDigest, "formatted_digest": change.ResultDigest,
			"diagnostics": []string{}, "direct_writes": 0})
		return 0, string(raw) + "\n", ""
	case "format\x00--check\x00" + canonicalPath:
		return 0, "canonical: " + canonicalPath + "\n", ""
	case "fix\x00--json\x00" + unformattedPath:
		raw, _ := json.Marshal(change)
		return 0, string(raw) + "\n", ""
	case "fix\x00--json\x00" + canonicalPath:
		raw, _ := json.Marshal(fixed)
		return 0, string(raw) + "\n", ""
	case "fix\x00" + unformattedPath:
		return 0, fmt.Sprintf("%s: %s edits=1 writes=0 digest=%s\n",
			change.Decision, unformattedPath, change.PlanDigest), ""
	default:
		return fakeGuardrail(key)
	}
}

func fakeGuardrail(key string) (int, string, string) {
	switch key {
	case "format":
		return 2, "", "usage: gooo format [--check] [--json] <file.gooo>\n"
	case "format\x00" + malformedPath:
		return 1, "", "gooo: " + malformedPath + ": FORMAT_FIX_SOURCE_UNKNOWN\n"
	case "fix\x00" + malformedPath:
		return 1, "", "gooo: " + malformedPath + ": FORMAT_FIX_SOURCE_UNKNOWN\n"
	case "format\x00--unknown\x00" + canonicalPath:
		return 2, "", "usage: gooo format [--check] [--json] <file.gooo>\n"
	default:
		return 2, "", "usage: gooo fix [--json] <file.gooo>\n"
	}
}
