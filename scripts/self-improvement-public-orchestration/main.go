package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func main() {
	mode := flag.String("mode", "", "prepare, resume, or verify")
	policy := flag.String("policy", "", "canonical orchestration .gooo source")
	source := flag.String("source", "", "canonical source .gooo")
	contract := flag.String("contract", "", "public continuity contract .gooo")
	gooo := flag.String("gooo", "", "compiled public gooo command")
	testReuse := flag.String("test-reuse", "", "compiled public test-reuse command")
	projectTest := flag.String("project-test", "", "generated project test contract")
	repoRoot := flag.String("repo-root", "", "repository root for public commands")
	out := flag.String("out", "", "caller-owned output directory")
	handoff := flag.String("handoff", "", "caller-owned orchestration handoff")
	candidate := flag.String("candidate", "", "caller-owned discovery candidate")
	authorization := flag.String("authorization", "", "caller-supplied authorization artifact")
	evidenceManifest := flag.String("evidence-manifest", "", "verification evidence manifest")
	output := flag.String("output", "", "verification report")
	humanOutput := flag.String("human-output", "", "verification dossier")
	flag.Parse()

	var err error
	switch *mode {
	case "prepare":
		err = runPrepare(orchestrationInput{Policy: *policy, Source: *source, Gooo: *gooo, RepoRoot: *repoRoot, Output: *out})
	case "resume":
		err = runResume(orchestrationInput{Policy: *policy, Source: *source, Contract: *contract, Gooo: *gooo, TestReuse: *testReuse, ProjectTest: *projectTest, RepoRoot: *repoRoot, Output: *out, Handoff: *handoff, Candidate: *candidate, Authorization: *authorization})
	case "verify":
		err = runVerification(*evidenceManifest, *output, *humanOutput)
	default:
		err = errors.New("usage: self-improvement-public-orchestration -mode prepare|resume|verify")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
