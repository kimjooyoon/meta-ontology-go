package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/viewlens"
)

func runViewLens(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || (args[0] != "load" && args[0] != "record" && args[0] != "read") {
		fmt.Fprintln(stderr, "usage: gooo view-lens load|record|read [--actor id] [--audience agent|person] "+
			"[--state-dir directory] [--event json] [--kind event|profile] [--digest sha256:...]")
		return exitUsage
	}
	flags := flag.NewFlagSet("view-lens", flag.ContinueOnError)
	flags.SetOutput(stderr)
	actor := flags.String("actor", "agent", "profile identity for load")
	audience := flags.String("audience", "agent", "agent or person for load")
	directory := flags.String("state-dir", "", "detached local profile directory")
	event := flags.String("event", "", "source-bound record event JSON")
	kind := flags.String("kind", "event", "archive kind: event or profile")
	digest := flags.String("digest", "", "raw archive digest")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 {
		return exitUsage
	}
	if *directory == "" {
		if root, err := os.UserConfigDir(); err == nil {
			*directory = filepath.Join(root, "gooo", "view-lenses")
		}
	}
	var result any
	if args[0] == "load" {
		result = viewlens.Load(*directory, *actor, *audience)
	} else if args[0] == "read" {
		result = viewlens.ReadArchive(*directory, *kind, *digest)
	} else {
		result = viewlens.Record(*directory, []byte(*event))
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	// Storage refusal is advisory and is described by state/issue/persisted.
	return exitOK
}
