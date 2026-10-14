package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const bodyOriginsUsage = "usage: gooo body-origins --activity <name> [--fill-candidates | --fill-candidate <id>] [--inputs-for <package/path:Activity> [--field <stable-field-id>] [--include-control]] <file.gooo>"

func runBodyOrigins(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	flags := map[string]string{"--activity": "", "--inputs-for": "", "--field": "", "--fill-candidate": ""}
	file, control, list := "", false, false
	for i := 0; i < len(args); i++ {
		if args[i] == "--include-control" && !control {
			control = true
		} else if args[i] == "--fill-candidates" && !list {
			list = true
		} else if old, ok := flags[args[i]]; ok && old == "" && i+1 < len(args) && strings.TrimSpace(args[i+1]) != "" && !strings.HasPrefix(args[i+1], "-") {
			flags[args[i]] = args[i+1]
			i++
		} else if file == "" && !strings.HasPrefix(args[i], "-") {
			file = args[i]
		} else {
			fmt.Fprintln(stderr, bodyOriginsUsage)
			return exitUsage
		}
	}
	if file == "" || flags["--activity"] == "" || (control || flags["--field"] != "") && flags["--inputs-for"] == "" {
		fmt.Fprintln(stderr, bodyOriginsUsage)
		return exitUsage
	}
	if list && (flags["--fill-candidate"] != "" || flags["--inputs-for"] != "" || flags["--field"] != "" || control) {
		fmt.Fprintln(stderr, "--fill-candidates lists source assignments; use --fill-candidate to inspect or export one assignment")
		return exitUsage
	}
	if flags["--fill-candidate"] != "" && (control || flags["--field"] != "") {
		fmt.Fprintln(stderr, "--fill-candidate describes declaration/use locations; --field and --include-control require typed value origins")
		return exitUsage
	}
	source, err := reader.ReadFile(file)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	if list {
		return writeBodyFillOriginCandidates(file, source, flags["--activity"], stdout, stderr)
	}
	if flags["--fill-candidate"] != "" {
		return writeBodyFillCandidateOrigins(file, source, flags, stdout, stderr)
	}
	r, err := bodycodegen.InspectTypedValueOrigins(file, source, flags["--activity"])
	if err == nil && flags["--inputs-for"] != "" {
		var raw []byte
		raw, err = bodycodegen.TypedValueOriginInputs(r, flags["--inputs-for"], flags["--field"], control)
		if err == nil {
			_, err = fmt.Fprintln(stdout, string(raw))
		}
	} else if err == nil {
		err = json.NewEncoder(stdout).Encode(r)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	if r.Status != "RESOLVED" {
		return exitFailure
	}
	return exitOK
}
