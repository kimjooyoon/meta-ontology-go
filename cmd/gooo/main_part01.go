package main

import (
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/languageprofile"
)

const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

func main() { os.Exit(runWithInput(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
func run(args []string, stdout, stderr io.Writer) int {
	return runWithInput(args, os.Stdin, stdout, stderr)
}
func runWithInput(args []string, input io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return exitUsage
	}
	if result, handled := runWithInputCommandsOne(args, stdout, stderr); handled {
		return result
	}
	if result, handled := runWithInputCommandsTwo(args, input, stdout, stderr); handled {
		return result
	}
	return runExtensionCommand(args, stdout, stderr)
}

func runWithInputCommandsOne(args []string, stdout, stderr io.Writer) (int, bool) {
	switch args[0] {
	case "run":
		return runSource(args[1:], OSFileReader{}, stdout, stderr), true
	case "compare":
		return runCompareReplay(args[1:], OSFileReader{}, stdout, stderr), true
	case "propose-repair":
		return runProposeRepair(args[1:], OSFileReader{}, stdout, stderr), true
	case "consume-repair":
		return runConsumeRepair(args[1:], OSFileReader{}, stdout, stderr), true
	case "revise-source":
		return runReviseSource(args[1:], OSFileReader{}, stdout, stderr), true
	case "revise-from-handoff":
		return runReviseFromHandoff(args[1:], OSFileReader{}, stdout, stderr), true
	case "evaluate-revision":
		return runEvaluateRevision(args[1:], OSFileReader{}, stdout, stderr), true
	case "verify-revision-contract":
		return runVerifyRevisionContract(args[1:], OSFileReader{}, stdout, stderr), true
	case "run-accepted-revision":
		return runAcceptedRevision(args[1:], OSFileReader{}, stdout, stderr), true
	case "compare-accepted-revision":
		return runCompareAcceptedRevision(args[1:], OSFileReader{}, stdout, stderr), true
	case "stage-accepted-revision":
		return runStageAcceptedRevision(args[1:], OSFileReader{}, stdout, stderr), true
	case "profile":
		return runProfile(args[1:], OSFileReader{}, languageprofile.RuntimeMeasurer{}, stdout, stderr), true
	case "debug":
		return runDebug(args[1:], stdout, stderr), true
	case "test":
		return runLanguageTest(args[1:], OSFileReader{}, stdout, stderr), true
	case "check":
		return runCheck(args[1:], OSFileReader{}, EntityFieldsCLIParser{}, stdout, stderr), true
	case "generate":
		return runGenerate(args[1:], OSFileReader{}, EntityFieldsCLIParser{}, stdout, stderr), true
	default:
		return 0, false
	}
}

func runWithInputCommandsTwo(args []string, input io.Reader, stdout, stderr io.Writer) (int, bool) {
	switch args[0] {
	case "observe":
		return runObserve(args[1:], OSFileReader{}, EntityFieldsCLIParser{}, stdout, stderr), true
	case "propose":
		return runAdoptionProposal(args[1:], OSFileReader{}, stdout, stderr), true
	case "adopt":
		return runAdoption(args[1:], OSFileReader{}, EntityFieldsCLIParser{}, stdout, stderr), true
	case "roundtrip":
		return runRoundTrip(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr), true
	case "query":
		return runQuery(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr), true
	case "inspect":
		return runInspect(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr), true
	case "graph":
		return runPublicGraph(args[1:], OSFileReader{}, stdout, stderr), true
	case "claim":
		return runClaim(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr), true
	case "analyze":
		return runAnalyze(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr), true
	case "format":
		return runFormat(args[1:], OSFileReader{}, stdout, stderr), true
	case "fix":
		return runFix(args[1:], OSFileReader{}, stdout, stderr), true
	case "provenance":
		return runProvenance(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr), true
	case "selective-ci":
		return runSelectiveCI(args[1:], OSFileReader{}, stdout, stderr), true
	case "invoke":
		return runInvoke(args[1:], OSFileReader{}, stdout, stderr), true
	case "lsp":
		return runLSP(args[1:], input, stdout, stderr), true
	case "version":
		return runVersion(args[1:], stdout, stderr), true
	default:
		return 0, false
	}
}
