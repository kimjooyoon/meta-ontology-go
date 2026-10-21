package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainrelease"
)

func main() {
	mode := flag.String("mode", "check", "check or write the Gooo-derived release payload module table")
	root := flag.String("root", ".", "repository root")
	flag.Parse()
	if flag.NArg() != 0 || (*mode != "check" && *mode != "write") {
		fail("expected -mode check|write and no positional arguments")
	}
	result, err := toolchainrelease.ProjectPayloadModuleDocument(*root, *mode == "write")
	if err != nil {
		fail(err.Error())
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		fail(err.Error())
	}
	if *mode == "check" && result.Changed {
		fail("release payload module table differs from Gooo source")
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
