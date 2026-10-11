// producer records one fixed integration run of the published contract model.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

const modelSHA = "c53ef0386dfc8ee2422b2b9605b609b0c23315905599ecc2f4555c26de514b5f"

var names = []string{"bound-k7-r0-w0-direct-g0", "bound-k7-r0-w0-direct-g1",
	"offset-k11-r1-w0-assignment-g0", "offset-k11-r1-w0-assignment-g1"}

type nativeCase struct {
	Inputs   map[string]int64 `json:"inputs"`
	Expected map[string]int64 `json:"expected"`
}

type nativeCases struct {
	Schema string       `json:"schema"`
	Cases  []nativeCase `json:"cases"`
}

type commandRecord struct {
	Args     []string `json:"args"`
	WallNS   int64    `json:"wall_ns"`
	ExitCode int      `json:"exit_code"`
	Error    string   `json:"error,omitempty"`
}

func save(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	closeErr := f.Close()
	return errors.Join(err, closeErr)
}

func saveJSON(path string, data any) error {
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return save(path, append(raw, '\n'))
}

func observe(root, name, compiler string, args ...string) error {
	out, err := os.OpenFile(filepath.Join(root, name+".json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer out.Close()
	diagnostics, err := os.OpenFile(filepath.Join(root, name+".time-stderr"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer diagnostics.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/usr/bin/time", append([]string{"-lp", compiler}, args...)...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = root, out, diagnostics
	start := time.Now()
	err = cmd.Run()
	record := commandRecord{Args: args, WallNS: time.Since(start).Nanoseconds()}
	if err != nil {
		record.ExitCode, record.Error = -1, err.Error()
		if cmd.ProcessState != nil {
			record.ExitCode = cmd.ProcessState.ExitCode()
		}
	}
	return errors.Join(err, saveJSON(filepath.Join(root, name+".command.json"), record))
}

func prepare(root, fixtures, name string, model []byte) error {
	if err := os.Mkdir(root, 0700); err != nil {
		return err
	}
	source, err := os.ReadFile(filepath.Join(fixtures, name+".gooo"))
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(filepath.Join(fixtures, name+".json"))
	if err != nil {
		return err
	}
	doc, err := pathplan.DecodeDocument(raw)
	if err != nil {
		return err
	}
	rows := nativeCases{Schema: "gooo/body-composition-cases/v1"}
	for _, c := range doc.TestCases {
		rows.Cases = append(rows.Cases, nativeCase{map[string]int64{"Main": c.Input}, map[string]int64{"Main": c.Expected}})
	}
	if len(rows.Cases) != 8 || len(doc.Plan.ConditionCases) != 3 || doc.MaxAttempts != 4 {
		return errors.New("fixed source contract changed")
	}
	source = append(source, []byte("\nactivity Main(Integer) -> Integer computes \"return Choose(input)\"\n")...)
	return errors.Join(save(filepath.Join(root, "source.gooo"), source), save(filepath.Join(root, "model.json"), model),
		save(filepath.Join(root, "original-document.json"), raw), saveJSON(filepath.Join(root, "evaluation-cases.json"), rows),
		saveJSON(filepath.Join(root, "construction-cases.json"), nativeCases{rows.Schema, rows.Cases[1:2]}))
}

func measure(root, compiler, goBin string) error {
	commands := []struct {
		name string
		args []string
	}{
		{"preflight", []string{"body-context", "--activity", "Choose", "--model", "model.json", "source.gooo"}},
		{"deterministic", []string{"body-codegen", "--json", "--activity", "Choose", "source.gooo"}},
		{"initial", []string{"body-codegen", "--json", "--activity", "Choose", "--path-model", "model.json", "--path-step-attempts", "1", "source.gooo"}},
		{"construction", []string{"body-construct", "--source", "source.gooo", "--entry", "Main", "--model", "model.json", "--construction-cases", "construction-cases.json", "--cases", "evaluation-cases.json", "--attempts", "4", "--go-bin", goBin, "--out", "construction"}},
	}
	for _, command := range commands {
		if err := observe(root, command.name, compiler, command.args...); err != nil {
			return err
		}
	}
	if err := os.Remove(filepath.Join(root, "model.json")); err != nil {
		return err
	}
	return observe(root, "replay", compiler, "body-construct", "--source", "construction/original.gooo", "--construction", "construction/construction.json", "--cases", "evaluation-cases.json", "--go-bin", goBin)
}

func run(args []string) error {
	if len(args) != 5 {
		return errors.New("usage: producer COMPILER MODEL SDK_EXAMPLES GO_BINARY NEW_OUTPUT_DIRECTORY")
	}
	for _, arg := range args {
		if !filepath.IsAbs(arg) {
			return errors.New("absolute paths required")
		}
	}
	model, err := os.ReadFile(args[1])
	if err != nil {
		return err
	}
	if fmt.Sprintf("%x", sha256.Sum256(model)) != modelSHA {
		return errors.New("fixed model hash differs")
	}
	if err := os.Mkdir(args[4], 0700); err != nil {
		return err
	}
	if err := save(filepath.Join(args[4], "started.txt"), []byte(time.Now().UTC().Format(time.RFC3339Nano))); err != nil {
		return err
	}
	for _, name := range names {
		root := filepath.Join(args[4], name)
		if err := prepare(root, args[2], name, model); err != nil {
			return err
		}
		if err := measure(root, args[0], args[3]); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return save(filepath.Join(args[4], "completed.txt"), []byte(time.Now().UTC().Format(time.RFC3339Nano)))
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
