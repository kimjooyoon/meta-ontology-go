package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const sourceRefactorUsage = "usage: gooo refactor rename-activity (--from <name> | --at-byte <offset>) --to <name> [--json] <file.gooo>"

const sourceRefactorHelp = `Rename a Gooo activity while retaining its semantic identity

` + sourceRefactorUsage + `

Print renamed source to stdout, or a source-bound edit receipt with --json.
--at-byte selects a zero-based source byte within an activity declaration or a
free direct computes call. Escaped quoted bodies and raw bodies retain locations.
The input file is read only. Save source output to a different filename.
Updates local calls, bindings, criteria and observation references. Pins the
previous semantic ID when the declaration derived it from its name.
Comments and intent text retain their wording. Name collisions and unresolved
candidate scopes return an error without partial output.

Scope: one file without imports; fixed computes and declared assembly choices.
External callers, workspace entry names and input keys need separate updates.
The receipt records a source transformation. Execution observations remain bound
to their original source; use fresh checks for the renamed program.
See docs/language/activity-refactoring.md.
`

func runSourceRefactor(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "rename-activity" {
		fmt.Fprintln(stderr, sourceRefactorUsage)
		return exitUsage
	}
	flags := flag.NewFlagSet("rename-activity", flag.ContinueOnError)
	flags.SetOutput(stderr)
	from := flags.String("from", "", "existing activity name")
	atByte := flags.Int("at-byte", -1, "zero-based source byte of declaration or direct call")
	to := flags.String("to", "", "new activity name")
	asJSON := flags.Bool("json", false, "print source-bound edit receipt")
	if err := flags.Parse(args[1:]); err != nil {
		return exitUsage
	}
	atPresent := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "at-byte" {
			atPresent = true
		}
	})
	if (*from != "") == atPresent || *to == "" || flags.NArg() != 1 || (atPresent && *atByte < 0) {
		fmt.Fprintln(stderr, sourceRefactorUsage)
		return exitUsage
	}
	filename := flags.Arg(0)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(stderr, "gooo refactor:", err)
		return exitFailure
	}
	source, err := io.ReadAll(io.LimitReader(file, (128<<10)+1))
	_ = file.Close()
	if err != nil {
		fmt.Fprintln(stderr, "gooo refactor:", err)
		return exitFailure
	}
	if atPresent {
		reference, resolved, resolveErr := bodycodegen.ResolveActivityReference(context.Background(), filename, source, *atByte)
		if resolveErr != nil || !resolved {
			if resolveErr == nil {
				resolveErr = fmt.Errorf("source position does not resolve to a Gooo activity")
			}
			fmt.Fprintln(stderr, "gooo refactor:", resolveErr)
			return exitFailure
		}
		*from = reference.Name
	}
	result, err := bodycodegen.RenameActivitySource(context.Background(), filename, source, *from, *to)
	if err != nil {
		fmt.Fprintln(stderr, "gooo refactor:", err)
		return exitFailure
	}
	if *asJSON {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(result)
	} else {
		_, err = io.WriteString(stdout, result.Source)
	}
	if err != nil {
		fmt.Fprintln(stderr, "gooo refactor:", err)
		return exitFailure
	}
	return exitOK
}
