package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

const initUsage = "usage: gooo init [--template app|library] <new-directory>"

var starterFiles = map[string]string{
	"main.gooo": `package starter
namespace starter
entity Integer id "starter://integer"

activity Clamp(Integer) -> Integer computes "if input < 0 { return 0 } else if input < 10 { return input } else { return 10 }" assembling {
    choice "lower-bound" branch_layout at "0" intent "음수는 0으로 제한한다. Clamp negative values to zero."
    choice "upper-bound" operand_order at "1" intent "10을 넘는 값은 10으로 제한한다. Clamp values above ten to ten."
    case "-1" -> "0"
    case "5" -> "5"
    case "11" -> "10"
    attempts "4"
}
`,
	"README.md": `# Gooo starter

This project keeps its activity contract in [main.gooo](main.gooo). The
assembling block declares permitted source choices and a small finite
input/output suite. Gooo uses that contract for deterministic generation; a
local Laya service can optionally rank the eligible choices.

## Check and generate

Install the Gooo CLI, then run these commands from this directory:

~~~sh
go install github.com/kimjooyoon/meta-ontology-go/cmd/gooo@dev
gooo check main.gooo
gooo body-codegen --json --activity Clamp main.gooo
~~~

Generation prints a receipt to standard output and does not edit this project.
Without a model, the compiler uses its deterministic choice. To let local Laya
rank only the declared alternatives, set GOOO_LAYA_URL to its
/v1/systemone endpoint before running the same command. The model
cannot add choices or bypass type checking and the declared cases.

The finite score describes only these three examples; it is not a proof for all
integer inputs. See the compiler's body generation guide for the model setup and
the limits of this experiment:
https://github.com/kimjooyoon/meta-ontology-go/blob/dev/docs/language/body-codegen.md
`,
}

func runInit(args []string, stdout, stderr io.Writer) int {
	template := "app"
	if len(args) >= 2 && args[0] == "--template" {
		template = args[1]
		args = args[2:]
	}
	files := starterFiles
	if template == "library" {
		files = libraryStarterFiles
	} else if template != "app" {
		fmt.Fprintf(stderr, "gooo init: unknown template %q (choose app or library)\n", template)
		return exitUsage
	}
	if len(args) != 1 || strings.TrimSpace(args[0]) == "" || strings.HasPrefix(args[0], "-") {
		fmt.Fprintln(stderr, initUsage)
		return exitUsage
	}
	destination := filepath.Clean(args[0])
	if destination == "." || destination == string(filepath.Separator) {
		fmt.Fprintf(stderr, "gooo init: destination must be a new directory, got %q\n", args[0])
		return exitFailure
	}
	if err := os.Mkdir(destination, 0o755); err != nil {
		fmt.Fprintf(stderr, "gooo init: create %s: %v\n", destination, err)
		return exitFailure
	}
	created := make([]string, 0, len(files))
	for name, content := range files {
		path := filepath.Join(destination, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			for _, previous := range created {
				_ = os.Remove(previous)
			}
			_ = os.Remove(destination)
			fmt.Fprintf(stderr, "gooo init: write %s: %v\n", path, err)
			return exitFailure
		}
		created = append(created, path)
	}
	if template == "library" {
		fmt.Fprintf(stdout, "Created Gooo library starter in %s\n", destination)
		fmt.Fprintf(stdout, "Next: cd %s && gooo package execute --json --cases cases.json --body-plans body-fill-plans.json gooo.workspace.json\n", destination)
	} else {
		fmt.Fprintf(stdout, "Created Gooo starter in %s\n", destination)
		fmt.Fprintf(stdout, "Next: cd %s && gooo check main.gooo\n", destination)
	}
	return exitOK
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
	case "init":
		return runInit(args[1:], stdout, stderr), true
	case "package":
		return runPackageCommand(args[1:], OSFileReader{}, stdout, stderr), true
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
	case "decide":
		return runDecide(args[1:], OSFileReader{}, stdout, stderr), true
	case "generate":
		return runGenerate(args[1:], OSFileReader{}, EntityFieldsCLIParser{}, stdout, stderr), true
	case "body-context":
		return runBodyContext(args[1:], OSFileReader{}, stdout, stderr), true
	case "body-codegen":
		return runBodyCodegen(args[1:], OSFileReader{}, stdout, stderr), true
	default:
		return 0, false
	}
}

func runWithInputCommandsTwo(args []string, input io.Reader, stdout, stderr io.Writer) (int, bool) {
	switch args[0] {
	case "body-path-run":
		return runBodyPathFiles(args[1:], stdout, stderr), true
	case "body-path-stream":
		return runBodyPathStream(args[1:], input, stdout, stderr), true
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
