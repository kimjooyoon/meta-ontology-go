package toolchainrelease

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Read retained observations on each release platform. No original construction,
// prediction, native evaluation or replay is repeated by this profile.
func smokeConstructionReports(binary, work string, input BuildInput) error {
	return runConstructionReportSmoke(binary, work, input, commandOutput)
}

func runConstructionReportSmoke(binary, work string, input BuildInput,
	run func(string, []string, string, ...string) ([]byte, error)) error {
	for _, sample := range []struct{ kind, directory, name string }{
		{"canonical", "cmd/gooo/testdata", "canonical-native-construct-20261011"},
		{"replay", "internal/meta/languagereadiness/toolchainrelease/testdata", "typed-path-replay"},
	} {
		if err := unpackOutcomeInput(filepath.Join(input.Root, sample.directory), work, sample.name); err != nil {
			return err
		}
		path := filepath.Join(work, sample.name+".json")
		saved, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, format := range []string{"text", "markdown", "json"} {
			args := []string{"body-construct", "--report", path, "--format", format}
			raw, runErr := run(input.Root, nil, binary, args...)
			if err := retainConstructionReport(input, sample.kind, format, raw, runErr); err != nil {
				return err
			}
			if err := validateConstructionReportOutput(raw, saved, sample.kind, format); err != nil {
				return fmt.Errorf("TOOLCHAIN_RELEASE_CONSTRUCTION_REPORT %s/%s: %w", sample.kind, format, err)
			}
		}
	}
	return nil
}

func retainConstructionReport(input BuildInput, kind, format string, raw []byte, runErr error) error {
	extension := map[string]string{"text": ".txt", "markdown": ".md", "json": ".json"}[format]
	if runErr != nil {
		extension = ".failed-output"
	}
	name := input.Target.ID + "-construction-report-" + kind + "-" + format + extension
	if err := os.WriteFile(filepath.Join(input.OutputDir, name), raw, 0o644); err != nil {
		return errors.Join(runErr, fmt.Errorf("retain construction report: %w", err))
	}
	return runErr
}

func validateConstructionReportOutput(raw, saved []byte, kind, format string) error {
	if format == "json" {
		if !bytes.Equal(raw, saved) {
			return fmt.Errorf("saved JSON bytes changed")
		}
		return nil
	}
	text := string(raw)
	for _, required := range []string{"Saved observation: every row below describes the recorded invocation.",
		fmt.Sprintf("Recorded file SHA-256: %x", sha256.Sum256(saved)),
		"Reading this report made 0 model calls and 0 program executions.",
		"Recorded claims have not been reverified.", "Exit 0 means this report was read, including recorded failures."} {
		if !strings.Contains(text, required) {
			return fmt.Errorf("saved observation boundary missing: %s", required)
		}
	}
	if format == "markdown" {
		text = html.UnescapeString(text)
	}
	for _, row := range constructionReportExpectedRows(kind) {
		expected := row[0] + ": " + strconv.Quote(row[1])
		if format == "markdown" {
			expected = "| " + row[0] + " | <code>" + row[1] + "</code> |"
		}
		if !strings.Contains("\n"+text, "\n"+expected+"\n") {
			return fmt.Errorf("recorded row differs: %s", row[0])
		}
	}
	return nil
}

func constructionReportExpectedRows(kind string) [][2]string {
	if kind == "canonical" {
		return [][2]string{
			{"Mode", "new construction"}, {"Attempted programs / budget", "1 / 4"},
			{"Initial local path search", "4 recorded candidates; 1 recorded local model calls"},
			{"First local candidate", "mask 3; CONDITION_REJECTED"},
			{"First local candidate output checks", "8/8 (100.00%)"},
			{"First local candidate recorded condition checks", "1/3 (33.33%); 0 not reached"},
			{"First unmet local condition", "choice \"comparison\"; input 66; expected true; observed false"},
			{"Selected caller-program body checks", "mask 0; outputs 8/8 (100.00%); conditions 3/3 (100.00%); 0 not reached"},
			{"Evaluation input overlap", "8 unique tuples: 8 consumed during caller construction, 0 other; 0 duplicate rows"},
		}
	}
	return [][2]string{
		{"Mode", "saved construction replay"}, {"Attempted programs / budget", "11 / 16"},
		{"Selected attempt", "11"}, {"Saved history replayed", "true"},
		{"New model calls during replay/evaluation", "0"},
		{"Evaluation expected outputs", "3/3 (100.00%) matched; 0 mismatched, 0 faulted, 0 blocked, 0 unobserved; 0 additional unscored faults/blocks"},
		{"Evaluation input overlap", "3 unique tuples: 0 consumed during caller construction, 3 other; 0 duplicate rows"},
	}
}
