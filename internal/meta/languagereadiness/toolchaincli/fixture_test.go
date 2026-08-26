package toolchaincli

import (
	"os"
	"strings"

	cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

const testHead = "0000000000000000000000000000000000000000"
const testDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

type fakeExecutor struct {
	calls, writeOn int
}

func (*fakeExecutor) BinaryDigest() (string, error) { return testDigest, nil }

func (fake *fakeExecutor) Invoke(arguments []string) (cliruntime.Observation, error) {
	fake.calls++
	result := fakeObservation(arguments)
	if fake.calls == fake.writeOn {
		result.RepositoryWrites, result.TreeAfterDigest = 1, digestJSON("changed")
	}
	return result, nil
}

func registryFixture(t interface{ Fatal(...any) }) []byte {
	raw, err := os.ReadFile("../../../../examples/toolchain-cli/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func fakeObservation(arguments []string) cliruntime.Observation {
	result := cliruntime.Observation{Arguments: append([]string(nil), arguments...),
		TreeBeforeDigest: testDigest, TreeAfterDigest: testDigest}
	result.ExitCode, result.Stdout, result.Stderr = fakeOutput(strings.Join(arguments, "\x00"))
	return result
}
