package toolchainrelease

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestConstructionReportProfileRetainsNamedFormats(t *testing.T) {
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"matching", "changed-row", "changed-json", "changed-hash", "command-failure"} {
		t.Run(mode, func(t *testing.T) {
			input := BuildInput{Root: root, OutputDir: t.TempDir(), Target: Target{ID: "fixture"}}
			calls := 0
			err := runConstructionReportSmoke("fixture-compiler", t.TempDir(), input, constructionReportTestCommand(t, input, mode, &calls))
			wantCalls := map[string]int{"matching": 6, "changed-json": 3, "changed-row": 1, "changed-hash": 1, "command-failure": 1}[mode]
			if (err == nil) != (mode == "matching") || calls != wantCalls {
				t.Fatal(mode, err, calls)
			}
			names := []string{"fixture-construction-report-canonical-text.txt"}
			if mode == "command-failure" {
				names = []string{"fixture-construction-report-canonical-text.failed-output"}
			}
			if mode == "matching" {
				names = append(names, "fixture-construction-report-canonical-markdown.md", "fixture-construction-report-canonical-json.json",
					"fixture-construction-report-replay-text.txt", "fixture-construction-report-replay-markdown.md", "fixture-construction-report-replay-json.json")
			}
			for _, name := range names {
				if raw, err := os.ReadFile(filepath.Join(input.OutputDir, name)); err != nil || len(raw) == 0 {
					t.Fatal("command output not retained", name, err)
				}
			}
		})
	}
}

func constructionReportTestCommand(t *testing.T, input BuildInput, mode string, calls *int) func(string, []string, string, ...string) ([]byte, error) {
	t.Helper()
	return func(root string, env []string, binary string, args ...string) ([]byte, error) {
		*calls++
		if root != input.Root || len(env) != 0 || binary != "fixture-compiler" || len(args) != 5 ||
			args[0] != "body-construct" || args[1] != "--report" || args[3] != "--format" {
			t.Fatal("unexpected invocation", root, env, binary, args)
		}
		saved, err := os.ReadFile(args[2])
		if err != nil || !bytes.Contains(saved, []byte(`"generated_now"`)) {
			t.Fatal("saved input missing", err)
		}
		if mode == "command-failure" {
			return []byte("failed read"), errors.New("exit1")
		}
		kind := "canonical"
		if strings.Contains(args[2], "typed-path-replay") {
			kind = "replay"
		}
		raw := constructionReportTestOutput(saved, kind, args[4])
		switch mode {
		case "changed-row":
			raw = bytes.ReplaceAll(raw, []byte("1/3 (33.33%)"), []byte("3/3 (100.00%)"))
		case "changed-hash":
			raw = bytes.ReplaceAll(raw, []byte(fmt.Sprintf("%x", sha256.Sum256(saved))), []byte("wrong"))
		case "changed-json":
			if args[4] == "json" {
				raw = append(raw, '\n')
			}
		}
		return raw, nil
	}
}

// This is only a command stub. The packaged-compiler test below consumes the
// real formatter and retained observations without constructing a new program.
func constructionReportTestOutput(saved []byte, kind, format string) []byte {
	if format == "json" {
		return append([]byte(nil), saved...)
	}
	var out strings.Builder
	fmt.Fprintln(&out, "Saved observation: every row below describes the recorded invocation.")
	fmt.Fprintf(&out, "Recorded file SHA-256: %x\n", sha256.Sum256(saved))
	fmt.Fprintln(&out, "Reading this report made 0 model calls and 0 program executions.")
	fmt.Fprintln(&out, "Recorded claims have not been reverified. Exit 0 means this report was read, including recorded failures.")
	for _, row := range constructionReportExpectedRows(kind) {
		if format == "markdown" {
			fmt.Fprintf(&out, "| %s | <code>%s</code> |\n", row[0], row[1])
		} else {
			fmt.Fprintf(&out, "%s: %s\n", row[0], strconv.Quote(row[1]))
		}
	}
	return []byte(out.String())
}

func TestConstructionReportProfileOnPackagedCompiler(t *testing.T) {
	binary := os.Getenv("GOOO_PACKAGE_SMOKE_BINARY")
	if binary == "" {
		t.Skip("set GOOO_PACKAGE_SMOKE_BINARY for an actual packaged compiler")
	}
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	input := BuildInput{Root: root, OutputDir: t.TempDir(), Target: Target{ID: "local-native"}}
	if output := os.Getenv("GOOO_REPORT_SMOKE_OUTPUT"); output != "" {
		input.OutputDir = output
		if err := os.MkdirAll(output, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := smokeConstructionReports(binary, t.TempDir(), input); err != nil {
		t.Fatal(err)
	}
}
