package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func parseSavedConstructionFlags(flags map[string]string) (map[string]string, error) {
	for _, key := range []string{"--source", "--cases", "--construction-cases", "--attempts", "--entry",
		"--model", "--fill-model", "--construction", "--go-bin", "--out"} {
		if flags[key] != "" {
			return nil, fmt.Errorf("saved report excludes %s; use --report with --format only", key)
		}
	}
	if flags["--format"] == "" {
		flags["--format"] = "text"
	}
	return flags, nil
}

func runSavedConstructionReport(path, format string, stdout, stderr io.Writer) int {
	fail := func(err error) int { fmt.Fprintln(stderr, "gooo body-construct report:", err); return exitFailure }
	raw, err := readBodyExecutionFile(path, 32<<20)
	if err != nil {
		return fail(err)
	}
	result, err := decodeSavedConstructionReport(raw)
	if err != nil {
		return fail(err)
	}
	if err := writeSavedConstructionReport(stdout, raw, result, format); err != nil {
		return fail(err)
	}
	return exitOK
}

// Decode the full saved CLI envelope, including failed or partial observations.
// Do not load source/models, reconstruct candidates, or replay a program here.
func decodeSavedConstructionReport(raw []byte) (bodyConstructOutput, error) {
	var saved struct {
		GeneratedNow *bool                            `json:"generated_now"`
		Construction *bodyexecution.JointConstruction `json:"construction"`
		Evaluation   *bodyexecution.JointEvaluation   `json:"evaluation"`
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&saved); err != nil {
		return bodyConstructOutput{}, fmt.Errorf("read full body-construct JSON output: %w", err)
	}
	if saved.GeneratedNow == nil || saved.Construction == nil || saved.Evaluation == nil {
		return bodyConstructOutput{}, fmt.Errorf("report requires generated_now, construction and evaluation from full body-construct --format json output")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return bodyConstructOutput{}, fmt.Errorf("report requires one JSON observation")
	}
	return bodyConstructOutput{GeneratedNow: *saved.GeneratedNow, Construction: *saved.Construction,
		Evaluation: *saved.Evaluation}, nil
}

func writeSavedConstructionReport(out io.Writer, raw []byte, result bodyConstructOutput, format string) error {
	if format == "json" {
		_, err := out.Write(raw)
		return err
	}
	var report strings.Builder
	fmt.Fprintln(&report, "Saved observation: every row below describes the recorded invocation.")
	fmt.Fprintf(&report, "Recorded file SHA-256: %x\n", sha256.Sum256(raw))
	fmt.Fprintln(&report, "Reading this report made 0 model calls and 0 program executions.")
	fmt.Fprint(&report, "Recorded claims have not been reverified. Exit 0 means this report was read, including recorded failures.\n\n")
	if err := writeBodyConstructResult(&report, result, format); err != nil {
		return err
	}
	_, err := io.WriteString(out, report.String())
	return err
}
