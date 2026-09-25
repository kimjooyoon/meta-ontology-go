package policycompilation

import (
	_ "embed"
	"fmt"
	"runtime"
)

//go:embed testdata/go-error-guard/human-output-original.go.golden
var goHumanGuardOriginal []byte

//go:embed testdata/go-error-guard/human-output-oracle.go.golden
var goHumanGuardOracle []byte

const goHumanGuardHandlerBlock = "{\n\t\treturn exitFailure\n\t}"

func goHumanGuardFixture(source []byte, pipeline bool) []byte {
	profile := fmt.Sprintf("go-error-guard:v3;function=reportGenerateSuccess;writer=stdout;mode=jsonMode;handler-call=writeJSONReport;writes=3;source=%s;handler=%s",
		DigestBytes(source), DigestBytes([]byte(goHumanGuardHandlerBlock)))
	if pipeline {
		return []byte(fmt.Sprintf("package goerrorguard\nnamespace goerrorguard\nentity Source id \"gooo://error-guard/source\"\nentity RawCandidate id \"gooo://error-guard/raw-candidate\"\nentity Candidate id \"gooo://error-guard/candidate\"\nactivity GuardHumanOutput(Source) -> RawCandidate computes %q\nactivity Canonicalize(RawCandidate) -> Candidate computes %q\n",
			profile, "go-source-format:v1;toolchain="+runtime.Version()))
	}
	return []byte(fmt.Sprintf("package goerrorguard\nnamespace goerrorguard\nentity Source id \"gooo://error-guard/source\"\nentity Candidate id \"gooo://error-guard/candidate\"\nactivity GuardHumanOutput(Source) -> Candidate computes %q\n", profile))
}
